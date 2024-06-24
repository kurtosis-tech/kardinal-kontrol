package cluster_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kardinal.kontrol/kardinal-manager/types"
)

const (
	listOptionsTimeoutSeconds       int64 = 10
	fieldManager                          = "kardinal-manager"
	deleteOptionsGracePeriodSeconds int64 = 0
	istioLabel                            = "istio-injection"
	enabledIstioValue                     = "enabled"
)

var (
	globalListOptions = metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       int64Ptr(listOptionsTimeoutSeconds),
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	}

	globalCreateOptions = metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun: nil,
		// We need every object to have this field manager so that the Kurtosis objects can all seamlessly modify Kubernetes resources
		FieldManager:    fieldManager,
		FieldValidation: "",
	}

	globalUpdateOptions = metav1.UpdateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun: nil,
		// We need every object to have this field manager so that the Kurtosis objects can all seamlessly modify Kubernetes resources
		FieldManager:    fieldManager,
		FieldValidation: "",
	}

	globalDeletePolicy = metav1.DeletePropagationForeground

	globalDeleteOptions = metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds: int64Ptr(deleteOptionsGracePeriodSeconds),
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  &globalDeletePolicy,
		DryRun:             nil,
	}
)

type ClusterManager struct {
	kubernetesClient *kubernetesClient
	istioClient      *istioClient
}

func NewClusterManager(kubernetesClient *kubernetesClient, istioClient *istioClient) *ClusterManager {
	return &ClusterManager{kubernetesClient: kubernetesClient, istioClient: istioClient}
}

func (manager *ClusterManager) ApplyClusterResources(ctx context.Context, clusterResources *types.ClusterResources) error {

	allNSs := [][]string{
		lo.Uniq(lo.Map(clusterResources.Services, func(item apiv1.Service, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.Deployments, func(item apps.Deployment, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.VirtualServices, func(item istioclient.VirtualService, _ int) string { return item.Namespace })),
		lo.Uniq(lo.Map(clusterResources.DestinationRules, func(item istioclient.DestinationRule, _ int) string { return item.Namespace })),
		{clusterResources.Gateway.Namespace},
	}

	uniqueNamespaces := lo.Uniq(lo.Flatten(allNSs))

	var ensureNamespacesErr error
	lo.ForEach(uniqueNamespaces, func(namespace string, _ int) {
		ensureNamespacesErr = manager.ensureNamespace(ctx, namespace)
	})
	if ensureNamespacesErr != nil {
		return stacktrace.Propagate(ensureNamespacesErr, "An error occurred while creating or updating cluster namespaces")
	}

	var createOrUpdateServicesErr error
	lo.ForEach(clusterResources.Services, func(service apiv1.Service, _ int) {
		createOrUpdateServicesErr = manager.createOrUpdateService(ctx, &service)
	})
	if createOrUpdateServicesErr != nil {
		return stacktrace.Propagate(createOrUpdateServicesErr, "An error occurred while creating or updating cluster services")
	}

	var createOrUpdateDeploymentsErr error
	lo.ForEach(clusterResources.Deployments, func(deployment apps.Deployment, _ int) {
		createOrUpdateDeploymentsErr = manager.createOrUpdateDeployment(ctx, &deployment)
	})
	if createOrUpdateDeploymentsErr != nil {
		return stacktrace.Propagate(createOrUpdateDeploymentsErr, "An error occurred while creating or updating cluster deployments")
	}

	var createOrUpdateVirtualServicesErr error
	lo.ForEach(clusterResources.VirtualServices, func(virtualService istioclient.VirtualService, _ int) {
		createOrUpdateVirtualServicesErr = manager.createOrUpdateVirtualService(ctx, &virtualService)
	})
	if createOrUpdateVirtualServicesErr != nil {
		return stacktrace.Propagate(createOrUpdateVirtualServicesErr, "An error occurred while creating or updating cluster virtual services")
	}

	var createOrUpdateDestinationRulesErr error
	lo.ForEach(clusterResources.DestinationRules, func(destinationRule istioclient.DestinationRule, _ int) {
		createOrUpdateDestinationRulesErr = manager.createOrUpdateDestinationRule(ctx, &destinationRule)
	})
	if createOrUpdateDestinationRulesErr != nil {
		return stacktrace.Propagate(createOrUpdateDestinationRulesErr, "An error occurred while creating or updating cluster destination rules")
	}

	if err := manager.createOrUpdateGateway(ctx, &clusterResources.Gateway); err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating or updating the cluster gateway")
	}

	return nil
}

func (manager *ClusterManager) CleanUpClusterResources(ctx context.Context, clusterResources *types.ClusterResources) error {

	// Clean up services
	servicesByNS := lo.GroupBy(clusterResources.Services, func(item apiv1.Service) string {
		return item.Namespace
	})
	var cleanUpServicesErr error
	lo.MapEntries(servicesByNS, func(namespace string, services []apiv1.Service) (string, []apiv1.Service) {
		cleanUpServicesErr = manager.cleanUpServicesInNamespace(ctx, namespace, services)
		return namespace, services
	})
	if cleanUpServicesErr != nil {
		return stacktrace.Propagate(cleanUpServicesErr, "An error occurred while cleaning up cluster services")
	}

	// Clean up deployments
	deploymentsByNS := lo.GroupBy(clusterResources.Deployments, func(item apps.Deployment) string {
		return item.Namespace
	})
	var cleanUpDeploymentsErr error
	lo.MapEntries(deploymentsByNS, func(namespace string, deployments []apps.Deployment) (string, []apps.Deployment) {
		cleanUpDeploymentsErr = manager.cleanUpDeploymentsInNamespace(ctx, namespace, deployments)
		return namespace, deployments
	})
	if cleanUpDeploymentsErr != nil {
		return stacktrace.Propagate(cleanUpDeploymentsErr, "An error occurred while cleaning up cluster deployments")
	}

	// Clean up virtual services
	virtualServicesByNS := lo.GroupBy(clusterResources.VirtualServices, func(item istioclient.VirtualService) string {
		return item.Namespace
	})
	var cleanUpVirtualServicesErr error
	lo.MapEntries(virtualServicesByNS, func(namespace string, virtualServices []istioclient.VirtualService) (string, []istioclient.VirtualService) {
		cleanUpVirtualServicesErr = manager.cleanUpVirtualServicesInNamespace(ctx, namespace, virtualServices)
		return namespace, virtualServices
	})
	if cleanUpVirtualServicesErr != nil {
		return stacktrace.Propagate(cleanUpVirtualServicesErr, "An error occurred while cleaning up cluster virtual services")
	}

	// Clean up destination rules
	destinationRulesByNS := lo.GroupBy(clusterResources.DestinationRules, func(item istioclient.DestinationRule) string {
		return item.Namespace
	})
	var cleanUpDestinationRulesErr error
	lo.MapEntries(destinationRulesByNS, func(namespace string, destinationRules []istioclient.DestinationRule) (string, []istioclient.DestinationRule) {
		cleanUpDestinationRulesErr = manager.cleanUpDestinationRulesInNamespace(ctx, namespace, destinationRules)
		return namespace, destinationRules
	})
	if cleanUpDestinationRulesErr != nil {
		return stacktrace.Propagate(cleanUpDestinationRulesErr, "An error occurred while cleaning up cluster destination rules")
	}

	// Clean up gateway
	gatewaysByNs := map[string][]istioclient.Gateway{
		clusterResources.Gateway.GetNamespace(): {clusterResources.Gateway},
	}
	var cleanUpGatewaysErr error
	lo.MapEntries(gatewaysByNs, func(namespace string, gateways []istioclient.Gateway) (string, []istioclient.Gateway) {
		cleanUpGatewaysErr = manager.cleanUpGatewaysInNamespace(ctx, namespace, gateways)
		return namespace, gateways
	})
	if cleanUpGatewaysErr != nil {
		return stacktrace.Propagate(cleanUpGatewaysErr, "An error occurred while cleaning up cluster gateways")
	}

	return nil
}

func (manager *ClusterManager) ensureNamespace(ctx context.Context, name string) error {

	existingNamespace, err := manager.kubernetesClient.clientSet.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil && existingNamespace != nil {
		value, found := existingNamespace.Labels[istioLabel]
		if !found || value != enabledIstioValue {
			existingNamespace.Labels[istioLabel] = enabledIstioValue
			manager.kubernetesClient.clientSet.CoreV1().Namespaces().Update(ctx, existingNamespace, globalUpdateOptions)
		}
	} else {
		newNamespace := apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					istioLabel: enabledIstioValue,
				},
			},
		}
		_, err = manager.kubernetesClient.clientSet.CoreV1().Namespaces().Create(ctx, &newNamespace, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create Namespace: %s", name)
		}
	}

	return nil
}

func (manager *ClusterManager) createOrUpdateService(ctx context.Context, service *apiv1.Service) error {
	serviceClient := manager.kubernetesClient.clientSet.CoreV1().Services(service.Namespace)
	existingService, err := serviceClient.Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		// Resource does not exist, create new one
		_, err = serviceClient.Create(ctx, service, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create service: %s", service.GetName())
		}
	} else {
		// Update the resource version to the latest before updating
		service.ResourceVersion = existingService.ResourceVersion
		_, err = serviceClient.Update(ctx, service, globalUpdateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update service: %s", service.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) createOrUpdateDeployment(ctx context.Context, deployment *apps.Deployment) error {
	deploymentClient := manager.kubernetesClient.clientSet.AppsV1().Deployments(deployment.Namespace)
	existingDeployment, err := deploymentClient.Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil {
		_, err = deploymentClient.Create(ctx, deployment, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create deployment: %s", deployment.GetName())
		}
	} else {
		deployment.ResourceVersion = existingDeployment.ResourceVersion
		_, err = deploymentClient.Update(ctx, deployment, globalUpdateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update deployment: %s", deployment.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) createOrUpdateVirtualService(ctx context.Context, virtualService *istioclient.VirtualService) error {

	virtServiceClient := manager.istioClient.clientSet.NetworkingV1alpha3().VirtualServices(virtualService.GetNamespace())

	existingVirtService, err := virtServiceClient.Get(ctx, virtualService.Name, metav1.GetOptions{})
	if err != nil {
		_, err = virtServiceClient.Create(ctx, virtualService, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create virtual service: %s", virtualService.GetName())
		}
	} else {
		virtualService.ResourceVersion = existingVirtService.ResourceVersion
		_, err = virtServiceClient.Update(ctx, virtualService, globalUpdateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update virtual service: %s", virtualService.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) createOrUpdateDestinationRule(ctx context.Context, destinationRule *istioclient.DestinationRule) error {

	destRuleClient := manager.istioClient.clientSet.NetworkingV1alpha3().DestinationRules(destinationRule.GetNamespace())

	existingDestRule, err := destRuleClient.Get(ctx, destinationRule.Name, metav1.GetOptions{})
	if err != nil {
		_, err = destRuleClient.Create(ctx, destinationRule, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create destination rule: %s", destinationRule.GetName())
		}
	} else {
		destinationRule.ResourceVersion = existingDestRule.ResourceVersion
		_, err = destRuleClient.Update(ctx, destinationRule, globalUpdateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update destination rule: %s", destinationRule.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) createOrUpdateGateway(ctx context.Context, gateway *istioclient.Gateway) error {

	gatewayClient := manager.istioClient.clientSet.NetworkingV1alpha3().Gateways(gateway.GetNamespace())
	existingGateway, err := gatewayClient.Get(ctx, gateway.Name, metav1.GetOptions{})
	if err != nil {
		_, err = gatewayClient.Create(ctx, gateway, globalCreateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create gateway: %s", gateway.GetName())
		}
	} else {
		gateway.ResourceVersion = existingGateway.ResourceVersion
		_, err = gatewayClient.Update(ctx, gateway, globalUpdateOptions)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to update gateway: %s", gateway.GetName())
		}
	}

	return nil
}

func (manager *ClusterManager) cleanUpServicesInNamespace(ctx context.Context, namespace string, servicesToKeep []apiv1.Service) error {
	serviceClient := manager.kubernetesClient.clientSet.CoreV1().Services(namespace)
	allServices, err := serviceClient.List(ctx, globalListOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to list services in namespace %s", namespace)
	}
	for _, service := range allServices.Items {
		_, exists := lo.Find(servicesToKeep, func(item apiv1.Service) bool { return item.Name == service.Name })
		if !exists {
			err = serviceClient.Delete(ctx, service.Name, globalDeleteOptions)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to delete service %s", service.GetName())
			}
		}
	}
	return nil
}

func (manager *ClusterManager) cleanUpDeploymentsInNamespace(ctx context.Context, namespace string, deploymentsToKeep []apps.Deployment) error {
	deploymentClient := manager.kubernetesClient.clientSet.AppsV1().Deployments(namespace)
	allDeployments, err := deploymentClient.List(ctx, globalListOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to list deployments in namespace %s", namespace)
	}
	for _, deployment := range allDeployments.Items {
		_, exists := lo.Find(deploymentsToKeep, func(item apps.Deployment) bool { return item.Name == deployment.Name })
		if !exists {
			err = deploymentClient.Delete(ctx, deployment.Name, globalDeleteOptions)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to delete deployment %s", deployment.GetName())
			}
		}
	}
	return nil
}

func (manager *ClusterManager) cleanUpVirtualServicesInNamespace(ctx context.Context, namespace string, virtualServicesToKeep []istioclient.VirtualService) error {

	virtServiceClient := manager.istioClient.clientSet.NetworkingV1alpha3().VirtualServices(namespace)
	allVirtServices, err := virtServiceClient.List(ctx, globalListOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to list virtual services in namespace %s", namespace)
	}
	for _, virtService := range allVirtServices.Items {
		_, exists := lo.Find(virtualServicesToKeep, func(item istioclient.VirtualService) bool { return item.Name == virtService.Name })
		if !exists {
			err = virtServiceClient.Delete(ctx, virtService.Name, globalDeleteOptions)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to delete virtual service %s", virtService.GetName())
			}
		}
	}

	return nil
}

func (manager *ClusterManager) cleanUpDestinationRulesInNamespace(ctx context.Context, namespace string, destinationRulesToKeep []istioclient.DestinationRule) error {

	destRuleClient := manager.istioClient.clientSet.NetworkingV1alpha3().DestinationRules(namespace)
	allDestRules, err := destRuleClient.List(ctx, globalListOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to list destination rules in namespace %s", namespace)
	}
	for _, destRule := range allDestRules.Items {
		_, exists := lo.Find(destinationRulesToKeep, func(item istioclient.DestinationRule) bool { return item.Name == destRule.Name })
		if !exists {
			err = destRuleClient.Delete(ctx, destRule.Name, globalDeleteOptions)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to delete destination rule %s", destRule.GetName())
			}
		}
	}

	return nil
}

func (manager *ClusterManager) cleanUpGatewaysInNamespace(ctx context.Context, namespace string, gatewaysToKeep []istioclient.Gateway) error {

	gatewayClient := manager.istioClient.clientSet.NetworkingV1alpha3().Gateways(namespace)
	allGateways, err := gatewayClient.List(ctx, globalListOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to list gateways in namespace %s", namespace)
	}
	for _, gateway := range allGateways.Items {
		_, exists := lo.Find(gatewaysToKeep, func(item istioclient.Gateway) bool { return item.Name == gateway.Name })
		if !exists {
			err = gatewayClient.Delete(ctx, gateway.Name, globalDeleteOptions)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to delete gateway %s", gateway.GetName())
			}
		}
	}

	return nil
}

func int64Ptr(i int64) *int64 { return &i }
