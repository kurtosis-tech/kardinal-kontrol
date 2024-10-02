package engine

import (
	"fmt"
	"strings"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"

	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
	kardinal "kardinal.kontrol-service/types/kardinal"
)

func GenerateProdOnlyCluster(
	flowID string,
	serviceConfigs []apitypes.ServiceConfig,
	deploymentConfigs []apitypes.DeploymentConfig,
	statefulSetConfigs []apitypes.StatefulSetConfig,
	ingressConfigs []apitypes.IngressConfig,
	gatewayConfigs []apitypes.GatewayConfig,
	routeConfigs []apitypes.RouteConfig,
	namespace string,
) (*resolved.ClusterTopology, error) {
	clusterTopology, err := generateClusterTopology(serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routeConfigs, namespace, flowID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func GenerateProdDevCluster(baseClusterTopologyMaybeWithTemplateOverrides *resolved.ClusterTopology, baseTopology *resolved.ClusterTopology, pluginRunner *plugins.PluginRunner, flowSpec flow_spec.FlowPatchSpec) (*resolved.ClusterTopology, error) {
	patches := []flow_spec.ServicePatch{}
	for _, item := range flowSpec.ServicePatches {
		devServiceName := item.Service
		devService, err := baseClusterTopologyMaybeWithTemplateOverrides.GetService(devServiceName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Service with UUID %s not found", devServiceName)
		}

		workloadSpec := devService.WorkloadSpec
		clonedWorkloadSpec := workloadSpec.DeepCopy()

		// TODO: find a better way to update deploymentSpec, this assumes there is only container in the pod
		clonedWorkloadSpec.GetTemplateSpec().Containers[0].Image = item.Image

		patches = append(patches, flow_spec.ServicePatch{
			Service:      devServiceName,
			WorkloadSpec: &clonedWorkloadSpec,
		})
	}

	flowPatch := flow_spec.FlowPatch{
		FlowId:         flowSpec.FlowId,
		ServicePatches: patches,
	}

	clusterTopology, err := flow.CreateDevFlow(pluginRunner, *baseClusterTopologyMaybeWithTemplateOverrides, *baseTopology, flowPatch)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func generateClusterTopology(
	serviceConfigs []apitypes.ServiceConfig,
	deploymentConfigs []apitypes.DeploymentConfig,
	statefulSetConfig []apitypes.StatefulSetConfig,
	ingressConfigs []apitypes.IngressConfig,
	gatewayConfigs []apitypes.GatewayConfig,
	routeConfigs []apitypes.RouteConfig,
	namespace string,
	version string,
) (*resolved.ClusterTopology, error) {
	clusterTopologyGatewayAndRoutes := processGatewayAndRouteConfigs(gatewayConfigs, routeConfigs, version)
	clusterTopologyIngress := processIngressConfigs(ingressConfigs, version)
	clusterTopologyServices, clusterTopologyServiceDependencies, err := processServiceConfigs(
		serviceConfigs,
		deploymentConfigs,
		statefulSetConfig,
		version)
	if err != nil {
		return nil, stacktrace.NewError("an error occurred processing the service configs")
	}

	// some validations
	if len(clusterTopologyIngress.Ingresses) == 0 && len(clusterTopologyGatewayAndRoutes.Gateways) == 0 && len(clusterTopologyGatewayAndRoutes.GatewayRoutes) == 0 {
		logrus.Warnf("No ingress or gateway found in the service configs")
	}
	if len(clusterTopologyServices) == 0 {
		return nil, stacktrace.NewError("At least one service is required in addition to the ingress service(s)")
	}

	clusterTopology := resolved.ClusterTopology{}
	clusterTopology.Namespace = namespace
	clusterTopology.Ingress = clusterTopologyIngress
	clusterTopology.GatewayAndRoutes = clusterTopologyGatewayAndRoutes
	clusterTopology.Services = clusterTopologyServices
	clusterTopology.ServiceDependencies = clusterTopologyServiceDependencies

	return &clusterTopology, nil
}

func processGatewayAndRouteConfigs(gatewayConfigs []apitypes.GatewayConfig, routeConfigs []apitypes.RouteConfig, version string) *resolved.GatewayAndRoutes {
	gatewayAndRoutes := &resolved.GatewayAndRoutes{
		ActiveFlowIDs: []string{version},
		Gateways:      []*gateway.Gateway{},
		GatewayRoutes: []*gateway.HTTPRouteSpec{},
	}
	for _, gatewayConfig := range gatewayConfigs {
		gateway := gatewayConfig.Gateway
		gatewayAnnotations := gateway.GetObjectMeta().GetAnnotations()
		isGateway, ok := gatewayAnnotations["kardinal.dev.service/gateway"]
		if ok && isGateway == "true" {
			if gateway.Spec.Listeners == nil {
				logrus.Warnf("Gateway %v is missing listeners", gateway.Name)
			} else {
				for _, listener := range gateway.Spec.Listeners {
					if listener.Hostname != nil && !strings.HasPrefix(string(*listener.Hostname), "*.") {
						logrus.Warnf("Gateway %v listener %v is missing a wildcard, creating flow entry points will not work properly.", gateway.Name, listener.Hostname)
					}
				}
			}
			logrus.Infof("Managing gateway: %v", gateway.Name)
			gatewayAndRoutes.Gateways = append(gatewayAndRoutes.Gateways, &gateway)
		} else {
			logrus.Infof("Gateway %v is not a Kardinal gateway", gateway.Name)
		}
	}
	for _, routeConfig := range routeConfigs {
		route := routeConfig.HttpRoute
		routeAnnotations := route.GetObjectMeta().GetAnnotations()
		isRoute, ok := routeAnnotations["kardinal.dev.service/route"]
		if ok && isRoute == "true" {
			gatewayAndRoutes.GatewayRoutes = append(gatewayAndRoutes.GatewayRoutes, &route.Spec)
		}
	}
	return gatewayAndRoutes
}

func getDeploymentForService(
	serviceConfig apitypes.ServiceConfig,
	workloadConfigs []apitypes.DeploymentConfig,
) *apitypes.DeploymentConfig {
	service := serviceConfig.Service
	workload, foundworkload := lo.Find(workloadConfigs, func(workloadConfig apitypes.DeploymentConfig) bool {
		deploymentLabels := workloadConfig.Deployment.GetLabels()
		matchSelectors := true
		for key, value := range service.Spec.Selector {
			label, found := deploymentLabels[key]
			if !found || value != label {
				return false
			}
		}
		return matchSelectors
	})

	if foundworkload {
		return &workload
	}

	return nil
}

func getSatefulSetForService(
	serviceConfig apitypes.ServiceConfig,
	workloadConfigs []apitypes.StatefulSetConfig,
) *apitypes.StatefulSetConfig {
	service := serviceConfig.Service
	workload, foundworkload := lo.Find(workloadConfigs, func(workloadConfig apitypes.StatefulSetConfig) bool {
		workloadLabel := workloadConfig.StatefulSet.GetLabels()
		matchSelectors := true
		for key, value := range service.Spec.Selector {
			label, found := workloadLabel[key]
			if !found || value != label {
				return false
			}
		}
		return matchSelectors
	})

	if foundworkload {
		return &workload
	}

	return nil
}

func processServiceConfigs(
	serviceConfigs []apitypes.ServiceConfig,
	deploymentConfigs []apitypes.DeploymentConfig,
	statefulSetConfigs []apitypes.StatefulSetConfig,
	version string,
) ([]*resolved.Service, []resolved.ServiceDependency, error) {
	clusterTopologyServices := []*resolved.Service{}
	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	externalServicesDependencies := []resolved.ServiceDependency{}

	type serviceWithDependenciesAnnotation struct {
		service                *resolved.Service
		dependenciesAnnotation string
	}
	serviceWithDependencies := []*serviceWithDependenciesAnnotation{}

	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// 1- Service
		logrus.Infof("Processing service: %v", service.GetObjectMeta().GetName())

		deploymentConfig := getDeploymentForService(serviceConfig, deploymentConfigs)
		statefulSetConfig := getSatefulSetForService(serviceConfig, statefulSetConfigs)
		clusterTopologyService, error := newClusterTopologyServiceFromServiceConfig(serviceConfig, deploymentConfig, statefulSetConfig, version)
		if error != nil {
			return nil, nil, stacktrace.Propagate(error, "An error occurred creating new cluster topology service from service config '%s'", service.Name)
		}

		// 2- Service plugins
		serviceStatefulPlugins, externalServices, newExternalServicesDependencies, err := newStatefulPluginsAndExternalServicesFromServiceConfig(serviceConfig, version, &clusterTopologyService)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred creating new stateful plugins and external services from service config '%s'", service.Name)
		}
		clusterTopologyService.StatefulPlugins = serviceStatefulPlugins
		clusterTopologyServices = append(clusterTopologyServices, externalServices...)
		externalServicesDependencies = append(externalServicesDependencies, newExternalServicesDependencies...)

		// 3- Service dependencies (creates a list of services with dependencies)
		dependencies, ok := serviceAnnotations["kardinal.dev.service/dependencies"]
		if ok {
			newServiceWithDependenciesAnnotation := &serviceWithDependenciesAnnotation{&clusterTopologyService, dependencies}
			serviceWithDependencies = append(serviceWithDependencies, newServiceWithDependenciesAnnotation)
		}
		clusterTopologyServices = append(clusterTopologyServices, &clusterTopologyService)
	}

	// Set the service dependencies in the clusterTopologyService
	// first iterate on the service with dependencies list
	for _, svcWithDependenciesAnnotation := range serviceWithDependencies {

		serviceAndPorts := strings.Split(svcWithDependenciesAnnotation.dependenciesAnnotation, ",")
		for _, serviceAndPort := range serviceAndPorts {
			serviceAndPortParts := strings.Split(serviceAndPort, ":")
			depService, depServicePort, err := getServiceAndPortFromClusterTopologyServices(serviceAndPortParts[0], serviceAndPortParts[1], clusterTopologyServices)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred finding the service dependency for service %s and port %s", serviceAndPortParts[0], serviceAndPortParts[1])
			}

			serviceDependency := resolved.ServiceDependency{
				Service:          svcWithDependenciesAnnotation.service,
				DependsOnService: depService,
				DependencyPort:   depServicePort,
			}

			clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, serviceDependency)
		}
	}
	// then add the external services dependencies
	clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, externalServicesDependencies...)

	return clusterTopologyServices, clusterTopologyServiceDependencies, nil
}

func newStatefulPluginsAndExternalServicesFromServiceConfig(serviceConfig apitypes.ServiceConfig, version string, clusterTopologyService *resolved.Service) ([]*resolved.StatefulPlugin, []*resolved.Service, []resolved.ServiceDependency, error) {
	var serviceStatefulPlugins []*resolved.StatefulPlugin
	externalServices := []*resolved.Service{}
	externalServiceDependencies := []resolved.ServiceDependency{}

	service := serviceConfig.Service
	serviceAnnotations := service.GetObjectMeta().GetAnnotations()

	sPluginsAnnotation, ok := serviceAnnotations["kardinal.dev.service/plugins"]
	if ok {
		var statefulPlugins []resolved.StatefulPlugin
		err := yaml.Unmarshal([]byte(sPluginsAnnotation), &statefulPlugins)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred parsing the plugins for service %s", service.GetObjectMeta().GetName())
		}
		serviceStatefulPlugins = make([]*resolved.StatefulPlugin, len(statefulPlugins))

		for index := range statefulPlugins {
			// TODO: consider giving external service plugins their own type, instead of using StatefulPlugins
			// if this is an external service plugin, represent that service as a service in the cluster topology
			plugin := statefulPlugins[index]
			if plugin.Type == "external" {
				logrus.Infof("Adding external service to topology..")
				serviceName := plugin.ServiceName
				logrus.Infof("plugin service name: %v", plugin.ServiceName)
				if serviceName == "" {
					serviceID := service.GetObjectMeta().GetName()
					serviceName = fmt.Sprintf("%v:%v", serviceID, "external")
				}
				externalService := resolved.Service{
					ServiceID:    serviceName,
					Version:      version,
					ServiceSpec:  nil, // leave empty for now
					WorkloadSpec: nil, // leave empty for now
					IsExternal:   true,
					// external services can definitely be stateful but for now treat external and stateful services as mutually exclusive to make plugin logic easier to handle
					IsStateful: false,
				}

				externalServices = append(externalServices, &externalService)

				externalServiceDependency := resolved.ServiceDependency{
					Service:          clusterTopologyService,
					DependsOnService: &externalService,
					DependencyPort:   nil,
				}
				externalServiceDependencies = append(externalServiceDependencies, externalServiceDependency)
			}
			serviceStatefulPlugins[index] = &plugin
		}
	}

	return serviceStatefulPlugins, externalServices, externalServiceDependencies, nil
}

func newClusterTopologyServiceFromServiceConfig(
	serviceConfig apitypes.ServiceConfig,
	deploymentConfig *apitypes.DeploymentConfig,
	statefulSetConfig *apitypes.StatefulSetConfig,
	version string,
) (resolved.Service, error) {
	service := serviceConfig.Service
	serviceName := service.GetObjectMeta().GetName()

	if deploymentConfig == nil && statefulSetConfig == nil {
		logrus.Warnf("Service %s has no workload", serviceName)
	}

	if deploymentConfig != nil && statefulSetConfig != nil {
		workloads := []string{
			deploymentConfig.Deployment.GetObjectMeta().GetName(),
			statefulSetConfig.StatefulSet.GetObjectMeta().GetName(),
		}
		logrus.Error("Service %s is associated with more than one workload: %v", serviceName, workloads)
	}

	serviceAnnotations := service.GetObjectMeta().GetAnnotations()

	clusterTopologyService := resolved.Service{
		ServiceID:   service.GetObjectMeta().GetName(),
		Version:     version,
		ServiceSpec: &service.Spec,
	}

	if deploymentConfig != nil {
		workload := kardinal.NewDeploymentWorkloadSpec(deploymentConfig.Deployment.Spec)
		clusterTopologyService.WorkloadSpec = &workload
	}
	if statefulSetConfig != nil {
		workload := kardinal.NewStatefulSetWorkloadSpec(statefulSetConfig.StatefulSet.Spec)
		clusterTopologyService.WorkloadSpec = &workload
	}

	if clusterTopologyService.WorkloadSpec == nil {
		return clusterTopologyService, stacktrace.NewError("Service %s has no workload", serviceName)
	}

	// Set default for IsStateful to true if the workload is a StatefulSet, otherwise false
	clusterTopologyService.IsExternal = clusterTopologyService.WorkloadSpec.IsStatefulSet()

	// Override the IsStateful value by manual annotations
	isStateful, ok := serviceAnnotations["kardinal.dev.service/stateful"]
	if ok && isStateful == "true" {
		clusterTopologyService.IsStateful = true
	}
	if ok && isStateful == "false" {
		clusterTopologyService.IsStateful = false
	}

	isExternal, ok := serviceAnnotations["kardinal.dev.service/external"]
	if ok && isExternal == "true" {
		clusterTopologyService.IsExternal = true
	}

	isShared, ok := serviceAnnotations["kardinal.dev.service/shared"]
	if ok && isShared == "true" {
		clusterTopologyService.IsShared = true
	}
	return clusterTopologyService, nil
}

func getServiceAndPortFromClusterTopologyServices(serviceName string, servicePortName string, clusterTopologyServices []*resolved.Service) (*resolved.Service, *corev1.ServicePort, error) {
	for _, service := range clusterTopologyServices {
		if service.ServiceID == serviceName {
			for _, port := range service.ServiceSpec.Ports {
				if port.Name == servicePortName {
					return service, &port, nil
				}
			}
		}
	}

	return nil, nil, stacktrace.NewError("Service %s and Port %s not found in the list of services", serviceName, servicePortName)
}

func processIngressConfigs(ingressConfigs []apitypes.IngressConfig, version string) *resolved.Ingress {
	clusterTopologyIngress := &resolved.Ingress{
		ActiveFlowIDs: []string{version},
		Ingresses:     []net.Ingress{},
	}
	for _, ingressConfig := range ingressConfigs {
		ingress := ingressConfig.Ingress
		ingressAnnotations := ingress.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := ingressAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			clusterTopologyIngress.Ingresses = append(clusterTopologyIngress.Ingresses, ingress)
		}
	}
	return clusterTopologyIngress
}
