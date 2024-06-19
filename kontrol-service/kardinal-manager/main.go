package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kardinal.kontrol/kardinal-manager/cluster_manager"
	"kardinal.kontrol/kardinal-manager/fetcher"
	"kardinal.kontrol/kardinal-manager/logger"
	"kardinal.kontrol/kardinal-manager/utils"
	"os"
)

const (
	successExitCode                = 0
	clusterConfigEndpointEnvVarKey = "CLUSTER_CONFIG_ENDPOINT"
)

func main() {

	// Create context
	ctx := context.Background()

	if err := logger.ConfigureLogger(); err != nil {
		logrus.Fatal("An error occurred configuring the logger!\nError was: %s", err)
	}

	configEndpoint, err := utils.GetFromEnvVar(clusterConfigEndpointEnvVarKey, "the config endpoint")
	if err != nil {
		logrus.Fatal("An error occurred getting the config endpoint from the env vars!\nError was: %s", err)
	}

	clusterManager, err := cluster_manager.CreateClusterManager()
	if err != nil {
		logrus.Fatal("An error occurred while creating the cluster manager!\nError was: %s", err)
	}

	fetcher := fetcher.NewFetcher(clusterManager, configEndpoint)

	if err = fetcher.Run(ctx); err != nil {
		logrus.Fatalf("An error occurred while running the fetcher!\nError was: %s", err)
	}

	// Uncomment if you want to test basic interaction with K8s cluster and Istio resources
	//if err := basicInteractionWithK8sAndIstio(ctx, clusterManager); err != nil {
	//	logrus.Fatalf("An error occurred while calling basicInteractionWithK8sAndIstio()!\nError was: %s", err)
	//}

	// No external clients connection so-far
	//if err := server.CreateAndStartRestAPIServer(); err != nil {
	//	logrus.Fatalf("The REST API server is down, exiting!\nError was: %s", err)
	//}

	os.Exit(successExitCode)
}

func basicInteractionWithK8sAndIstio(ctx context.Context, clusterManager *cluster_manager.ClusterManager) error {

	podLabels := map[string]string{}

	pods, err := clusterManager.GetPodsByLabels(ctx, "", podLabels)

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("Pod Name: %s\n", pod.Name)
	}

	// Istio Client
	ic, err := versionedclient.NewForConfig(clusterManager.GetKubernetesClientConfig())
	if err != nil {
		logrus.Fatalf("Failed to create istio client: %s", err)
	}

	// Test VirtualServices
	vsList, err := ic.NetworkingV1alpha3().VirtualServices("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to get VirtualService in %s namespace: %s", "default", err)
	}

	for i := range vsList.Items {
		vs := vsList.Items[i]
		logrus.Printf("Index: %d VirtualService Hosts: %+v\n", i, vs.Spec.GetHosts())
	}

	// Test DestinationRules
	drList, err := ic.NetworkingV1alpha3().DestinationRules("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to get DestinationRule in %s namespace: %s", "", err)
	}

	for i := range drList.Items {
		dr := drList.Items[i]
		logrus.Printf("Index: %d DestinationRule Host: %+v\n", i, dr.Spec.GetHost())
	}

	return nil
}
