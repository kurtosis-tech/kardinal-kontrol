package flow

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"

	"github.com/dominikbraun/graph"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"

	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func CreateDevFlow(pluginRunner *plugins.PluginRunner, flowID string, serviceID string, deploymentSpec appsv1.DeploymentSpec, baseTopology resolved.ClusterTopology) (*resolved.ClusterTopology, error) {
	// shallow copy the base topology
	topology := baseTopology

	// duplicate slices
	topology.Services = deepCopySlice(baseTopology.Services)
	topology.ServiceDependencies = deepCopySlice(baseTopology.ServiceDependencies)
	topology.Ingresses = lo.Map(baseTopology.Ingresses, func(item *resolved.Ingress, _ int) *resolved.Ingress {
		copiedIngress := resolved.Ingress{
			ActiveFlowIDs: []string{flowID},
			IngressID:     item.IngressID,
			IngressRules:  deepCopySlice(item.IngressRules),
			ServiceSpec:   item.ServiceSpec,
		}
		return &copiedIngress
	})

	targetService, err := topology.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	logrus.Infof("calculating new flow for service %s", serviceID)

	g := topologyToGraph(topology)
	statefulPaths := findAllDownstreamStatefulPaths(targetService, g, topology)
	statefulServices := make([]*resolved.Service, 0)
	for _, path := range statefulPaths {
		statefulService, err := lo.Last(path)
		if statefulService == nil || err != nil {
			logrus.Infof("Error finding last service in path %v: %v", path, err)
		}
		statefulServices = append(statefulServices, statefulService)
	}
	statefulServices = lo.Uniq(statefulServices)

	// handle external service plugins
	logrus.Infof("Checking if this service has any external services...")
	for pluginIdx, plugin := range targetService.StatefulPlugins {
		if plugin.Type == "external" {
			logrus.Infof("This service contains an external dependency plugin: %v", plugin.Name)

			// find the existing external service and update it in the topology to get a new version
			externalService, err := topology.GetService(plugin.ServiceName)
			if err != nil {
				return nil, fmt.Errorf("external service specified by plugin '%v' was not found in base topology.", plugin.ServiceName)
			}

			resultSpec := DeepCopyDeploymentSpec(targetService.DeploymentSpec)

			logrus.Infof("Calling external service plugin...")
			pluginId := plugins.GetPluginId(flowID, targetService.ServiceID, pluginIdx)
			spec, _, err := pluginRunner.CreateFlow(plugin.Name, *targetService.ServiceSpec, *resultSpec, pluginId, plugin.Args)
			if err != nil {
				return nil, stacktrace.Propagate(err, "error creating flow for external service '%s'", externalService.ServiceID)
			}
			logrus.Infof("External service plugin successfully called.")

			deploymentSpec = spec
			topology.DuplicateAndUpdateService(externalService, flowID)
		}
	}

	modifiedTargetService := DeepCopyService(targetService)
	modifiedTargetService.DeploymentSpec = &deploymentSpec
	modifiedTargetService.Version = flowID
	err = topology.UpdateService(targetService.ServiceID, modifiedTargetService)
	if err != nil {
		return nil, err
	}
	topology.UpdateDependencies(targetService, modifiedTargetService)

	for statefulServiceIx, statefulService := range topology.Services {
		if lo.Contains(statefulServices, statefulService) {
			logrus.Debugf("applying stateful plugins on service: %s", statefulService.ServiceID)
			// Don't modify the original service
			modifiedService := DeepCopyService(statefulService)
			modifiedService.Version = flowID

			if !modifiedService.IsStateful {
				panic(fmt.Sprintf("Service %s is not stateful but is in stateful paths", modifiedService.ServiceID))
			}

			// Apply a chain of stateful plugins to the stateful service
			resultSpec := DeepCopyDeploymentSpec(modifiedService.DeploymentSpec)
			for pluginIdx, plugin := range modifiedService.StatefulPlugins {
				if plugin.Type == "external" {
					// we handle external plugins above
					// might need to handle this if stateful services can have external dependencies
					continue
				}
				logrus.Infof("Applying plugin %s for service %s with flow id %s", plugin.Name, modifiedService.ServiceID, flowID)
				pluginId := plugins.GetPluginId(flowID, modifiedService.ServiceID, pluginIdx)
				spec, _, err := pluginRunner.CreateFlow(plugin.Name, *modifiedService.ServiceSpec, *resultSpec, pluginId, plugin.Args)
				if err != nil {
					return nil, fmt.Errorf("error creating flow for service %s: %v", modifiedService.ServiceID, err)
				}
				resultSpec = &spec
			}

			// Update service with final deployment spec
			modifiedService.DeploymentSpec = resultSpec

			topology.Services[statefulServiceIx] = modifiedService
			topology.UpdateDependencies(statefulService, modifiedService)

			// create versioned parents for non http stateful services
			// TODO - this should be done for all non http services and not just the stateful ones
			// 	every child should be copied; immediate parent duplicated
			// 	if children of non http services support http then our routing will have to be modified
			//  we should treat those http services as non http; a hack could be to remove the appProtocol HTTP marking
			if !modifiedService.IsHTTP() {
				logrus.Infof("Stateful service %s is non http; its parents shall be duplicated", modifiedService.ServiceID)
				parents := topology.FindImmediateParents(statefulService)
				for _, parent := range parents {
					logrus.Infof("Duplicating parent %s", parent.ServiceID)
					topology.DuplicateAndUpdateService(parent, flowID)
				}
			}
		}
	}

	return &topology, nil
}

func DeleteFlow(pluginRunner *plugins.PluginRunner, topology resolved.ClusterTopology, flowId string) error {
	for _, service := range topology.Services {
		err := DeleteDevFlow(pluginRunner, flowId, service)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting flow '%v' for service '%v'", flowId, service.ServiceID)
		}
	}
	return nil
}

func DeleteDevFlow(pluginRunner *plugins.PluginRunner, flowId string, service *resolved.Service) error {
	// call delete flow on all the plugins in this flow
	for pluginIdx, plugin := range service.StatefulPlugins {
		logrus.Infof("Attempting to delete flow for plugin '%v' on flow '%v'", plugin.Name, flowId)
		pluginId := plugins.GetPluginId(flowId, service.ServiceID, pluginIdx)
		err := pluginRunner.DeleteFlow(plugin.Name, pluginId, map[string]string{})
		if err != nil {
			logrus.Errorf("Error deleting flow: %v.", err)
			return stacktrace.Propagate(err, "An error occurred while trying to call delete flow of plugin '%v' on service '%v' for flow '%v'", plugin.Name, service.ServiceID, flowId)
		}
	}
	return nil
}

func topologyToGraph(topology resolved.ClusterTopology) graph.Graph[*resolved.Service, *resolved.Service] {
	serviceHash := func(service *resolved.Service) *resolved.Service {
		return service
	}
	graph := graph.New(serviceHash, graph.Directed())

	for _, service := range topology.Services {
		graph.AddVertex(service)
	}

	for _, dependency := range topology.ServiceDependencies {
		graph.AddEdge(dependency.Service, dependency.DependsOnService)
	}

	return graph
}

func findAllDownstreamStatefulPaths(targetService *resolved.Service, g graph.Graph[*resolved.Service, *resolved.Service], topology resolved.ClusterTopology) [][]*resolved.Service {
	allPaths := make([][]*resolved.Service, 0)
	for _, service := range topology.Services {
		if service.IsStateful {
			paths, err := graph.AllPathsBetween(g, targetService, service)
			if err != nil {
				logrus.Infof("Error finding paths between %s and %s: %v", targetService.ServiceID, service.ServiceID, err)
				paths = [][]*resolved.Service{}
			}
			allPaths = append(allPaths, paths...)
		}
	}
	return allPaths
}
