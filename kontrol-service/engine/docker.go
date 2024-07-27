package engine

import (
	"strings"

	"github.com/kurtosis-tech/stacktrace"
	"gopkg.in/yaml.v2"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	v1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"

	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

func GenerateProdOnlyCluster(serviceConfigs []apitypes.ServiceConfig) (*resolved.ClusterTopology, error) {
	version := "prod"
	clusterTopology, err := generateClusterTopology(serviceConfigs, version)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occured generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func GenerateProdDevCluster(baseTopology *resolved.ClusterTopology, pluginRunner plugins.PluginRunner, flowID string, devServiceName string, devImage string) (*resolved.ClusterTopology, error) {
	devService, found := flow.FindServiceByID(*baseTopology, devServiceName)
	if !found {
		return nil, stacktrace.NewError("Service with UUID %s not found", devServiceName)
	}
	if devService.DeploymentSpec == nil {
		return nil, stacktrace.NewError("Service with UUID %s has no DeploymentSpec", devServiceName)
	}

	deploymentSpec := flow.DeepCopyDeploymentSpec(devService.DeploymentSpec)

	// TODO: find a better way to update deploymentSpec, this assumes there is only container in the pod
	deploymentSpec.Template.Spec.Containers[0].Image = devImage

	clusterTopology, err := flow.CreateDevFlow(pluginRunner, flowID, devServiceName, *deploymentSpec, *baseTopology)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occured generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func generateClusterTopology(serviceConfigs []apitypes.ServiceConfig, version string) (*resolved.ClusterTopology, error) {
	clusterTopology := resolved.ClusterTopology{}

	clusterTopologyServices := []*resolved.Service{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		deployment := serviceConfig.Deployment
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := serviceAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			clusterTopology.Ingress = resolved.Ingress{
				IngressID:   service.ObjectMeta.Name,
				ServiceSpec: &service.Spec,
			}
			host, ok := serviceAnnotations["kardinal.dev.service/host"]
			if ok {
				clusterTopology.Ingress.IngressRules = []*net.IngressRule{
					{
						Host: host,
					},
				}
			}
			continue
		}

		// Service
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

		// Service plugin?
		plugins, ok := serviceAnnotations["kardinal.dev.service/plugins"]
		if ok {
			var statefulPlugins []resolved.StatefulPlugin
			err := yaml.Unmarshal([]byte(plugins), &statefulPlugins)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred parsing the plugins for service %s", service.GetObjectMeta().GetName())
			}
			serviceStatefulPlugins := make([]*resolved.StatefulPlugin, len(statefulPlugins))
			for index := range statefulPlugins {
				serviceStatefulPlugins[index] = &statefulPlugins[index]
			}
			clusterTopologyService.StatefulPlugins = serviceStatefulPlugins
		}

		clusterTopologyServices = append(clusterTopologyServices, &clusterTopologyService)
	}

	clusterTopology.Services = clusterTopologyServices

	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		if service.GetObjectMeta().GetName() == clusterTopology.Ingress.IngressID {
			continue
		}

		clusterTopologyService, err := clusterTopology.GetService(service.GetObjectMeta().GetName())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred finding service %s in the list of services", service.GetObjectMeta().GetName())
		}

		// Service dependencies?
		deps, ok := serviceAnnotations["kardinal.dev.service/dependencies"]
		if ok {
			serviceAndPorts := strings.Split(deps, ",")
			for _, serviceAndPort := range serviceAndPorts {
				serviceAndPortParts := strings.Split(serviceAndPort, ":")
				depService, depServicePort, err := clusterTopology.GetServiceAndPort(serviceAndPortParts[0], serviceAndPortParts[1])
				if err != nil {
					return nil, stacktrace.Propagate(err, "An error occurred finding the service dependency for service %s and port %s", serviceAndPortParts[0], serviceAndPortParts[1])
				}

				serviceDependency := resolved.ServiceDependency{
					Service:          clusterTopologyService,
					DependsOnService: depService,
					DependencyPort:   depServicePort,
				}

				clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, serviceDependency)
			}
		}
	}

	clusterTopology.ServiceDependecies = clusterTopologyServiceDependencies

	return &clusterTopology, nil
}

func getServiceAndPort(serviceName string, servicePortName string, services []*resolved.Service) (*resolved.Service, *v1.ServicePort, error) {
	for _, service := range services {
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

func getService(serviceName string, services []*resolved.Service) (*resolved.Service, error) {
	for _, service := range services {
		if service.ServiceID == serviceName {
			return service, nil
		}
	}

	return nil, stacktrace.NewError("Service %s not found in the list of services", serviceName)
}
