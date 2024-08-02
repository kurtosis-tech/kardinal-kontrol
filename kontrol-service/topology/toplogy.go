package topology

import (
	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/samber/lo"

	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// Aggregate mode: clusterTopology is the base topology and flowsClusterTopology contains the topologies for all the flows.
// Single flow mode: clusterToppology is the flow topology and flowsClusterTopology is set to nil.
func ClusterTopology(clusterTopology *resolved.ClusterTopology, flowsClusterTopology *[]resolved.ClusterTopology) *apiTypes.ClusterTopology {

	var topology *resolved.ClusterTopology
	if flowsClusterTopology != nil {
		topology = flow.MergeClusterTopologies(*clusterTopology, *flowsClusterTopology)
	} else {
		topology = clusterTopology
	}

	edges := getClusterTopologyEdges(clusterTopology)

	servicesToVersions := map[string][]string{}
	groupedServices := lo.GroupBy(topology.Services, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		servicesToVersions[serviceID] = lo.Map(services, func(item *resolved.Service, _ int) string { return item.Version })
	}

	nodes := lo.MapToSlice(servicesToVersions, func(key string, value []string) apiTypes.Node {
		nodeType := apiTypes.Service
		label := key
		return apiTypes.Node{
			Type:     nodeType,
			Id:       label,
			Label:    &label,
			Versions: &value,
		}
	})

	gateways := lo.Map(clusterTopology.Ingress, func(ingress *resolved.Ingress, _ int) apiTypes.Node {
		gwLabel := ingress.IngressID
		return apiTypes.Node{
			Id:    gwLabel,
			Label: &gwLabel,
			Type:  apiTypes.Gateway,
		}
	})

	allNodes := append(nodes, gateways...)
	return &apiTypes.ClusterTopology{
		Nodes: allNodes,
		Edges: edges,
	}
}

func getClusterTopologyEdges(clusterTopology *resolved.ClusterTopology) []apiTypes.Edge {
	edges := []apiTypes.Edge{}

	for _, ingress := range clusterTopology.Ingress {
		gwLabel := ingress.IngressID

		ingressAppName := ingress.GetSelectorAppName()
		if ingressAppName != nil {
			ingressTargetService, _ := clusterTopology.GetService(*ingressAppName)
			if ingressTargetService != nil {
				edges = append(edges, apiTypes.Edge{
					Source: gwLabel,
					Target: ingressTargetService.ServiceID,
				})
			}
		}
	}

	for _, serviceDependency := range clusterTopology.ServiceDependecies {
		edges = append(edges, apiTypes.Edge{
			Source: serviceDependency.Service.ServiceID,
			Target: serviceDependency.DependsOnService.ServiceID,
		})
	}

	return edges
}
