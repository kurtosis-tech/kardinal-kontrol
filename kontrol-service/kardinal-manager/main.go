package main

import (
	"context"
	"fmt"
	istio "istio.io/api/networking/v1alpha3"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

func main() {
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

	// create ist io client
	ic, err := istioclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	namespace := "default"
	vsClient := ic.NetworkingV1alpha3().VirtualServices(namespace)
	drClient := ic.NetworkingV1alpha3().DestinationRules(namespace)

	// Test VirtualServices
	vsList, err := vsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to retrieve VirtualServices in namespace: %s with error:\n%s", namespace, err)
	}
	var reviewsVirtualService *istionetworking.VirtualService
	for i := range vsList.Items {
		vs := vsList.Items[i]
		if vs.Name == "reviews" {
			reviewsVirtualService = vs
		}
		log.Printf("Index: %d VirtualService Hosts: %+v\n", i, vs.Spec.GetHosts())
	}

	// Test DestinationRules
	drList, err := drClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}
	for i := range drList.Items {
		dr := drList.Items[i]
		log.Printf("Index: %d DestinationRule Host: %+v\n", i, dr.Spec.GetHost())
	}

	fmt.Printf("Migrate to v3...")
	migrateToV3Route := istio.HTTPRoute{
		Route: []*istio.HTTPRouteDestination{
			{
				Destination: &istio.Destination{
					Host:   "reviews",
					Subset: "v3",
				},
			},
		},
	}
	reviewsVirtualService.Spec.Http = []*istio.HTTPRoute{&migrateToV3Route}
	reviewsVirtualService, err = vsClient.Update(ctx, reviewsVirtualService, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("An error occurred updating reviews virtual service: %v\n", err.Error())
	}
	fmt.Println("Reviews virtual service configured successfully.")
}
