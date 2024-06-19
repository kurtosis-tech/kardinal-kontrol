package fetcher

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"kardinal.kontrol/kardinal-manager/cluster_manager"
	"kardinal.kontrol/kardinal-manager/utils"
	"net/http"
	"sigs.k8s.io/yaml"
	"time"
)

const (
	defaultTickerDuration              = time.Second * 5
	fetcherJobDurationSecondsEnvVarKey = "FETCHER_JOB_DURATION_SECONDS"
)

var (
	yamlDelimiter = []byte("---\n")
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
	clusterManager *cluster_manager.ClusterManager
	configEndpoint string
}

func NewFetcher(clusterManager *cluster_manager.ClusterManager, configEndpoint string) *fetcher {
	return &fetcher{clusterManager: clusterManager, configEndpoint: configEndpoint}
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
		if err = fetcher.clusterManager.CreateNamespaceIfNotExists(ctx, configResponseObj.Namespace); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating namespace '%s'", namespace)
		}
	}

	if err = fetcher.clusterManager.ApplyYamlFileContentInNamespace(ctx, namespace, yamlFileContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred applying the config in the cluster!")
	}

	if configResponseObj.PruneLabels != nil {
		if err := fetcher.clusterManager.RemoveNamespaceResourcesByLabels(ctx, namespace, *labels); err != nil {
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
