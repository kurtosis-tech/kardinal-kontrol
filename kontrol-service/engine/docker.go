package engine

import (
	"fmt"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	"strings"

	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
)

// GenerateProdOnlyCluster create the baseline cluster which can be also called prod cluster which was the first name used
func GenerateProdOnlyCluster(flowID string, serviceConfigs []apitypes.ServiceConfig, ingressConfigs []apitypes.IngressConfig, namespace string) (*resolved.ClusterTopology, error) {
	clusterTopology, err := generateClusterTopology(serviceConfigs, ingressConfigs, namespace, flowID)
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

		deploymentSpec := flow.DeepCopyDeploymentSpec(devService.DeploymentSpec)

		// TODO: find a better way to update deploymentSpec, this assumes there is only container in the pod
		deploymentSpec.Template.Spec.Containers[0].Image = item.Image

		patches = append(patches, flow_spec.ServicePatch{
			Service:        devServiceName,
			DeploymentSpec: deploymentSpec,
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

func generateClusterTopology(serviceConfigs []apitypes.ServiceConfig, ingressConfigs []apitypes.IngressConfig, namespace, version string) (*resolved.ClusterTopology, error) {
	clusterTopology := resolved.ClusterTopology{}
	clusterTopology.Namespace = namespace

	clusterTopologyIngress := processIngressConfigs(ingressConfigs, version)

	clusterTopologyIngressModified, clusterTopologyServices, clusterTopologyServiceDependencies, err := processServiceConfigs(serviceConfigs, version, clusterTopologyIngress)
	if err != nil {
		return nil, stacktrace.NewError("an error occurred processing the service configs")
	}

	if len(clusterTopologyIngressModified) == 0 {
		return nil, stacktrace.NewError("At least one service needs to be annotated as an ingress service")
	}
	clusterTopology.Ingresses = clusterTopologyIngressModified

	if len(clusterTopologyServices) == 0 {
		return nil, stacktrace.NewError("At least one service is required in addition to the ingress service(s)")
	}
	clusterTopology.Services = clusterTopologyServices
	clusterTopology.ServiceDependencies = clusterTopologyServiceDependencies

	return &clusterTopology, nil
}

func processServiceConfigs(serviceConfigs []apitypes.ServiceConfig, version string, clusterTopologyIngress []*resolved.Ingress) ([]*resolved.Ingress, []*resolved.Service, []resolved.ServiceDependency, error) {
	var err error
	clusterTopologyServices := []*resolved.Service{}
	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	externalServicesDependencies := []resolved.ServiceDependency{}
	availablePlugins := map[string]*resolved.StatefulPlugin{}

	type serviceWithDependenciesAnnotation struct {
		service                *resolved.Service
		dependenciesAnnotation string
	}
	serviceWithDependencies := []*serviceWithDependenciesAnnotation{}

	// First, iterate the services to get all the available plugins
	for _, serviceConfig := range serviceConfigs {
		// availablePlugins list contains both stateful and external plugins and, externalServices is a list of Kardinal services that are also linked with a plugin inside the availablePlugins list
		availablePlugins, err = addAvailablePluginsFromServiceConfig(serviceConfig, availablePlugins)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred while parsing plugin '%s'", serviceConfig.Service.GetName())
		}
	}

	// Second, iterate the services to create the clusterTopology service with partial data (no dependencies set here)
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// 1- Ingress
		ingressNotFoundYet := len(clusterTopologyIngress) == 0
		// find ingress from a service config only if it wasn't found before from the ingress configs
		if ingressNotFoundYet && isIngres(serviceAnnotations) {
			ingress, err := newClusterTopologyIngresFromServiceConfig(serviceConfig, version)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred generating the cluster topology ingres from the service config '%s'", service.Name)
			}
			clusterTopologyIngress = append(clusterTopologyIngress, ingress)
		}

		if isIngres(serviceAnnotations) {
			// TODO: why this need to be a separated service?
			// Don't add ingress services to the list of resolved services
			continue
		}

		// 2- Service
		logrus.Infof("Processing service: %v", service.GetObjectMeta().GetName())
		clusterTopologyService := newClusterTopologyServiceFromServiceConfig(serviceConfig, version)

		// 3- Plugins
		// the servicePlugins list contains both stateful and external plugins and, externalServices is a list of Kardinal services that are also linked with a plugin inside the availablePlugins list
		servicePlugins, externalServices, newExternalServicesDependencies, err := newServicePluginsAndExternalServicesFromServiceConfig(serviceConfig, version, &clusterTopologyService, availablePlugins)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating new stateful availablePlugins and external services from service config '%s'", service.Name)
		}
		clusterTopologyService.StatefulPlugins = servicePlugins
		clusterTopologyServices = append(clusterTopologyServices, externalServices...)
		externalServicesDependencies = append(externalServicesDependencies, newExternalServicesDependencies...)

		// 4- Service dependencies (creates a list of services with dependencies)
		dependencies, ok := serviceAnnotations["kardinal.dev.service/dependencies"]
		if ok {
			newServiceWithDependenciesAnnotation := &serviceWithDependenciesAnnotation{&clusterTopologyService, dependencies}
			serviceWithDependencies = append(serviceWithDependencies, newServiceWithDependenciesAnnotation)
		}
		clusterTopologyServices = append(clusterTopologyServices, &clusterTopologyService)
	}

	// Third, set the service dependencies in the clusterTopologyService
	// a) iterate on the service with dependencies list
	for _, svcWithDependenciesAnnotation := range serviceWithDependencies {

		serviceAndPorts := strings.Split(svcWithDependenciesAnnotation.dependenciesAnnotation, ",")
		for _, serviceAndPort := range serviceAndPorts {
			serviceAndPortParts := strings.Split(serviceAndPort, ":")
			depService, depServicePort, err := getServiceAndPortFromClusterTopologyServices(serviceAndPortParts[0], serviceAndPortParts[1], clusterTopologyServices)
			if err != nil {
				return nil, nil, nil, stacktrace.Propagate(err, "An error occurred finding the service dependency for service %s and port %s", serviceAndPortParts[0], serviceAndPortParts[1])
			}

			serviceDependency := resolved.ServiceDependency{
				Service:          svcWithDependenciesAnnotation.service,
				DependsOnService: depService,
				DependencyPort:   depServicePort,
			}

			clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, serviceDependency)
		}
	}
	// b) add the external services dependencies
	clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, externalServicesDependencies...)

	return clusterTopologyIngress, clusterTopologyServices, clusterTopologyServiceDependencies, nil
}

func addAvailablePluginsFromServiceConfig(serviceConfig apitypes.ServiceConfig, availablePlugins map[string]*resolved.StatefulPlugin) (map[string]*resolved.StatefulPlugin, error) {
	service := serviceConfig.Service
	serviceAnnotations := service.GetObjectMeta().GetAnnotations()

	pluginAnnotation, ok := serviceAnnotations["kardinal.dev.service/plugin-definition"]
	if ok {
		var statefulPlugins []resolved.StatefulPlugin
		err := yaml.Unmarshal([]byte(pluginAnnotation), &statefulPlugins)
		if err != nil {
			return nil, stacktrace.Propagate(err, "an error occurred parsing the plugins for service %s", service.GetObjectMeta().GetName())
		}

		for index := range statefulPlugins {
			plugin := statefulPlugins[index]
			_, found := availablePlugins[plugin.ServiceName]
			if found {
				return nil, stacktrace.NewError("a plugin with service name '%s' already exists, the `plugin.servicename` value has to be unique", plugin.ServiceName)
			}
			availablePlugins[plugin.ServiceName] = &plugin
		}
	}

	return availablePlugins, nil
}

func newServicePluginsAndExternalServicesFromServiceConfig(
	serviceConfig apitypes.ServiceConfig,
	version string,
	clusterTopologyService *resolved.Service,
	availablePlugins map[string]*resolved.StatefulPlugin,
) (
	[]*resolved.StatefulPlugin,
	[]*resolved.Service,
	[]resolved.ServiceDependency,
	error,
) {
	servicePlugins := []*resolved.StatefulPlugin{}
	externalServices := []*resolved.Service{}
	externalServiceDependencies := []resolved.ServiceDependency{}

	service := serviceConfig.Service
	serviceAnnotations := service.GetObjectMeta().GetAnnotations()

	pluginsAnnotation, ok := serviceAnnotations["kardinal.dev.service/plugins"]
	if ok {
		pluginsServiceName := strings.Split(pluginsAnnotation, ",")
		for _, pluginSvcName := range pluginsServiceName {
			plugin, ok := availablePlugins[pluginSvcName]
			if !ok {
				return nil, nil, nil, stacktrace.NewError("expected to find plugin with service name %s but it is not available, make sure to add the resource for it in the manifest file", pluginSvcName)
			}
			servicePlugins = append(servicePlugins, plugin)
			// TODO: consider giving external service plugins their own type, instead of using StatefulPlugins
			// if this is an external service plugin, represent that service as a service in the cluster topology
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
		}
	}

	return servicePlugins, externalServices, externalServiceDependencies, nil
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

func newClusterTopologyIngresFromServiceConfig(serviceConfig apitypes.ServiceConfig, version string) (*resolved.Ingress, error) {
	service := serviceConfig.Service
	serviceAnnotations := service.GetObjectMeta().GetAnnotations()
	if !isIngres(serviceAnnotations) {
		return nil, stacktrace.NewError("Service %s is not an ingress service", service.GetObjectMeta().GetName())
	}
	ingress := &resolved.Ingress{
		ActiveFlowIDs: []string{version},
		IngressID:     service.ObjectMeta.Name,
		ServiceSpec:   &service.Spec,
	}
	host, ok := serviceAnnotations["kardinal.dev.service/host"]
	if ok {
		ingress.IngressRules = []*net.IngressRule{
			{
				Host: host,
			},
		}
	}
	return ingress, nil
}

func isIngres(serviceAnnotations map[string]string) bool {
	isIngress, ok := serviceAnnotations["kardinal.dev.service/ingress"]
	return ok && isIngress == "true"
}

func processIngressConfigs(ingressConfigs []apitypes.IngressConfig, version string) []*resolved.Ingress {
	clusterTopologyIngress := []*resolved.Ingress{}
	// First try to get it from the ingressConfigs
	for _, ingressConfig := range ingressConfigs {
		ingress := ingressConfig.Ingress
		ingressAnnotations := ingress.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := ingressAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			ingressObj := resolved.Ingress{
				ActiveFlowIDs: []string{version},
				IngressID:     ingress.ObjectMeta.Name,
				IngressSpec:   &ingress.Spec,
			}
			_, ok := ingressAnnotations["kardinal.dev.service/host"]
			if ok {
				logrus.Debugf("Found hostname Kardinal annotation on Ingress '%v' but using Ingress Rules provided by k8s Ingress object instead.", ingress.Name)
			}

			// A k8s ingress object should specify the Ingress rules so use those instead of creating one manually
			for _, ingressRule := range ingress.Spec.Rules {
				ingressObj.IngressRules = append(ingressObj.IngressRules, &ingressRule)
			}

			clusterTopologyIngress = append(clusterTopologyIngress, &ingressObj)
		}
	}
	return clusterTopologyIngress
}

func isServiceIngress(clusterTopology *resolved.ClusterTopology, service corev1.Service) bool {
	return lo.SomeBy(clusterTopology.Ingresses, func(item *resolved.Ingress) bool {
		return item.IngressID == service.GetObjectMeta().GetName()
	})
}
