package topology

import (
	"sort"

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

	groupedServices := lo.GroupBy(topology.Services, func(item *resolved.Service) string { return item.ServiceID })
	nodes := lo.MapToSlice(groupedServices, func(key string, services []*resolved.Service) apiTypes.Node {
		nodeType := apiTypes.Service
		if services[0].IsExternal {
			nodeType = apiTypes.External
		}
		label := key
		versions := lo.Map(services, func(service *resolved.Service, _ int) apiTypes.NodeVersion {
			isBaseline := service.Version == clusterTopology.Namespace
			return apiTypes.NodeVersion{
				FlowId:     service.Version,
				ImageTag:   service.DeploymentSpec.Template.Spec.Containers[0].Image,
				IsBaseline: isBaseline,
			}
		})
		sort.Slice(versions, func(i, j int) bool {
			if versions[i].IsBaseline && !versions[j].IsBaseline {
				return true
			} else if !versions[i].IsBaseline && versions[j].IsBaseline {
				return false
			} else {
				return versions[i].FlowId < versions[j].FlowId
			}
		})
		return apiTypes.Node{
			Type:     nodeType,
			Id:       label,
			Label:    label,
			Versions: &versions,
		}
	})

	gateways := lo.Map(clusterTopology.Ingresses, func(ingress *resolved.Ingress, _ int) apiTypes.Node {
		gwLabel := ingress.IngressID
		return apiTypes.Node{
			Id:       gwLabel,
			Label:    gwLabel,
			Type:     apiTypes.Gateway,
			Versions: &[]apiTypes.NodeVersion{},
		}
	})

	allNodes := append(nodes, gateways...)
	sort.Slice(allNodes, func(i, j int) bool {
		if len(*allNodes[i].Versions) == len(*allNodes[j].Versions) {
			return allNodes[i].Id < allNodes[j].Id
		} else {
			return len(*allNodes[i].Versions) > len(*allNodes[j].Versions)
		}
	})
	return &apiTypes.ClusterTopology{
		Nodes: allNodes,
		Edges: edges,
	}
}

func getClusterTopologyEdges(clusterTopology *resolved.ClusterTopology) []apiTypes.Edge {
	edges := []apiTypes.Edge{}

	for _, ingress := range clusterTopology.Ingresses {
		gwLabel := ingress.IngressID

		for _, targetService := range ingress.GetTargetServices() {
			edges = append(edges, apiTypes.Edge{
				Source: gwLabel,
				Target: targetService,
			})
		}
	}

	for _, serviceDependency := range clusterTopology.ServiceDependencies {
		edges = append(edges, apiTypes.Edge{
			Source: serviceDependency.Service.ServiceID,
			Target: serviceDependency.DependsOnService.ServiceID,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source == edges[j].Source {
			return edges[i].Target < edges[j].Target
		} else {
			return edges[i].Source < edges[j].Source
		}
	})
	return edges
}
