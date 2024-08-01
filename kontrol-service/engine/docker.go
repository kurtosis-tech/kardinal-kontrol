package engine

import (
	"strings"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
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

func GenerateProdOnlyCluster(flowID string, serviceConfigs []apitypes.ServiceConfig) (*resolved.ClusterTopology, error) {
	clusterTopology, err := generateClusterTopology(serviceConfigs, flowID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occured generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func GenerateProdDevCluster(baseTopology *resolved.ClusterTopology, pluginRunner plugins.PluginRunner, flowID string, devServiceName string, devImage string) (*resolved.ClusterTopology, error) {
	devService, _, found := flow.FindServiceByID(*baseTopology, devServiceName)
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
	clusterTopologyIngress := []*resolved.Ingress{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		deployment := serviceConfig.Deployment
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := serviceAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			ingress := resolved.Ingress{
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
			clusterTopologyIngress = append(clusterTopologyIngress, &ingress)
			// TODO: why this need to be a separeted service?
			// Don't add ingress services to the list of resolved services
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
	clusterTopology.Ingress = clusterTopologyIngress

	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		if isServiceIngress(&clusterTopology, service) {
			logrus.Infof("Service %s is an ingress service, skipping dependency resolution", service.GetObjectMeta().GetName())
			continue
		}

		clusterTopologyService, err := clusterTopology.GetService(service.GetObjectMeta().GetName())
		if err != nil {
			logrus.Fatalf("An error occurred finding service %s in the list of services", service.GetObjectMeta().GetName())
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

func isServiceIngress(clusterTopology *resolved.ClusterTopology, service v1.Service) bool {
	return lo.SomeBy(clusterTopology.Ingress, func(item *resolved.Ingress) bool {
		return item.IngressID == service.GetObjectMeta().GetName()
	})
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
