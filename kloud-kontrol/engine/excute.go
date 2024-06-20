package engine

import (
	"context"
	"log"
	"os"

	"github.com/samber/lo"
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
		log.Fatalf("Environment variables KUBECONFIG need to be set")
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
		existingService, err := serviceClient.Get(context.TODO(), service.Name, metav1.GetOptions{})
		if err != nil {
			// Resource does not exist, create new one
			_, err = serviceClient.Create(context.TODO(), &service, metav1.CreateOptions{})
			if err != nil {
				log.Fatalf("Failed to create service: %s", err)
			}
		} else {
			// Update the resource version to the latest before updating
			service.ResourceVersion = existingService.ResourceVersion
			_, err = serviceClient.Update(context.TODO(), &service, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("Failed to update service: %s", err)
			}
		}
	})

	lo.ForEach(clusterResources.Deployments, func(deployment apps.Deployment, _ int) {
		deploymentClient := clientset.AppsV1().Deployments(deployment.Namespace)
		existingDeployment, err := deploymentClient.Get(context.TODO(), deployment.Name, metav1.GetOptions{})
		if err != nil {
			_, err = deploymentClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
			if err != nil {
				log.Fatalf("Failed to create deployment: %s", err)
			}
		} else {
			deployment.ResourceVersion = existingDeployment.ResourceVersion
			_, err = deploymentClient.Update(context.TODO(), &deployment, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("Failed to update deployment: %s", err)
			}
		}
	})

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	lo.ForEach(clusterResources.VirtualServices, func(virtService istioclient.VirtualService, _ int) {
		virtServicesClient := ic.NetworkingV1alpha3().VirtualServices(virtService.Namespace)
		existingVirtService, err := virtServicesClient.Get(context.TODO(), virtService.Name, metav1.GetOptions{})
		if err != nil {
			_, err = virtServicesClient.Create(context.TODO(), &virtService, metav1.CreateOptions{})
			if err != nil {
				log.Fatalf("Failed to create virtual service: %s", err)
			}
		} else {
			virtService.ResourceVersion = existingVirtService.ResourceVersion
			_, err = virtServicesClient.Update(context.TODO(), &virtService, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("Failed to update virtual service: %s", err)
			}
		}
	})

	lo.ForEach(clusterResources.DestinationRules, func(destRule istioclient.DestinationRule, _ int) {
		destRulesClient := ic.NetworkingV1alpha3().DestinationRules(destRule.Namespace)
		existingDestRule, err := destRulesClient.Get(context.TODO(), destRule.Name, metav1.GetOptions{})
		if err != nil {
			_, err = destRulesClient.Create(context.TODO(), &destRule, metav1.CreateOptions{})
			if err != nil {
				log.Fatalf("Failed to create destination rule: %s", err)
			}
		} else {
			destRule.ResourceVersion = existingDestRule.ResourceVersion
			_, err = destRulesClient.Update(context.TODO(), &destRule, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("Failed to update destination rule: %s", err)
			}
		}
	})

	gateway := &clusterResources.Gateway
	existingGateway, err := ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Get(context.TODO(), gateway.Name, metav1.GetOptions{})
	if err != nil {
		_, err = ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Create(context.TODO(), gateway, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("Failed to create gateway: %s", err)
		}
	} else {
		gateway.ResourceVersion = existingGateway.ResourceVersion
		_, err = ic.NetworkingV1alpha3().Gateways(gateway.Namespace).Update(context.TODO(), gateway, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("Failed to update gateway: %s", err)
		}
	}
}
