package engine

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kardinal.kloud-kontrol/types"
)

func ApplyClusterResources(clusterResources *types.ClusterResources) {
	kubeconfig := os.Getenv("KUBECONFIG")

	if len(kubeconfig) == 0 {
		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	lo.ForEach(clusterResources.Services, func(service v1.Service, _ int) {
		serviceClient := clientset.CoreV1().Services(service.Namespace)

		result, err := serviceClient.Update(context.TODO(), &service, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("Failed to create service: %s", err)
		}

		resultStr, _ := yaml.Marshal(result)
		fmt.Println(string(resultStr))
	})

	lo.ForEach(clusterResources.Deployments, func(deployment apps.Deployment, _ int) {
		deploymentClient := clientset.AppsV1().Deployments(deployment.Namespace)
		result, err := deploymentClient.Update(context.TODO(), &deployment, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("Failed to create deployment: %s", err)
		}

		resultStr, _ := yaml.Marshal(result)
		fmt.Println(string(resultStr))
	})

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	lo.ForEach(clusterResources.VirtualServices, func(virtService istioclient.VirtualService, _ int) {
		resultStr, _ := yaml.Marshal(virtService)
		fmt.Println(string(resultStr))

		virtServicesClient := ic.NetworkingV1alpha3().VirtualServices(virtService.Namespace)
		_, err := virtServicesClient.Create(context.TODO(), &virtService, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Failed to create istio virtual service: %s", err)
		}
	})

	lo.ForEach(clusterResources.DestinationRules, func(destRule istioclient.DestinationRule, _ int) {
		destRulesClient := ic.NetworkingV1alpha3().DestinationRules(destRule.Namespace)
		result, err := destRulesClient.Create(context.TODO(), &destRule, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Failed to create destination rule: %s", err)
		}

		resultStr, _ := yaml.Marshal(result)
		fmt.Println(string(resultStr))
	})

	gateway := &clusterResources.Gateway
	result, err := ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Create(context.TODO(), gateway, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create gateway: %s", err)
	}
	resultStr, _ := yaml.Marshal(result)
	fmt.Println(string(resultStr))
}
