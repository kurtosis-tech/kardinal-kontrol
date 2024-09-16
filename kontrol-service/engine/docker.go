package engine

import (
	"fmt"
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

	clusterTopologyServices := []*resolved.Service{}
	clusterTopologyIngress := []*resolved.Ingress{}
	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	clusterTopology.Namespace = namespace

	alreadyFoundIngress := false
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
			alreadyFoundIngress = true
		}
	}

	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		deployment := serviceConfig.Deployment
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := serviceAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			if !alreadyFoundIngress {
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
			}
			// TODO: why this need to be a separated service?
			// Don't add ingress services to the list of resolved services
			continue
		}

		// Service
		logrus.Infof("Processing service: %v", service.GetObjectMeta().GetName())
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

		// Service plugin?
		sPlugins, ok := serviceAnnotations["kardinal.dev.service/plugins"]
		if ok {
			var statefulPlugins []resolved.StatefulPlugin
			err := yaml.Unmarshal([]byte(sPlugins), &statefulPlugins)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred parsing the plugins for service %s", service.GetObjectMeta().GetName())
			}
			serviceStatefulPlugins := make([]*resolved.StatefulPlugin, len(statefulPlugins))
			for index := range statefulPlugins {
				logrus.Infof("Voting App UI Plugin: %v", statefulPlugins[index].Name)
				// TODO: consider giving external service plugins their own type, instead of using StatefulPlugins
				// if this is an external service plugin, represent that service as a service in the cluster topology
				plugin := statefulPlugins[index]
				if plugin.Type == "external" {
					logrus.Infof("Adding external service to topology..")
					serviceName := plugin.ServiceName
					logrus.Infof("plugin service name: %v", plugin.ServiceName)
					if serviceName == "" {
						serviceName = fmt.Sprintf("%v:%v", clusterTopologyService.ServiceID, "external")
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

					clusterTopologyServices = append(clusterTopologyServices, &externalService)

					externalServiceDependency := resolved.ServiceDependency{
						Service:          &clusterTopologyService,
						DependsOnService: &externalService,
						DependencyPort:   nil,
					}
					clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, externalServiceDependency)
				}
				serviceStatefulPlugins[index] = &plugin
			}
			clusterTopologyService.StatefulPlugins = serviceStatefulPlugins
		}

		clusterTopologyServices = append(clusterTopologyServices, &clusterTopologyService)
	}

	if len(clusterTopologyIngress) == 0 {
		return nil, stacktrace.NewError("At least one service needs to be annotated as an ingress service")
	}
	clusterTopology.Ingresses = clusterTopologyIngress

	if len(clusterTopologyServices) == 0 {
		return nil, stacktrace.NewError("At least one service is required in addition to the ingress service(s)")
	}
	clusterTopology.Services = clusterTopologyServices

	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		if isServiceIngress(&clusterTopology, service) || alreadyFoundIngress {
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

	clusterTopology.ServiceDependencies = clusterTopologyServiceDependencies

	return &clusterTopology, nil
}

func isServiceIngress(clusterTopology *resolved.ClusterTopology, service v1.Service) bool {
	return lo.SomeBy(clusterTopology.Ingresses, func(item *resolved.Ingress) bool {
		return item.IngressID == service.GetObjectMeta().GetName()
	})
}
