package engine

import (
	"context"
	"log"
	"os"

	"github.com/samber/lo"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kardinal.kloud-kontrol/types"
)

const (
	istioLabel = "istio-injection"
)

func ConnectToCluster() *rest.Config {
	kubeconfig := os.Getenv("KUBECONFIG")

	if len(kubeconfig) == 0 {
		log.Fatalf("Environment variables KUBECONFIG need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}
	return restConfig
}

func EnsureNamespace(restConfig *rest.Config, namespace string) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}
	existingNamespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil && existingNamespace != nil {
		value, found := existingNamespace.Labels[istioLabel]
		if !found || value != "enabled" {
			existingNamespace.Labels[istioLabel] = "enabled"
			clientset.CoreV1().Namespaces().Update(context.TODO(), existingNamespace, metav1.UpdateOptions{})
		}
	} else {
		newNamespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					istioLabel: "enabled",
				},
			},
		}
		_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), &newNamespace, metav1.CreateOptions{})
		if err != nil {
			log.Panicf("Failed to create Namespace: %v", err)
			panic(err.Error())
		}
	}
}

func ApplyClusterResources(restConfig *rest.Config, clusterResources *types.ClusterResources) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	allNSs := [][]string{
		lo.Uniq(lo.Map(clusterResources.Services, func(item v1.Service, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.Deployments, func(item apps.Deployment, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.VirtualServices, func(item istioclient.VirtualService, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.DestinationRules, func(item istioclient.DestinationRule, _ int) string { return item.Namespace })),
		{clusterResources.Gateway.Namespace},
	}

	uniqueNamespaces := lo.Uniq(lo.Flatten(allNSs))
	lo.ForEach(uniqueNamespaces, func(namespace string, _ int) { EnsureNamespace(restConfig, namespace) })

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

func CleanUpClusterResources(restConfig *rest.Config, clusterResources *types.ClusterResources) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to connect to kubernetes: %s", err)
	}

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	// Clean up services
	servicesByNS := lo.GroupBy(clusterResources.Services, func(item v1.Service) string {
		return item.Namespace
	})
	lo.MapEntries(servicesByNS, func(namespace string, services []v1.Service) (string, []v1.Service) {
		serviceClient := clientset.CoreV1().Services(namespace)
		allServices, err := serviceClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list services: %s", err)
		}
		for _, service := range allServices.Items {
			_, exists := lo.Find(services, func(item v1.Service) bool { return (item.Name == service.Name) })
			if !exists {
				err := serviceClient.Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Failed to delete service: %s", err)
				}
			}
		}
		return namespace, services
	})

	// Clean up deployments
	deploymentsByNS := lo.GroupBy(clusterResources.Deployments, func(item apps.Deployment) string {
		return item.Namespace
	})
	lo.MapEntries(deploymentsByNS, func(namespace string, deployments []apps.Deployment) (string, []apps.Deployment) {
		deploymentClient := clientset.AppsV1().Deployments(namespace)
		allDeployments, err := deploymentClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list deployments: %s", err)
		}
		for _, deployment := range allDeployments.Items {
			_, exists := lo.Find(deployments, func(item apps.Deployment) bool { return item.Name == deployment.Name })
			if !exists {
				err := deploymentClient.Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Failed to delete deployment: %s", err)
				}
			}
		}
		return namespace, deployments
	})

	// Clean up virtual services
	virtualServicesByNS := lo.GroupBy(clusterResources.VirtualServices, func(item istioclient.VirtualService) string {
		return item.Namespace
	})
	lo.MapEntries(virtualServicesByNS, func(namespace string, virtualServices []istioclient.VirtualService) (string, []istioclient.VirtualService) {
		virtServiceClient := ic.NetworkingV1alpha3().VirtualServices(namespace)
		allVirtServices, err := virtServiceClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list virtual services: %s", err)
		}
		for _, virtService := range allVirtServices.Items {
			_, exists := lo.Find(virtualServices, func(item istioclient.VirtualService) bool { return item.Name == virtService.Name })
			if !exists {
				err := virtServiceClient.Delete(context.TODO(), virtService.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Failed to delete virtual service: %s", err)
				}
			}
		}
		return namespace, virtualServices
	})

	// Clean up destination rules
	destinationRulesByNS := lo.GroupBy(clusterResources.DestinationRules, func(item istioclient.DestinationRule) string {
		return item.Namespace
	})
	lo.MapEntries(destinationRulesByNS, func(namespace string, destinationRules []istioclient.DestinationRule) (string, []istioclient.DestinationRule) {
		destRuleClient := ic.NetworkingV1alpha3().DestinationRules(namespace)
		allDestRules, err := destRuleClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list destination rules: %s", err)
		}
		for _, destRule := range allDestRules.Items {
			_, exists := lo.Find(destinationRules, func(item istioclient.DestinationRule) bool { return item.Name == destRule.Name })
			if !exists {
				err := destRuleClient.Delete(context.TODO(), destRule.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Failed to delete destination rule: %s", err)
				}
			}
		}
		return namespace, destinationRules
	})

	// Clean up gateway
	gateway := clusterResources.Gateway
	gatewayClient := ic.NetworkingV1alpha3().Gateways(gateway.Namespace)
	existingGateway, err := gatewayClient.Get(context.TODO(), gateway.Name, metav1.GetOptions{})
	if err == nil {
		_, exists := lo.Find([]istioclient.Gateway{gateway}, func(item istioclient.Gateway) bool { return item.Name == existingGateway.Name })
		if !exists {
			err := gatewayClient.Delete(context.TODO(), gateway.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Failed to delete gateway: %s", err)
			}
		}
	}
}
