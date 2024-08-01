package flow

import (
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"

	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func FindServiceByID(topology resolved.ClusterTopology, serviceID string) (*resolved.Service, int, bool) {
	var targetService *resolved.Service
	found := false
	pos := -1
	for ix, service := range topology.Services {
		if service.ServiceID == serviceID {
			targetService = service
			found = true
			pos = ix
			break
		}
	}
	return targetService, pos, found
}

func updateDependeciesInplace(serviceDependencies []resolved.ServiceDependency, targetService *resolved.Service, modifiedService *resolved.Service) {
	for ix, dependency := range serviceDependencies {
		if dependency.Service == targetService {
			dependency.Service = modifiedService
		}
		if dependency.DependsOnService == targetService {
			dependency.DependsOnService = modifiedService
		}
		serviceDependencies[ix] = dependency
	}
}

func CreateDevFlow(pluginRunner plugins.PluginRunner, flowID string, serviceID string, deploymentSpec appsv1.DeploymentSpec, baseTopology resolved.ClusterTopology) (*resolved.ClusterTopology, error) {
	// shallow copy the base topology
	topology := baseTopology

	// duplicate slices
	topology.Services = deepCopySlice(baseTopology.Services)
	topology.ServiceDependecies = deepCopySlice(baseTopology.ServiceDependecies)
	topology.Ingress = lo.Map(baseTopology.Ingress, func(item *resolved.Ingress, _ int) *resolved.Ingress {
		copiedIngress := resolved.Ingress{
			ActiveFlowIDs: []string{flowID},
			IngressID:     item.IngressID,
			IngressRules:  deepCopySlice(item.IngressRules),
			ServiceSpec:   item.ServiceSpec,
		}
		return &copiedIngress
	})

	// use deep copy the enforce immutability
	// deepCopy(baseTopology, topology)

	targetService, pos, found := FindServiceByID(topology, serviceID)
	if !found {
		return nil, fmt.Errorf("service with UUID %s not found", serviceID)
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

	modifiedTargetService := DeepCopyService(targetService)
	modifiedTargetService.DeploymentSpec = &deploymentSpec
	modifiedTargetService.Version = flowID

	topology.Services[pos] = modifiedTargetService
	updateDependeciesInplace(topology.ServiceDependecies, targetService, modifiedTargetService)

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
			for pluginIx, plugin := range modifiedService.StatefulPlugins {
				pluginId := fmt.Sprintf("%s-%s-%d", flowID, serviceID, pluginIx)
				logrus.Infof("Applying plugin %s for service %s with %s", plugin.Name, modifiedService.ServiceID, pluginId)
				spec, _, err := pluginRunner.CreateFlow(plugin.Name, *modifiedService.ServiceSpec, *resultSpec, pluginId, plugin.Args)
				if err != nil {
					return nil, fmt.Errorf("error creating flow for service %s: %v", modifiedService.ServiceID, err)
				}
				resultSpec = &spec
			}

			// Update service with final deployment spec
			modifiedService.DeploymentSpec = resultSpec

			topology.Services[statefulServiceIx] = modifiedService
			updateDependeciesInplace(topology.ServiceDependecies, statefulService, modifiedService)
		}
	}

	return &topology, nil
}

func topologyToGraph(topology resolved.ClusterTopology) graph.Graph[*resolved.Service, *resolved.Service] {
	serviceHash := func(service *resolved.Service) *resolved.Service {
		return service
	}
	graph := graph.New(serviceHash, graph.Directed())

	for _, service := range topology.Services {
		graph.AddVertex(service)
	}

	for _, dependency := range topology.ServiceDependecies {
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
