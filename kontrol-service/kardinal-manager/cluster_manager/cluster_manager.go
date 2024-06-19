package cluster_manager

import (
	"bytes"
	"context"
	"errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	yamlV3 "gopkg.in/yaml.v3"
	"io"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"kardinal.kontrol/kardinal-manager/cluster_manager/istio_manager"
	"time"
)

const (
	listOptionsTimeoutSeconds       int64 = 10
	fieldManager                          = "kardinal-manager"
	deleteOptionsGracePeriodSeconds int64 = 0
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
