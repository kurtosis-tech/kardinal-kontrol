package flow

import (
	"fmt"

	"kardinal.kontrol-service/constants"

	"github.com/kurtosis-tech/stacktrace"

	"github.com/dominikbraun/graph"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"

	v1 "k8s.io/api/apps/v1"
)

// CreateDevFlow creates a dev flow from the given topologies
// baseClusterTopologyMaybeWithTemplateOverrides - if a template is used then this is a modified version of the baseTopology
// we pass in the base topology anyway as we use services which remain in `prod` version from it
func CreateDevFlow(
	pluginRunner *plugins.PluginRunner,
	baseClusterTopologyMaybeWithTemplateOverrides resolved.ClusterTopology,
	baseTopology resolved.ClusterTopology,
	flowPatch flow_spec.FlowPatch,
) (*resolved.ClusterTopology, error) {
	flowID := flowPatch.FlowId

	// shallow copy the base topology
	topology := baseClusterTopologyMaybeWithTemplateOverrides

	// duplicate slices
	topology.FlowID = flowID
	topology.Services = deepCopySlice(baseClusterTopologyMaybeWithTemplateOverrides.Services)
	topology.ServiceDependencies = deepCopySlice(baseClusterTopologyMaybeWithTemplateOverrides.ServiceDependencies)
	topology.Ingress = &resolved.Ingress{
		ActiveFlowIDs: []string{flowID},
		Ingresses:     deepCopySlice(baseClusterTopologyMaybeWithTemplateOverrides.Ingress.Ingresses),
	}
	topology.GatewayAndRoutes = &resolved.GatewayAndRoutes{
		ActiveFlowIDs: []string{flowID},
		Gateways:      deepCopySlice(baseClusterTopologyMaybeWithTemplateOverrides.GatewayAndRoutes.Gateways),
		GatewayRoutes: deepCopySlice(baseClusterTopologyMaybeWithTemplateOverrides.GatewayAndRoutes.GatewayRoutes),
	}

	topologyRef := &topology

	clusterGraph := topologyToGraph(topologyRef)
	for _, servicePatch := range flowPatch.ServicePatches {
		serviceID := servicePatch.Service
		logrus.Infof("calculating new flow for service %s", serviceID)
		targetService, err := topologyRef.GetService(serviceID)
		if err != nil {
			return nil, err
		}
		_, err = applyPatch(pluginRunner, topologyRef, clusterGraph, flowID, targetService, servicePatch.DeploymentSpec)
		if err != nil {
			return nil, err
		}
	}

	// the baseline topology flow ID and flow version are equal to the namespace these three should use same value
	baselineFlowVersion := baseTopology.Namespace
	// Replace "baseline" version services with baseTopology versions
	for i, service := range topologyRef.Services {

		if service.Version == baselineFlowVersion {
			prodService, err := baseTopology.GetService(service.ServiceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get prod service %s: %v", service.ServiceID, err)
			}
			topologyRef.Services[i] = prodService
		}
	}

	// TODO(shared-annotation) - we could store "shared" versions somewhere so that the pointers are the same
	// if we do that then the render work around isn't necessary
	// perhaps top sort this; currently the following is possible
	// postgres is marked as shared, we mark its parent "cartservice" as shared
	// cartservice then happens in the loop and we try again (currently we don't as we check if version isn't shared)
	for _, service := range topology.Services {
		if service.IsShared && service.Version != baselineFlowVersion && service.Version != constants.SharedVersionVersionString {
			logrus.Infof("Marking service '%v' as shared, current version '%v'", service.ServiceID, service.Version)
			originalVersion := service.Version
			service.Version = constants.SharedVersionVersionString
			service.OriginalVersionIfShared = originalVersion

			if !service.IsHTTP() {
				logrus.Infof("Service '%v' isn't http; marking its parents as shared", service.ServiceID)
				err := markParentsAsShared(&topology, service)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Update service dependencies
	for i, dependency := range topologyRef.ServiceDependencies {
		if dependency.Service.Version == baselineFlowVersion {
			prodService, err := baseTopology.GetService(dependency.Service.ServiceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get prod service %s for dependency: %v", dependency.Service.ServiceID, err)
			}
			topologyRef.ServiceDependencies[i].Service = prodService
		}
		if dependency.DependsOnService.Version == baselineFlowVersion {
			prodDependsOnService, err := baseTopology.GetService(dependency.DependsOnService.ServiceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get prod service %s for dependsOn: %v", dependency.DependsOnService.ServiceID, err)
			}
			topologyRef.ServiceDependencies[i].DependsOnService = prodDependsOnService
		}
	}

	return topologyRef, nil
}

func markParentsAsShared(topology *resolved.ClusterTopology, service *resolved.Service) error {
	parents := topology.FindImmediateParents(service)
	for _, parent := range parents {
		if parent.Version == constants.SharedVersionVersionString {
			continue
		}
		logrus.Infof("Marking parent '%v' as shared, current verson '%s'", parent.ServiceID, parent.Version)
		parent.IsShared = true
		originalVersion := parent.Version
		parent.Version = constants.SharedVersionVersionString
		parent.OriginalVersionIfShared = originalVersion

		if !parent.IsHTTP() {
			err := markParentsAsShared(topology, parent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func applyPatch(
	pluginRunner *plugins.PluginRunner,
	topology *resolved.ClusterTopology,
	clusterGraph graph.Graph[resolved.ServiceHash, *resolved.Service],
	flowID string,
	targetService *resolved.Service,
	deploymentSpec *v1.DeploymentSpec,
) (*resolved.ClusterTopology, error) {
	// Find downstream stateful services
	statefulPaths := findAllDownstreamStatefulPaths(targetService, clusterGraph, topology)
	statefulServices := make([]*resolved.Service, 0)
	for _, path := range statefulPaths {
		statefulServiceHash, err := lo.Last(path)
		if statefulServiceHash == "" || err != nil {
			logrus.Infof("Error finding last stateful service hash in path %v: %v", path, err)
		}
		statefulService, err := clusterGraph.Vertex(statefulServiceHash)
		if err != nil {
			return nil, fmt.Errorf("an error occurred getting stateful service vertex from graph: %s", err)
		}
		statefulServices = append(statefulServices, statefulService)
	}
	statefulServices = lo.Uniq(statefulServices)

	externalPaths := findAllDownstreamExternalPaths(targetService, clusterGraph, topology)
	externalServices := make([]*resolved.Service, 0)
	alreadyHandledExternalServices := make([]string, 0)
	for _, path := range externalPaths {
		externalServiceHash, err := lo.Last(path)
		if externalServiceHash == "" || err != nil {
			logrus.Infof("Error finding last external service hash in path %v: %v", path, err)
		}
		externalService, err := clusterGraph.Vertex(externalServiceHash)
		if err != nil {
			return nil, fmt.Errorf("an error occurred getting external service vertex from graph: %s", err)
		}
		externalServices = append(externalServices, externalService)
	}
	externalServices = lo.Uniq(externalServices)

	// handle external service plugins on this service
	logrus.Infof("Checking if this service has any external services...")
	for pluginIdx, plugin := range targetService.StatefulPlugins {
		if plugin.Type == "external" {
			logrus.Infof("service %s contains an external dependency plugin: %v", targetService.ServiceID, plugin.Name)

			// find the existing external service and update it in the topology to get a new version
			externalService, err := topology.GetService(plugin.ServiceName)
			if err != nil {
				return nil, fmt.Errorf("external service specified by plugin '%v' was not found in base topology.", plugin.ServiceName)
			}

			err = applyExternalServicePlugin(pluginRunner, targetService, externalService, plugin, pluginIdx, flowID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating external servie plugin for external service '%v' depended on by '%v'", externalService.ServiceID, targetService.ServiceID)
			}

			err = topology.MoveServiceToVersion(externalService, flowID)
			if err != nil {
				return nil, err
			}
		}
	}

	modifiedTargetService := DeepCopyService(targetService)
	modifiedTargetService.DeploymentSpec = deploymentSpec
	modifiedTargetService.Version = flowID
	err := topology.UpdateWithService(modifiedTargetService)
	if err != nil {
		return nil, err
	}

	for serviceIdx, service := range topology.Services {
		if lo.Contains(statefulServices, service) {
			logrus.Debugf("applying stateful plugins on service: %s", service.ServiceID)
			// Don't modify the original service
			modifiedService := DeepCopyService(service)
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

			topology.Services[serviceIdx] = modifiedService
			topology.UpdateDependencies(service, modifiedService)

			// create versioned parents for non http stateful services
			// TODO - this should be done for all non http services and not just the stateful ones
			// 	every child should be copied; immediate parent duplicated
			// 	if children of non http services support http then our routing will have to be modified
			//  we should treat those http services as non http; a hack could be to remove the appProtocol HTTP marking
			if !modifiedService.IsHTTP() {
				logrus.Infof("Stateful service %s is non http; its parents shall be duplicated", modifiedService.ServiceID)
				parents := topology.FindImmediateParents(service)
				for _, parent := range parents {
					logrus.Infof("Duplicating parent %s", parent.ServiceID)
					err = topology.MoveServiceToVersion(parent, flowID)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// if the service is an external service of the target service, it was already handled above
		if lo.Contains(externalServices, service) && !lo.Contains(alreadyHandledExternalServices, service.ServiceID) {
			// 	assume there's only one parent service for now but eventually we'll likely need to account for multiple parents to external service
			parentServices := topology.FindImmediateParents(service)
			if len(parentServices) == 0 {
				return nil, stacktrace.NewError("Expected to find a parent service to the external service '%v' but did not find one. All external services should have a parent so this is a bug in Kardinal.", service.ServiceID)
			}
			parentService := parentServices[0]
			modifiedParentService := DeepCopyService(parentService)

			_, found := lo.Find(parentService.StatefulPlugins, func(plugin *resolved.StatefulPlugin) bool {
				return plugin.ServiceName == service.ServiceID
			})
			if !found {
				return nil, stacktrace.NewError("Did not find plugin on parent service '%v' for the corresponding external service '%v'.This is a bug in Kardinal.", parentService.ServiceID, service.ServiceID)
			}

			for pluginIdx, plugin := range parentService.StatefulPlugins {
				// assume there's only one plugin on the parent service for this external service
				if plugin.ServiceName == service.ServiceID {
					err := applyExternalServicePlugin(pluginRunner, parentService, service, plugin, pluginIdx, flowID)
					if err != nil {
						return nil, stacktrace.Propagate(err, "error creating flow for external service '%s'", service.ServiceID)
					}
				}
			}

			// add a flow version of the external service to the plugin
			err := topology.MoveServiceToVersion(service, flowID)
			if err != nil {
				return nil, err
			}

			// add the parent to the topology replacing the deployment spec with the spec returned from the flow
			err = topology.MoveServiceToVersion(modifiedParentService, flowID)
			if err != nil {
				return nil, err
			}
		}
	}

	return topology, nil
}

// TODO: have this handle stateful service plugins
func applyExternalServicePlugin(
	pluginRunner *plugins.PluginRunner,
	dependentService *resolved.Service,
	externalService *resolved.Service,
	externalServicePlugin *resolved.StatefulPlugin,
	pluginIdx int,
	flowId string,
) error {
	if externalServicePlugin.Type != "external" {
		return nil
	}

	logrus.Infof("Calling external service '%v' plugin with parent service '%v'...", externalService.ServiceID, dependentService.ServiceID)
	pluginId := plugins.GetPluginId(flowId, dependentService.ServiceID, pluginIdx)
	spec, _, err := pluginRunner.CreateFlow(externalServicePlugin.Name, *dependentService.ServiceSpec, *dependentService.DeploymentSpec, pluginId, externalServicePlugin.Args)
	if err != nil {
		return stacktrace.Propagate(err, "error creating flow for external service '%s'", externalService.ServiceID)
	}

	dependentService.DeploymentSpec = &spec
	return nil
}

func DeleteFlow(pluginRunner *plugins.PluginRunner, topology resolved.ClusterTopology, flowId string) error {
	for _, service := range topology.Services {
		// don't need to delete flow for services in the topology that aren't a part of this flow
		if service.Version != flowId {
			continue
		}
		err := DeleteDevFlow(pluginRunner, flowId, service)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting flow '%v' for service '%v'", flowId, service.ServiceID)
		}
	}
	return nil
}

func DeleteDevFlow(pluginRunner *plugins.PluginRunner, flowId string, service *resolved.Service) error {
	for pluginIdx, plugin := range service.StatefulPlugins {
		logrus.Infof("Attempting to delete flow for plugin '%v' on flow '%v'", plugin.Name, flowId)
		pluginId := plugins.GetPluginId(flowId, service.ServiceID, pluginIdx)
		err := pluginRunner.DeleteFlow(plugin.Name, pluginId)
		if err != nil {
			logrus.Errorf("Error deleting flow: %v.", err)
			return stacktrace.Propagate(err, "An error occurred while trying to call delete flow of plugin '%v' on service '%v' for flow '%v'", plugin.Name, service.ServiceID, flowId)
		}
	}
	return nil
}

func topologyToGraph(topology *resolved.ClusterTopology) graph.Graph[resolved.ServiceHash, *resolved.Service] {
	serviceHash := func(service *resolved.Service) resolved.ServiceHash {
		return service.Hash()
	}
	graph := graph.New(serviceHash, graph.Directed())

	for _, service := range topology.Services {
		graph.AddVertex(service)
	}

	for _, dependency := range topology.ServiceDependencies {
		graph.AddEdge(dependency.Service.Hash(), dependency.DependsOnService.Hash())
	}

	return graph
}

func findAllDownstreamStatefulPaths(targetService *resolved.Service, clusterGraph graph.Graph[resolved.ServiceHash, *resolved.Service], topology *resolved.ClusterTopology) [][]resolved.ServiceHash {
	allPaths := make([][]resolved.ServiceHash, 0)
	for _, service := range topology.Services {
		if service.IsStateful {
			paths, err := graph.AllPathsBetween(clusterGraph, targetService.Hash(), service.Hash())
			if err != nil {
				logrus.Infof("Error finding paths between %s and %s: %v", targetService.ServiceID, service.ServiceID, err)
				paths = [][]resolved.ServiceHash{}
			}
			allPaths = append(allPaths, paths...)
		}
	}
	return allPaths
}

func findAllDownstreamExternalPaths(targetService *resolved.Service, g graph.Graph[resolved.ServiceHash, *resolved.Service], topology *resolved.ClusterTopology) [][]resolved.ServiceHash {
	allPaths := make([][]resolved.ServiceHash, 0)
	for _, service := range topology.Services {
		if service.IsExternal {
			paths, err := graph.AllPathsBetween(g, targetService.Hash(), service.Hash())
			if err != nil {
				logrus.Infof("Error finding paths between %s and %s: %v", targetService.ServiceID, service.ServiceID, err)
				paths = [][]resolved.ServiceHash{}
			}
			allPaths = append(allPaths, paths...)
		}
	}
	return allPaths
}
