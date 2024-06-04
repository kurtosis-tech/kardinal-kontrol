package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	istio "istio.io/api/networking/v1alpha3"
	istioconfig "istio.io/client-go/pkg/applyconfiguration/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	namespace := ""
	vsClient := ic.NetworkingV1alpha3().VirtualServices(namespace)
	drClient := ic.NetworkingV1alpha3().DestinationRules(namespace)

	// Test VirtualServices
	vsList, err := vsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to retrieve VirtualServices in namespace: %s with error:\n%s", namespace, err)
	}
	for i := range vsList.Items {
		vs := vsList.Items[i]
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

	// turn this command into a programmatic k8s api call:
	// kubectl apply -f samples/bookinfo/networking/virtual-service-all-v1.yaml
	// creates Virtual Services for each service that routes all traffic to one subset of that service
	fmt.Printf("Attempting to apply routing rule ...\n")
	//apiVersion: networking.istio.io/v1alpha3
	//kind: VirtualService
	//metadata:
	//name: reviews
	//spec:
	//hosts:
	//	- reviews
	//http:
	//	- match:
	//	- headers:
	//	end-user:
	//exact: jason
	//route:
	//	- destination:
	//host: reviews
	//subset: v2
	//	- route:
	//	- destination:
	//host: reviews
	//subset: v1
	vs := istio.VirtualService{
		Hosts: []string{"reviews"},
		Http: []*istio.HTTPRoute{
			{
				Route: []*istio.HTTPRouteDestination{
					{
						Destination: &istio.Destination{
							Host:   "reviews",
							Subset: "v1",
						},
					},
				},
			},
			{
				Match: []*istio.HTTPMatchRequest{
					{
						Uri: &istio.StringMatch{
							MatchType: &istio.StringMatch_Prefix{
								Prefix: "/productpage/v2",
							},
						},
					},
				},
				Route: []*istio.HTTPRouteDestination{
					{
						Destination: &istio.Destination{
							Host:   "reviews",
							Subset: "v2",
						},
					},
				},
			},
		},
	}
	virtualServiceConfig := istioconfig.VirtualService("reviews", "")
	virtualServiceConfig.
		WithAPIVersion("networking.istio.io/v1alpha3").
		WithKind("VirtualService").
		WithAnnotations(map[string]string{"name": "reviews"}).
		WithSpec(vs)
	_, err = vsClient.Apply(ctx, virtualServiceConfig, metav1.ApplyOptions{})
	if err != nil {
		fmt.Printf("An error occurred applying virtual service config: %s", err)
	}
	fmt.Println("Virtual Service configured successfully.")

	// Test Gateway
	//gwList, err := ic.NetworkingV1alpha3().Gateways(namespace).List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	log.Fatalf("Failed to get Gateway in %s namespace: %s", namespace, err)
	//}
	//
	//for i := range gwList.Items {
	//	gw := gwList.Items[i]
	//	for _, s := range gw.Spec.GetServers() {
	//		log.Printf("Index: %d Gateway servers: %+v\n", i, s)
	//	}
	//}

	time.Sleep(100000 * time.Minute)
}
