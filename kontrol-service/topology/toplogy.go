package topology

import (
	"fmt"

	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/samber/lo"

	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func ClusterTopology(clusterTopology *resolved.ClusterTopology) *apiTypes.ClusterTopology {
	nodes := lo.Map(clusterTopology.Services, func(service resolved.Service, _ int) apiTypes.Node {
		label := fmt.Sprintf("%s (%s)", service.ServiceID, service.Version)
		nodeType := apiTypes.ServiceVersion
		serviceName := service.ServiceID
		return apiTypes.Node{
			Id:     label,
			Label:  &label,
			Type:   nodeType,
			Parent: &serviceName,
		}
	})

	uniqServices := lo.UniqBy(clusterTopology.Services, func(item resolved.Service) string { return item.ServiceID })
	services := lo.Map(uniqServices, func(service resolved.Service, _ int) apiTypes.Node {
		label := service.ServiceID
		return apiTypes.Node{
			Id:    label,
			Label: &label,
			Type:  apiTypes.Service,
		}
	})

	gwLabel := clusterTopology.Ingress.IngressID
	gateway := apiTypes.Node{
		Id:    gwLabel,
		Label: &gwLabel,
		Type:  apiTypes.Gateway,
	}

	edges := []apiTypes.Edge{}
	ingressAppName := clusterTopology.Ingress.GetSelectorAppName()
	if ingressAppName != nil {
		ingressTargetService, _ := clusterTopology.GetService(*ingressAppName)
		if ingressTargetService != nil {
			edges = append(edges, apiTypes.Edge{
				Source: gwLabel,
				Target: fmt.Sprintf("%s (%s)", ingressTargetService.ServiceID, ingressTargetService.Version),
			})
		}
	}

	for _, serviceDependency := range clusterTopology.ServiceDependecies {
		edges = append(edges, apiTypes.Edge{
			Source: fmt.Sprintf("%s (%s)", serviceDependency.Service.ServiceID, serviceDependency.Service.Version),
			Target: fmt.Sprintf("%s (%s)", serviceDependency.DependsOnService.ServiceID, serviceDependency.DependsOnService.Version),
		})
	}

	allNodes := append(nodes, append(services, gateway)...)
	return &apiTypes.ClusterTopology{
		Nodes: allNodes,
		Edges: edges,
	}
}
