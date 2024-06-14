package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kardinal.kontrol/kardinal-manager/fetcher"
	"kardinal.kontrol/kardinal-manager/kubernetes_client"
	"os"
)

const (
	successExitCode = 0
)

func main() {

	// Create context
	ctx := context.Background()

	kubernetesClient, err := kubernetes_client.CreateKubernetesClient()
	if err != nil {
		logrus.Fatal("An error occurred while creating the Kubernetes client!\nError was: %s", err)
	}

	//TODO get this from  the deployment yaml file with an ENV VAR
	configEndpoint := "https://gist.githubusercontent.com/leoporoli/d9afda02795f18abef04fa74afe3b555/raw/a231255e66585dd295dd1e83318245fd725b30dd/deployment-example.yml"

	yamlFileContent, err := fetcher.FetchConfig(configEndpoint)
	if err != nil {
		logrus.Fatalf("An error occurred fetching config from!\nError was: %s", err)
	}

	if err := fetcher.ApplyConfig(ctx, kubernetesClient, yamlFileContent); err != nil {
		logrus.Fatalf("An error occurred applying the config in the cluster!\nError was: %s", err)
	}

	// Uncomment if you want to test basic interaction with K8s cluster and Istio resources
	//if err := basicInteractionWithK8sAndIstio(kubernetesClient); err != nil {
	//	logrus.Fatalf("An error occurred while calling basicInteractionWithK8sAndIstio()!\nError was: %s", err)
	//}

	// No external clients connection so-far
	//if err := server.CreateAndStartRestAPIServer(); err != nil {
	//	logrus.Fatalf("The REST API server is down, exiting!\nError was: %s", err)
	//}

	os.Exit(successExitCode)
}

func basicInteractionWithK8sAndIstio(ctx context.Context, kubernetesClient *kubernetes_client.KubernetesClient) error {

	clientset := kubernetesClient.GetClientSet()

	// List pods
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("Pod Name: %s\n", pod.Name)
	}

	// Istio Client
	ic, err := versionedclient.NewForConfig(kubernetesClient.GetConfig())
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
