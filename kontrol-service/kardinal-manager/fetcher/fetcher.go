package fetcher

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"kardinal.kontrol/kardinal-manager/cluster_manager"
	"kardinal.kontrol/kardinal-manager/types"
	"kardinal.kontrol/kardinal-manager/utils"
	"net/http"
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
	clusterResources, err := fetcher.getClusterResourcesFromCloud()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching cluster resources from cloud")
	}

	logrus.Debugf("Cluster resources %+v", clusterResources)

	//TODO handle error
	fetcher.clusterManager.ApplyClusterResources(ctx, clusterResources)

	return nil
}

func (fetcher *fetcher) getClusterResourcesFromCloud() (*types.ClusterResources, error) {
	resp, err := http.Get(fetcher.configEndpoint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error fetching cluster resources from endpoint '%s'", fetcher.configEndpoint)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error reading the response from '%v'", fetcher.configEndpoint)
	}

	var clusterResources *types.ClusterResources

	if err = json.Unmarshal(responseBodyBytes, &clusterResources); err != nil {
		return nil, stacktrace.Propagate(err, "And error occurred unmarshalling the response to a config response object")
	}

	return clusterResources, nil
}
