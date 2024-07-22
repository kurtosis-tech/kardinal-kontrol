package flow

import (
	"github.com/dominikbraun/graph"
	appsv1 "k8s.io/api/apps/v1"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func createDevFlow(serviceID string, deploymentSpec appsv1.DeploymentSpec, topology resolved.ClusterTopology) types.ClusterResources {
	// TODO: Implement
	return types.ClusterResources{}
}

func topologyToGraph(topology resolved.ClusterTopology) graph.Graph[string, resolved.Service] {
	serviceHash := func(service resolved.Service) string {
		return service.ServiceID
	}
	graph := graph.New(serviceHash)

	for _, service := range topology.Services {
		graph.AddVertex(service)
	}

	for _, dependency := range topology.ServiceDependecies {
		graph.AddEdge(dependency.Service.ServiceID, dependency.DependsOnService.ServiceID)
	}

	return graph
}
