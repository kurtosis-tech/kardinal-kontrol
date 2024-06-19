package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
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
	"kardinal.kontrol/kardinal-manager/kubernetes_client"
	"kardinal.kontrol/kardinal-manager/utils"
	"net/http"
	"sigs.k8s.io/yaml"
	"time"
)

const (
	defaultTickerDuration                    = time.Second * 5
	listOptionsTimeoutSeconds          int64 = 10
	deleteOptionsGracePeriodSeconds    int64 = 0
	fieldManager                             = "kardinal-manager"
	fetcherJobDurationSecondsEnvVarKey       = "FETCHER_JOB_DURATION_SECONDS"
)

var (
	yamlDelimiter       = []byte("---\n")
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

type configResponse struct {
	// the main namespace where the workflow will be applied
	Namespace string `json:"namespace,omitempty"`
	// Create or not the namespace
	CreateNamespace bool `json:"create_namespace,omitempty"`
	// these labels identify resources to remove from the cluster
	PruneLabels *map[string]string `json:"prune_labels,omitempty"`
	// the Kubernetes resources to apply in the cluster
	KubernetesResources []interface{} `json:"kubernetes_resources"`
}

type fetcher struct {
	kubernetesClient *kubernetes_client.KubernetesClient
	configEndpoint   string
}

func NewFetcher(kubernetesClient *kubernetes_client.KubernetesClient, configEndpoint string) *fetcher {
	return &fetcher{kubernetesClient: kubernetesClient, configEndpoint: configEndpoint}
}

func (fetcher *fetcher) Run(ctx context.Context) error {

	fetcherTickerDuration := defaultTickerDuration

	fetcherJobDurationSecondsEnVarValue, err := utils.GetIntFromEnvVar(fetcherJobDurationSecondsEnvVarKey, "fetcher job duration seconds")
	if err != nil {
		logrus.Debugf("an error occurred while getting the fetcher job durations seconds from the env var, using default value '%s'. Error:\n%s", defaultTickerDuration, err)
	}

	if fetcherJobDurationSecondsEnVarValue != 0 {
		fetcherTickerDuration = time.Second * time.Duration(int64(fetcherJobDurationSecondsEnVarValue))
	}

	ticker := time.NewTicker(fetcherTickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logrus.Debugf("New fetcher execution at %s", time.Now())
			if err := fetcher.fetchAndApply(ctx); err != nil {
				return stacktrace.Propagate(err, "Failed to fetch and apply the cluster configuration")
			}
		}
	}
}

func (fetcher *fetcher) fetchAndApply(ctx context.Context) error {
	configResponseObj, err := fetcher.getConfigResponseFromEndpoint()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching config from endpoint")
	}

	kubernetesResources := configResponseObj.KubernetesResources
	namespace := configResponseObj.Namespace
	labels := configResponseObj.PruneLabels

	yamlFileContent, err := fetcher.getYamlContentFromKubernetesResources(kubernetesResources)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the YAML content from kubernetes resources")
	}

	if configResponseObj.CreateNamespace && configResponseObj.Namespace != "" {
		if err = fetcher.CreateNamespaceIfNotExists(ctx, configResponseObj.Namespace); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating namespace '%s'", namespace)
		}
	}

	if err = fetcher.applyConfig(ctx, namespace, yamlFileContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred applying the config in the cluster!")
	}

	if configResponseObj.PruneLabels != nil {
		if err := fetcher.removeNamespaceResourcesByLabels(ctx, namespace, *labels); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the namespace resources in '%s' with labels '%+v'", namespace, labels)
		}
	}

	return nil
}

func (fetcher *fetcher) getConfigResponseFromEndpoint() (*configResponse, error) {
	resp, err := http.Get(fetcher.configEndpoint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching configuration from endpoint '%s'", fetcher.configEndpoint)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error reading the response from '%v'", fetcher.configEndpoint)
	}

	var configResponseObj *configResponse

	if err = json.Unmarshal(responseBodyBytes, &configResponseObj); err != nil {
		return nil, stacktrace.Propagate(err, "And error occurred unmarshalling the response to a config response object")
	}

	if configResponseObj.Namespace == "" {
		return nil, stacktrace.Propagate(err, "An error occurred fetching configuration from endpoint, namespace is empty")
	}

	return configResponseObj, nil
}

func (fetcher *fetcher) getYamlContentFromKubernetesResources(kubernetesResources []interface{}) ([]byte, error) {

	var concatenatedYamlContent []byte
	for _, jsonData := range kubernetesResources {
		jsonDataMap, ok := jsonData.(map[string]interface{})
		if !ok {
			return nil, stacktrace.NewError("An error occurred while casting the JSON data to a map[string]interface{}")
		}
		jsonByte, err := json.Marshal(jsonDataMap)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred marshalling the JSON data map")
		}
		yamlData, err := yaml.JSONToYAML(jsonByte)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting the JSON content to YAML")
		}
		concatenatedYamlContent = append(concatenatedYamlContent, yamlDelimiter...)
		concatenatedYamlContent = append(concatenatedYamlContent, yamlData...)
	}

	return concatenatedYamlContent, nil
}

func (fetcher *fetcher) applyConfig(ctx context.Context, namespace string, yamlFileContent []byte) error {
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
		restMapping, err := fetcher.kubernetesClient.GetDiscoveryMapper().RESTMapping(groupVersionKind.GroupKind(), groupVersionKind.Version)
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
		namespaceResource := fetcher.kubernetesClient.GetDynamicClient().Resource(groupVersionResource).Namespace(namespace)

		_, err = namespaceResource.Apply(ctx, unstructuredObject.GetName(), unstructuredObject, applyOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred applying the k8s resource with name '%s' in namespace '%s'", unstructuredObject.GetName(), unstructuredObject.GetNamespace())
		}
	}
}

func (fetcher *fetcher) removeNamespaceResourcesByLabels(ctx context.Context, namespace string, labels map[string]string) error {
	clientset := fetcher.kubernetesClient.GetClientSet()

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
	//TODO move inside the Kubernetes client and we could probably create a Cluster Manager to wrap both clients
	ic, err := versionedclient.NewForConfig(fetcher.kubernetesClient.GetConfig())
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

// TODO move into the Cluster Manager
func (fetcher *fetcher) CreateNamespaceIfNotExists(ctx context.Context, namespace string) error {
	_, err := fetcher.GetNamespace(ctx, namespace)
	if err != nil {
		//TODO improve err check, create only when statusError.ErrStatus.Coded == 404 or Reason == "NotFound"
		namespaceLabels := map[string]string{
			"istio-injection": "enabled",
		}

		namespaceAnnotations := map[string]string{}

		_, err = fetcher.CreateNamespace(ctx, namespace, namespaceLabels, namespaceAnnotations)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating namespace '%s' with labels '%+v'", namespace, namespaceLabels)
		}
	}

	return nil
}

// TODO move into the Cluster Manager
func (fetcher *fetcher) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {

	namespace, err := fetcher.kubernetesClient.GetClientSet().CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{
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

// TODO move into the Cluster Manager
func (fetcher *fetcher) CreateNamespace(
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

	namespaceResult, err := fetcher.kubernetesClient.GetClientSet().CoreV1().Namespaces().Create(ctx, namespace, globalCreateOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

// TODO move into the Cluster Manager
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
