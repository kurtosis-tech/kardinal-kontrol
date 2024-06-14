package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"kardinal.kontrol/kardinal-manager/fetcher"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	versionedclient "istio.io/client-go/pkg/clientset/versioned"
)

const (
	successExitCode = 0
)

func main() {

	if err := basicInteractionWithK8sAndIstio(); err != nil {
		logrus.Fatalf("An error occurred while calling basicInteractionWithK8sAndIstio()!\nError was: %s", err)
	}

	// No clients connection so-far
	//if err := server.CreateAndStartRestAPIServer(); err != nil {
	//	logrus.Fatalf("The REST API server is down, exiting!\nError was: %s", err)
	//}

	os.Exit(successExitCode)
}

func basicInteractionWithK8sAndIstio() error {
	var config *rest.Config
	var err error

	// Load in-cluster configuration
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to out-of-cluster configuration (for local development)
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	discoveryClient := memory.NewMemCacheClient(clientset.Discovery())
	discoveryMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)

	yamlFileContent, err := fetcher.FetchConfig()
	if err != nil {
		panic(err.Error())
	}

	if err := fetcher.ApplyConfig2(dynamicClient, discoveryMapper, yamlFileContent); err != nil {
		panic(err.Error())
	}

	// Create context
	ctx := context.Background()

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
	ic, err := versionedclient.NewForConfig(config)
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
