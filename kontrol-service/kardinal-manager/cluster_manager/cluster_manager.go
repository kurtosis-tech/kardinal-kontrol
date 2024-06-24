package cluster_manager

import (
	"bytes"
	"context"
	"errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	yamlV3 "gopkg.in/yaml.v3"
	"io"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	apps "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"kardinal.kontrol/kardinal-manager/cluster_manager/istio_manager"
	"kardinal.kontrol/kardinal-manager/types"
	"time"
)

const (
	listOptionsTimeoutSeconds       int64 = 10
	fieldManager                          = "kardinal-manager"
	deleteOptionsGracePeriodSeconds int64 = 0
	istioLabel                            = "istio-injection"
)

var (
	globalCreateOptions = metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun: nil,
		// We need every object to have this field manager so that the Kurtosis objects can all seamlessly modify Kubernetes resources
		FieldManager:    fieldManager,
		FieldValidation: "",
	}
)

type ClusterManager struct {
	kubernetesClient *kubernetesClient
	istioManager     *istio_manager.IstioManager
}

func NewClusterManager(kubernetesClient *kubernetesClient, istioManager *istio_manager.IstioManager) *ClusterManager {
	return &ClusterManager{kubernetesClient: kubernetesClient, istioManager: istioManager}
}

// TODO remove this method once IstioManager (or IstioClient) has been migrated inside the cluster manager
func (manager *ClusterManager) GetKubernetesClientConfig() *rest.Config {
	return manager.kubernetesClient.GetConfig()
}

func (manager *ClusterManager) GetPodsByLabels(ctx context.Context, namespace string, podLabels map[string]string) (*apiv1.PodList, error) {
	namespacePodClient := manager.kubernetesClient.GetClientSet().CoreV1().Pods(namespace)

	opts := buildListOptionsFromLabels(podLabels)
	pods, err := namespacePodClient.List(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get pods with labels '%+v', instead a non-nil error was returned", podLabels)
	}

	// Only return objects not tombstoned by Kubernetes
	var podsNotMarkedForDeletionList []apiv1.Pod
	for _, pod := range pods.Items {
		deletionTimestamp := pod.GetObjectMeta().GetDeletionTimestamp()
		if deletionTimestamp == nil {
			podsNotMarkedForDeletionList = append(podsNotMarkedForDeletionList, pod)
		}
	}
	podsNotMarkedForDeletionPodList := apiv1.PodList{
		Items:    podsNotMarkedForDeletionList,
		TypeMeta: pods.TypeMeta,
		ListMeta: pods.ListMeta,
	}
	return &podsNotMarkedForDeletionPodList, nil
}

func (manager *ClusterManager) CreateNamespaceIfNotExists(ctx context.Context, namespace string) error {
	_, err := manager.GetNamespace(ctx, namespace)
	if err != nil {
		//TODO improve err check, create only when statusError.ErrStatus.Coded == 404 or Reason == "NotFound"
		namespaceLabels := map[string]string{
			"istio-injection": "enabled",
		}

		namespaceAnnotations := map[string]string{}

		_, err = manager.CreateNamespace(ctx, namespace, namespaceLabels, namespaceAnnotations)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating namespace '%s' with labels '%+v'", namespace, namespaceLabels)
		}
	}

	return nil
}

func (manager *ClusterManager) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {

	namespace, err := manager.kubernetesClient.GetClientSet().CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name '%s'", name)
	}
	deletionTimestamp := namespace.GetObjectMeta().GetDeletionTimestamp()
	if deletionTimestamp != nil {
		return nil, stacktrace.Propagate(err, "Namespace with name '%s' has been marked for deletion", namespace)
	}
	return namespace, nil
}

func (manager *ClusterManager) CreateNamespace(
	ctx context.Context,
	name string,
	namespaceLabels map[string]string,
	namespaceAnnotations map[string]string,
) (*apiv1.Namespace, error) {

	namespace := &apiv1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     namespaceLabels,
			Annotations:                namespaceAnnotations,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec: apiv1.NamespaceSpec{
			Finalizers: nil,
		},
		Status: apiv1.NamespaceStatus{
			Phase:      "",
			Conditions: nil,
		},
	}

	namespaceResult, err := manager.kubernetesClient.GetClientSet().CoreV1().Namespaces().Create(ctx, namespace, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

func (manager *ClusterManager) ApplyYamlFileContentInNamespace(ctx context.Context, namespace string, yamlFileContent []byte) error {
	yamlReader := bytes.NewReader(yamlFileContent)

	dec := yamlV3.NewDecoder(yamlReader)

	for {
		unstructuredObject := &unstructured.Unstructured{Object: map[string]interface{}{}}
		err := dec.Decode(unstructuredObject.Object)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred decoding the unstructured object")
		}
		if unstructuredObject.Object == nil {
			return stacktrace.NewError("Expected to find the object value after decoding the unstructured object but it was not found")
		}

		groupVersionKind := unstructuredObject.GroupVersionKind()
		restMapping, err := manager.kubernetesClient.GetDiscoveryMapper().RESTMapping(groupVersionKind.GroupKind(), groupVersionKind.Version)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the rest mapping for GVK")
		}

		groupVersionResource := restMapping.Resource

		if namespace != unstructuredObject.GetNamespace() {
			return stacktrace.NewError(
				"The namespace '%s' in resource '%s' kind '%s' is different from the main namespace '%s'",
				unstructuredObject.GetNamespace(),
				unstructuredObject.GetName(),
				unstructuredObject.GetKind(),
				namespace,
			)
		}

		applyOpts := metav1.ApplyOptions{FieldManager: fieldManager}
		namespaceResource := manager.kubernetesClient.GetDynamicClient().Resource(groupVersionResource).Namespace(namespace)

		_, err = namespaceResource.Apply(ctx, unstructuredObject.GetName(), unstructuredObject, applyOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred applying the k8s resource with name '%s' in namespace '%s'", unstructuredObject.GetName(), unstructuredObject.GetNamespace())
		}
	}
}

func (manager *ClusterManager) RemoveNamespaceResourcesByLabels(ctx context.Context, namespace string, labels map[string]string) error {
	clientset := manager.kubernetesClient.GetClientSet()

	opts := buildListOptionsFromLabels(labels)

	deleteOptions := metav1.NewDeleteOptions(deleteOptionsGracePeriodSeconds)

	// Delete deployments
	if err := clientset.AppsV1().Deployments(namespace).DeleteCollection(ctx, *deleteOptions, opts); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing deployments in namespace '%s'", namespace)
	}

	// Delete services one by one because there is not DeleteCollection function for services
	servicesToRemove, err := clientset.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred listing services")
	}

	for _, service := range servicesToRemove.Items {
		if err := clientset.CoreV1().Services(namespace).Delete(ctx, service.GetName(), *deleteOptions); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing service '%s' from namespace '%s'", service.GetName(), namespace)
		}
	}

	// Istio Client
	//TODO Use the internal manager.Istio field instead of creating the client again
	ic, err := versionedclient.NewForConfig(manager.kubernetesClient.GetConfig())
	if err != nil {
		logrus.Fatalf("Failed to create istio client: %s", err)
	}

	// Delete VirtualServices
	if err = ic.NetworkingV1alpha3().VirtualServices(namespace).DeleteCollection(ctx, *deleteOptions, opts); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing virtual services in namespace '%s'", namespace)
	}

	// Delete DestinationRules
	if err = ic.NetworkingV1alpha3().DestinationRules(namespace).DeleteCollection(ctx, *deleteOptions, opts); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing destination rules in namespace '%s'", namespace)
	}

	return nil
}

func buildListOptionsFromLabels(labelsMap map[string]string) metav1.ListOptions {
	return metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        labels.SelectorFromSet(labelsMap).String(),
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       int64Ptr(listOptionsTimeoutSeconds),
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	}
}

func int64Ptr(i int64) *int64 { return &i }

func (manager *ClusterManager) createOrUpdateService(ctx context.Context, service apiv1.Service) error {
	serviceClient := manager.kubernetesClient.GetClientSet().CoreV1().Services(service.Namespace)
	existingService, err := serviceClient.Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		// Resource does not exist, create new one
		_, err = serviceClient.Create(ctx, &service, metav1.CreateOptions{})
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create service: %s", service.GetName())
		}
	} else {
		// Update the resource version to the latest before updating
		service.ResourceVersion = existingService.ResourceVersion
		_, err = serviceClient.Update(ctx, &service, metav1.UpdateOptions{})
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update service: %s", service.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) ApplyClusterResources(ctx context.Context, clusterResources *types.ClusterResources) {

	allNSs := [][]string{
		lo.Uniq(lo.Map(clusterResources.Services, func(item apiv1.Service, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.Deployments, func(item apps.Deployment, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.VirtualServices, func(item istioclient.VirtualService, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.DestinationRules, func(item istioclient.DestinationRule, _ int) string { return item.Namespace })),
		{clusterResources.Gateway.Namespace},
	}

	uniqueNamespaces := lo.Uniq(lo.Flatten(allNSs))
	lo.ForEach(uniqueNamespaces, func(namespace string, _ int) { manager.EnsureNamespace(ctx, namespace) })

	lo.ForEach(clusterResources.Services, func(service apiv1.Service, _ int) {
		if err := manager.createOrUpdateService(ctx, service); err != nil {
		}
	})

	lo.ForEach(clusterResources.Deployments, func(deployment apps.Deployment, _ int) {
		deploymentClient := manager.kubernetesClient.GetClientSet().AppsV1().Deployments(deployment.Namespace)
		existingDeployment, err := deploymentClient.Get(ctx, deployment.Name, metav1.GetOptions{})
		if err != nil {
			_, err = deploymentClient.Create(ctx, &deployment, metav1.CreateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to create deployment: %s", err)
			}
		} else {
			deployment.ResourceVersion = existingDeployment.ResourceVersion
			_, err = deploymentClient.Update(ctx, &deployment, metav1.UpdateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to update deployment: %s", err)
			}
		}
	})

	ic, err := versionedclient.NewForConfig(manager.kubernetesClient.GetConfig())
	if err != nil {
		logrus.Fatalf("Failed to create istio client: %s", err)
	}

	lo.ForEach(clusterResources.VirtualServices, func(virtService istioclient.VirtualService, _ int) {
		virtServicesClient := ic.NetworkingV1alpha3().VirtualServices(virtService.Namespace)
		existingVirtService, err := virtServicesClient.Get(ctx, virtService.Name, metav1.GetOptions{})
		if err != nil {
			_, err = virtServicesClient.Create(ctx, &virtService, metav1.CreateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to create virtual service: %s", err)
			}
		} else {
			virtService.ResourceVersion = existingVirtService.ResourceVersion
			_, err = virtServicesClient.Update(ctx, &virtService, metav1.UpdateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to update virtual service: %s", err)
			}
		}
	})

	lo.ForEach(clusterResources.DestinationRules, func(destRule istioclient.DestinationRule, _ int) {
		destRulesClient := ic.NetworkingV1alpha3().DestinationRules(destRule.Namespace)
		existingDestRule, err := destRulesClient.Get(ctx, destRule.Name, metav1.GetOptions{})
		if err != nil {
			_, err = destRulesClient.Create(ctx, &destRule, metav1.CreateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to create destination rule: %s", err)
			}
		} else {
			destRule.ResourceVersion = existingDestRule.ResourceVersion
			_, err = destRulesClient.Update(ctx, &destRule, metav1.UpdateOptions{})
			if err != nil {
				logrus.Fatalf("Failed to update destination rule: %s", err)
			}
		}
	})

	gateway := &clusterResources.Gateway
	existingGateway, err := ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Get(ctx, gateway.Name, metav1.GetOptions{})
	if err != nil {
		_, err = ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Create(ctx, gateway, metav1.CreateOptions{})
		if err != nil {
			logrus.Fatalf("Failed to create gateway: %s", err)
		}
	} else {
		gateway.ResourceVersion = existingGateway.ResourceVersion
		_, err = ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Update(ctx, gateway, metav1.UpdateOptions{})
		if err != nil {
			logrus.Fatalf("Failed to update gateway: %s", err)
		}
	}
}

func (manager *ClusterManager) EnsureNamespace(ctx context.Context, namespace string) {

	existingNamespace, err := manager.kubernetesClient.GetClientSet().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil && existingNamespace != nil {
		value, found := existingNamespace.Labels[istioLabel]
		if !found || value != "enabled" {
			existingNamespace.Labels[istioLabel] = "enabled"
			manager.kubernetesClient.GetClientSet().CoreV1().Namespaces().Update(ctx, existingNamespace, metav1.UpdateOptions{})
		}
	} else {
		newNamespace := apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					istioLabel: "enabled",
				},
			},
		}
		_, err = manager.kubernetesClient.GetClientSet().CoreV1().Namespaces().Create(ctx, &newNamespace, metav1.CreateOptions{})
		if err != nil {
			logrus.Panicf("Failed to create Namespace: %v", err)
			panic(err.Error())
		}
	}
}
