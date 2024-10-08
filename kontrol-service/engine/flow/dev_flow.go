package flow

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	kardinal "kardinal.kontrol-service/types/kardinal"

	"kardinal.kontrol-service/constants"

	"github.com/kurtosis-tech/stacktrace"

	"github.com/dominikbraun/graph"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
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

	if err := applyPatch(pluginRunner, topologyRef, flowID, flowPatch.ServicePatches); err != nil {
		return nil, err
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
	topologyRef *resolved.ClusterTopology,
	flowID string,
	servicePatches []flow_spec.ServicePatch,
) error {

	// TODO could create a custom type for it with a Add method and a Get Method, in order to centralize the addition if someone else want o use it later in another part in the code
	pluginServices := map[string][]string{}
	pluginServicesMap := map[string]*resolved.StatefulPlugin{}

	clusterGraph := topologyToGraph(topologyRef)

	for _, servicePatch := range servicePatches {
		serviceID := servicePatch.Service
		logrus.Infof("calculating new flow for service %s", serviceID)
		targetService, err := topologyRef.GetService(serviceID)
		if err != nil {
			return err
		}
		if targetService.ServiceSpec == nil {
			return stacktrace.NewError("service '%v' does not have a service spec", targetService.ServiceID)
		}

		if targetService.WorkloadSpec == nil {
			return stacktrace.NewError("service '%v' does not have a workload spec", targetService.ServiceID)
		}

		// Find downstream stateful services
		statefulPaths := findAllDownstreamStatefulPaths(targetService, clusterGraph, topologyRef)
		statefulServices := make([]*resolved.Service, 0)
		for _, path := range statefulPaths {
			statefulServiceHash, err := lo.Last(path)
			if statefulServiceHash == "" || err != nil {
				logrus.Infof("Error finding last stateful service hash in path %v: %v", path, err)
			}
			statefulService, err := clusterGraph.Vertex(statefulServiceHash)
			if err != nil {
				return fmt.Errorf("an error occurred getting stateful service vertex from graph: %s", err)
			}
			statefulServices = append(statefulServices, statefulService)
		}
		statefulServices = lo.Uniq(statefulServices)

		externalPaths := findAllDownstreamExternalPaths(targetService, clusterGraph, topologyRef)
		externalServices := make([]*resolved.Service, 0)
		alreadyHandledExternalServices := make([]string, 0)
		for _, path := range externalPaths {
			externalServiceHash, err := lo.Last(path)
			if externalServiceHash == "" || err != nil {
				logrus.Infof("Error finding last external service hash in path %v: %v", path, err)
			}
			externalService, err := clusterGraph.Vertex(externalServiceHash)
			if err != nil {
				return fmt.Errorf("an error occurred getting external service vertex from graph: %s", err)
			}
			externalServices = append(externalServices, externalService)
		}
		externalServices = lo.Uniq(externalServices)

		// SECTION 1 - Create external plugins and move the external K8s Service to a new version with the FlowID
		// handle external service plugins on this service
		logrus.Infof("Checking if this service has any external services...")
		for _, plugin := range targetService.StatefulPlugins {

			alreadyServicesWithPlugin, ok := pluginServices[plugin.ServiceName]
			if ok {
				pluginServices[plugin.ServiceName] = append(alreadyServicesWithPlugin, targetService.ServiceID)
			} else {
				pluginServices[plugin.ServiceName] = []string{targetService.ServiceID}
			}
			pluginServicesMap[plugin.ServiceName] = plugin

			// Edit the external service k8s.Service resource setting it the flow ID
			if plugin.Type == "external" {
				logrus.Infof("service %s contains an external dependency plugin: %v", targetService.ServiceID, plugin.Name)

				// find the existing external service and update it in the topology to get a new version
				externalService, err := topologyRef.GetService(plugin.ServiceName)
				if err != nil {
					return fmt.Errorf("external service specified by plugin '%v' was not found in base topology", plugin.ServiceName)
				}

				err = topologyRef.MoveServiceToVersion(externalService, flowID)
				if err != nil {
					return err
				}
			}
		}

		// SECTION 2 - Target service updates with new modifications
		modifiedTargetService := DeepCopyService(targetService)
		modifiedTargetService.WorkloadSpec = servicePatch.WorkloadSpec
		err = topologyRef.MoveServiceToVersion(modifiedTargetService, flowID)
		if err != nil {
			return err
		}

		// SECTION 3 - handle stateful services
		for serviceIdx, service := range topologyRef.Services {
			if lo.Contains(statefulServices, service) {
				logrus.Debugf("applying stateful plugins on service: %s", service.ServiceID)
				// Don't modify the original service
				modifiedService := DeepCopyService(service)
				modifiedService.Version = flowID

				if !modifiedService.IsStateful {
					return fmt.Errorf("service %s is not stateful but is in stateful paths", modifiedService.ServiceID)
				}

				// Apply a chain of stateful plugins to the stateful service
				resultSpec := modifiedService.WorkloadSpec.DeepCopy()

				for _, plugin := range modifiedService.StatefulPlugins {
					if plugin.Type == "external" {
						//we handle external plugins above
						//might need to handle this if stateful services can have external dependencies
						continue
					}

					alreadyServicesWithPlugin, ok := pluginServices[plugin.ServiceName]
					if ok {
						pluginServices[plugin.ServiceName] = append(alreadyServicesWithPlugin, modifiedService.ServiceID)
					} else {
						pluginServices[plugin.ServiceName] = []string{modifiedService.ServiceID}
					}
					pluginServicesMap[plugin.ServiceName] = plugin
				}

				// Update service with final deployment spec
				modifiedService.WorkloadSpec = resultSpec

				topologyRef.Services[serviceIdx] = modifiedService
				topologyRef.UpdateDependencies(service, modifiedService)

				// create versioned parents for non http stateful services
				// TODO - this should be done for all non http services and not just the stateful ones
				// 	every child should be copied; immediate parent duplicated
				// 	if children of non http services support http then our routing will have to be modified
				//  we should treat those http services as non http; a hack could be to remove the appProtocol HTTP marking
				if !modifiedService.IsHTTP() {
					logrus.Infof("Stateful service %s is non http; its parents shall be duplicated", modifiedService.ServiceID)
					parents := topologyRef.FindImmediateParents(service)
					for _, parent := range parents {
						logrus.Infof("Duplicating parent %s", parent.ServiceID)
						err = topologyRef.MoveServiceToVersion(parent, flowID)
						if err != nil {
							return err
						}
					}
				}
			}

			// SECTION 4 - handle external services that are not target service dependencies
			// if the service is an external service of the target service, it was already handled above
			if lo.Contains(externalServices, service) && !lo.Contains(alreadyHandledExternalServices, service.ServiceID) {
				// 	assume there's only one parent service for now but eventually we'll likely need to account for multiple parents to external service
				parentServices := topologyRef.FindImmediateParents(service)
				if len(parentServices) == 0 {
					return stacktrace.NewError("Expected to find a parent service to the external service '%v' but did not find one. All external services should have a parent so this is a bug in Kardinal.", service.ServiceID)
				}
				parentService := parentServices[0]
				modifiedParentService := DeepCopyService(parentService)

				_, found := lo.Find(parentService.StatefulPlugins, func(plugin *resolved.StatefulPlugin) bool {
					return plugin.ServiceName == service.ServiceID
				})
				if !found {
					return stacktrace.NewError("Did not find plugin on parent service '%v' for the corresponding external service '%v'.This is a bug in Kardinal.", parentService.ServiceID, service.ServiceID)
				}

				for _, plugin := range parentService.StatefulPlugins {
					// assume there's only one plugin on the parent service for this external service
					if parentService.ServiceSpec == nil {
						return stacktrace.NewError("parent service '%v' does not have a service spec", parentService.ServiceID)
					}

					if parentService.WorkloadSpec == nil {
						return stacktrace.NewError("parent service '%v' does not have a workload spec", targetService.ServiceID)
					}

					alreadyServicesWithPlugin, ok := pluginServices[plugin.ServiceName]
					if ok {
						pluginServices[plugin.ServiceName] = append(alreadyServicesWithPlugin, parentService.ServiceID)
					} else {
						pluginServices[plugin.ServiceName] = []string{parentService.ServiceID}
					}
					pluginServicesMap[plugin.ServiceName] = plugin
				}

				// add a flow version of the external service to the plugin
				err := topologyRef.MoveServiceToVersion(service, flowID)
				if err != nil {
					return err
				}

				// add the parent to the topology replacing the deployment spec with the spec returned from the flow
				err = topologyRef.MoveServiceToVersion(modifiedParentService, flowID)
				if err != nil {
					return err
				}
			}
		}
	}

	// SECTION 5 - Execute plugins and update the services deployment specs with the plugin's modifications
	for pluginServiceName, serviceIds := range pluginServices {
		var servicesServiceSpecs []corev1.ServiceSpec
		var servicesWorkloadSpecs []*kardinal.WorkloadSpec
		var servicesToUpdate []*resolved.Service

		plugin, ok := pluginServicesMap[pluginServiceName]
		if !ok {
			return stacktrace.NewError("expected to find plugin with service name '%s' in the plugins service map, this is a bug in Kardinal", pluginServiceName)
		}

		if len(serviceIds) == 0 {
			return stacktrace.NewError("expected to find at least one service depending on plugin '%s' but none was found, please review your manifest file", plugin.ServiceName)
		}

		for _, serviceId := range serviceIds {
			service, err := topologyRef.GetService(serviceId)
			if err != nil {
				return stacktrace.Propagate(err, "an error occurred getting service '%s' from topology", serviceId)
			}
			servicesServiceSpecs = append(servicesServiceSpecs, *service.ServiceSpec)
			servicesWorkloadSpecs = append(servicesWorkloadSpecs, service.WorkloadSpec)
			servicesToUpdate = append(servicesToUpdate, service)
		}

		pluginId := plugins.GetPluginId(plugin.ServiceName, flowID)
		logrus.Infof("Calling plugin '%v'...", pluginId)

		servicesModifiedWorkloadSpecs, _, err := pluginRunner.CreateFlow(plugin.Name, servicesServiceSpecs, servicesWorkloadSpecs, pluginId, plugin.Args)
		if err != nil {
			return stacktrace.Propagate(err, "error when creating plugin flow for plugin '%s'", pluginId)
		}

		if len(servicesToUpdate) != len(servicesModifiedWorkloadSpecs) {
			return fmt.Errorf("an error occurred executing plugin '%s', the number of workload specs returned by the plugin.CreateFlow function are not equal to the number of service depending on it, please check the plugin code or report a bug in the Kardinal repository", plugin.ServiceName)
		}

		// updating the service.workload_spec after the plugin execution
		for serviceIndex := range serviceIds {
			service := servicesToUpdate[serviceIndex]
			modifiedWorkloadSpec := servicesModifiedWorkloadSpecs[serviceIndex]
			service.WorkloadSpec = modifiedWorkloadSpec
			if err = topologyRef.MoveServiceToVersion(service, flowID); err != nil {
				return fmt.Errorf("an error occurred updating service '%s'", service.ServiceID)
			}
		}
	}

	return nil
}

func DeleteFlow(pluginRunner *plugins.PluginRunner, topology resolved.ClusterTopology, flowId string) error {
	pluginsToDeleteFromThisFlow := map[string]string{}

	for _, service := range topology.Services {
		// don't need to delete flow for services in the topology that aren't a part of this flow
		if service.Version != flowId {
			continue
		}
		for _, plugin := range service.StatefulPlugins {
			pluginId := plugins.GetPluginId(plugin.ServiceName, flowId)
			pluginsToDeleteFromThisFlow[pluginId] = plugin.Name
		}
	}

	for pluginId, pluginName := range pluginsToDeleteFromThisFlow {
		err := pluginRunner.DeleteFlow(pluginName, pluginId)
		if err != nil {
			logrus.Errorf("Error deleting flow: %v.", err)
			return stacktrace.Propagate(err, "An error occurred while trying to call delete flow of plugin '%v' for flow '%v'", pluginName, flowId)
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
