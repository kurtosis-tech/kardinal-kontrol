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
			var imageTag *string
			containers := service.WorkloadSpec.GetTemplateSpec().Containers
			if containers != nil && len(containers) > 0 {
				imageTag = &containers[0].Image
			}
			isBaseline := service.Version == clusterTopology.Namespace
			return apiTypes.NodeVersion{
				FlowId:     service.Version,
				ImageTag:   imageTag,
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

	if topology.GatewayAndRoutes != nil {
		for _, gw := range topology.GatewayAndRoutes.Gateways {
			gwLabel := gw.Name
			nodes = append(nodes, apiTypes.Node{
				Id:       gwLabel,
				Label:    gwLabel,
				Type:     apiTypes.Gateway,
				Versions: &[]apiTypes.NodeVersion{},
			})
		}
	}

	if topology.Ingress != nil {
		for _, ingress := range topology.Ingress.Ingresses {
			ingressLabel := ingress.Name
			nodes = append(nodes, apiTypes.Node{
				Id:       ingressLabel,
				Label:    ingressLabel,
				Type:     apiTypes.Gateway,
				Versions: &[]apiTypes.NodeVersion{},
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		if len(*nodes[i].Versions) == len(*nodes[j].Versions) {
			return nodes[i].Id < nodes[j].Id
		} else {
			return len(*nodes[i].Versions) > len(*nodes[j].Versions)
		}
	})
	return &apiTypes.ClusterTopology{
		Nodes: nodes,
		Edges: edges,
	}
}

func getIngressEdges(ingress *resolved.Ingress) []apiTypes.Edge {
	edges := []apiTypes.Edge{}

	if ingress == nil {
		return edges
	}

	for _, ingress := range ingress.Ingresses {
		gwLabel := ingress.Name
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					edges = append(edges, apiTypes.Edge{
						Source: gwLabel,
						Target: path.Backend.Service.Name,
					})
				}
			}
		}
	}

	return edges
}

func getGatewayEdges(gw *resolved.GatewayAndRoutes) []apiTypes.Edge {
	edges := []apiTypes.Edge{}

	if gw == nil {
		return edges
	}

	for _, route := range gw.GatewayRoutes {
		for _, ref := range route.ParentRefs {
			gwLabel := string(ref.Name)
			for _, rule := range route.Rules {
				for _, backRef := range rule.BackendRefs {
					edges = append(edges, apiTypes.Edge{
						Source: gwLabel,
						Target: string(backRef.Name),
					})
				}
			}
		}
	}

	return edges
}

func getClusterTopologyEdges(clusterTopology *resolved.ClusterTopology) []apiTypes.Edge {
	edges := []apiTypes.Edge{}

	ingressEdges := getIngressEdges(clusterTopology.Ingress)
	edges = append(edges, ingressEdges...)

	gatewayEdges := getGatewayEdges(clusterTopology.GatewayAndRoutes)
	edges = append(edges, gatewayEdges...)

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
