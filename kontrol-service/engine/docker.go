package engine

import (
	"fmt"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"
	"strings"

	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
)

func GenerateProdOnlyCluster(
	flowID string,
	serviceConfigs []apitypes.ServiceConfig,
	ingressConfigs []apitypes.IngressConfig,
	gatewayConfigs []apitypes.GatewayConfig,
	routeConfigs []apitypes.RouteConfig,
	namespace string,
) (*resolved.ClusterTopology, error) {
	clusterTopology, err := generateClusterTopology(serviceConfigs, ingressConfigs, gatewayConfigs, routeConfigs, namespace, flowID)
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
		if devService.DeploymentSpec == nil {
			return nil, stacktrace.NewError("Service with UUID %s has no DeploymentSpec", devServiceName)
		}

		patches = append(patches, flow_spec.ServicePatch{
			Service: devServiceName,
			Image:   item.Image,
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
	ingressConfigs []apitypes.IngressConfig,
	gatewayConfigs []apitypes.GatewayConfig,
	routeConfigs []apitypes.RouteConfig,
	namespace string,
	version string,
) (*resolved.ClusterTopology, error) {
	clusterTopologyGatewayAndRoutes := processGatewayAndRouteConfigs(gatewayConfigs, routeConfigs, version)
	clusterTopologyIngress := processIngressConfigs(ingressConfigs, version)
	clusterTopologyServices, clusterTopologyServiceDependencies, err := processServiceConfigs(serviceConfigs, version)
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

func processServiceConfigs(serviceConfigs []apitypes.ServiceConfig, version string) ([]*resolved.Service, []resolved.ServiceDependency, error) {
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
		clusterTopologyService := newClusterTopologyServiceFromServiceConfig(serviceConfig, version)

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
					ServiceID:      serviceName,
					Version:        version,
					ServiceSpec:    nil, // leave empty for now
					DeploymentSpec: nil, // leave empty for now
					IsExternal:     true,
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

func newClusterTopologyServiceFromServiceConfig(serviceConfig apitypes.ServiceConfig, version string) resolved.Service {
	service := serviceConfig.Service
	deployment := serviceConfig.Deployment
	serviceAnnotations := service.GetObjectMeta().GetAnnotations()

	clusterTopologyService := resolved.Service{
		ServiceID:      service.GetObjectMeta().GetName(),
		Version:        version,
		ServiceSpec:    &service.Spec,
		DeploymentSpec: &deployment.Spec,
	}
	isStateful, ok := serviceAnnotations["kardinal.dev.service/stateful"]
	if ok && isStateful == "true" {
		clusterTopologyService.IsStateful = true
	}
	isExternal, ok := serviceAnnotations["kardinal.dev.service/external"]
	if ok && isExternal == "true" {
		clusterTopologyService.IsExternal = true
	}

	isShared, ok := serviceAnnotations["kardinal.dev.service/shared"]
	if ok && isShared == "true" {
		clusterTopologyService.IsShared = true
	}
	return clusterTopologyService
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
