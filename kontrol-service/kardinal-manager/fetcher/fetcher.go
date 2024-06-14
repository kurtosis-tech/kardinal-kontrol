package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	yamlV3 "gopkg.in/yaml.v3"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"net/http"
	"sigs.k8s.io/yaml"
)

func FetchConfig() ([]byte, error) {

	//TODO get this from the argument and set the value in the deployment yaml file with an ENV VAR
	configEndpoint := "https://gist.githubusercontent.com/leoporoli/d9afda02795f18abef04fa74afe3b555/raw/a231255e66585dd295dd1e83318245fd725b30dd/deployment-example.yml"

	resp, err := http.Get(configEndpoint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching configuration from endpoint '%s'", configEndpoint)
	}
	defer resp.Body.Close()

	responseBodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error reading YAML data: %v", err)
	}

	var jsonObject []interface{}

	err = json.Unmarshal(responseBodyContent, &jsonObject)
	if err != nil {
		return nil, stacktrace.Propagate(err, "And error occurred unmarshalling", err)
	}

	var concatenatedYamlContent []byte
	for _, jsonData := range jsonObject {
		jsonDataMap := jsonData.(map[string]interface{})
		jsonByte, err := json.Marshal(jsonDataMap)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error ", err)
		}
		yamlData, err := yaml.JSONToYAML(jsonByte)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting JSON to YAML: %v", err)
		}
		concatenatedYamlContent = append(concatenatedYamlContent, []byte("---\n")...)
		concatenatedYamlContent = append(concatenatedYamlContent, yamlData...)
	}

	return concatenatedYamlContent, nil
}

func ApplyConfig(dynaMicClient *dynamic.DynamicClient, yamlFileContent []byte) error {

	yamlReader := bytes.NewReader(yamlFileContent)

	// Read YAML file
	yamlFile, err := io.ReadAll(yamlReader)
	if err != nil {
		return stacktrace.Propagate(err, "And error occurred reading yaml content")
	}

	// Decode YAML to Unstructured object
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlFile, &obj)
	if err != nil {
		return stacktrace.Propagate(err, "And error occurred unmarshalling the yaml content")
	}

	// Apply the configuration

	groupVersionResource := obj.GroupVersionKind().GroupVersion().WithResource(obj.GetKind() + "s")

	namespaceableResource := dynaMicClient.Resource(groupVersionResource)

	resource := namespaceableResource.Namespace(obj.GetNamespace())

	_, err = resource.Create(context.Background(), &obj, metav1.CreateOptions{})
	//applyOpts := metav1.ApplyOptions{FieldManager: "kube-apply"}
	//_, err = resource.Apply(context.Background(), obj.GetName(), &obj, applyOpts)
	if err != nil {
		return stacktrace.Propagate(err, "And error occurred applying the configuration")
	}

	fmt.Println("Configuration applied successfully!")

	return nil
}

func ApplyConfig2(dynaMicClient *dynamic.DynamicClient, discoveryMapper *restmapper.DeferredDiscoveryRESTMapper, yamlFileContent []byte) error {
	yamlReader := bytes.NewReader(yamlFileContent)

	dec := yamlV3.NewDecoder(yamlReader)

	for {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
		err := dec.Decode(obj.Object)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred applying the configuration")
		}
		if obj.Object == nil {
			return stacktrace.Propagate(err, "Object is nil")
		}

		// get GroupVersionResource to invoke the dynamic client
		gvk := obj.GroupVersionKind()
		restMapping, err := discoveryMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}
		gvr := restMapping.Resource
		// apply the YAML doc
		namespace := obj.GetNamespace()
		if len(namespace) == 0 {
			namespace = "default"
		}
		applyOpts := metav1.ApplyOptions{FieldManager: "kube-apply"}
		_, err = dynaMicClient.Resource(gvr).Namespace(namespace).Apply(context.TODO(), obj.GetName(), obj, applyOpts)
		if err != nil {
			return fmt.Errorf("apply error: %w", err)
		}
	}

	return nil
}
