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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kardinal.kontrol/kardinal-manager/kubernetes_client"
	"net/http"
	"sigs.k8s.io/yaml"
	"time"
)

const (
	defaultNamespace = "default"
	tickerDuration   = time.Second * 5
)

var (
	yamlDelimiter = []byte("---\n")
)

type fetcher struct {
	kubernetesClient *kubernetes_client.KubernetesClient
	configEndpoint   string
}

func NewFetcher(kubernetesClient *kubernetes_client.KubernetesClient, configEndpoint string) *fetcher {
	return &fetcher{kubernetesClient: kubernetesClient, configEndpoint: configEndpoint}
}

func (fetcher *fetcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(tickerDuration)
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
	yamlFileContent, err := fetcher.fetchConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching config from '%s'", fetcher.configEndpoint)
	}

	if err := fetcher.applyConfig(ctx, yamlFileContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred applying the config in the cluster!")
	}

	return nil
}

func (fetcher *fetcher) fetchConfig() ([]byte, error) {
	resp, err := http.Get(fetcher.configEndpoint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching configuration from endpoint '%s'", fetcher.configEndpoint)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error reading the response from '%v'", fetcher.configEndpoint)
	}

	var jsonListObject []interface{}

	if err = json.Unmarshal(responseBodyBytes, &jsonListObject); err != nil {
		return nil, stacktrace.Propagate(err, "And error occurred unmarshalling the response to a JSON list", err)
	}

	var concatenatedYamlContent []byte
	for _, jsonData := range jsonListObject {
		jsonDataMap, ok := jsonData.(map[string]interface{})
		if !ok {
			return nil, stacktrace.NewError("An error occurred while casting the JSON data to a map[string]interface{}")
		}
		jsonByte, marshallErr := json.Marshal(jsonDataMap)
		if marshallErr != nil {
			return nil, stacktrace.Propagate(err, "An error occurred marshalling the JSON data map", err)
		}
		yamlData, toYAMLErr := yaml.JSONToYAML(jsonByte)
		if toYAMLErr != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting the JSON content to YAML")
		}
		concatenatedYamlContent = append(concatenatedYamlContent, yamlDelimiter...)
		concatenatedYamlContent = append(concatenatedYamlContent, yamlData...)
	}

	return concatenatedYamlContent, nil
}

func (fetcher *fetcher) applyConfig(ctx context.Context, yamlFileContent []byte) error {
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
		namespace := unstructuredObject.GetNamespace()
		if len(namespace) == 0 {
			namespace = defaultNamespace
		}

		applyOpts := metav1.ApplyOptions{FieldManager: "kube-apply"}
		namespaceResource := fetcher.kubernetesClient.GetDynamicClient().Resource(groupVersionResource).Namespace(namespace)

		_, err = namespaceResource.Apply(ctx, unstructuredObject.GetName(), unstructuredObject, applyOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred applying the k8s resource with name '%s' in namespace '%s'", unstructuredObject.GetName(), unstructuredObject.GetNamespace())
		}
	}

}
