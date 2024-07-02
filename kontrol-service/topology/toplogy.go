package topology

import (
	"fmt"
	"strings"

	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/samber/lo"

	"kardinal.kontrol-service/types"
)

func ClusterTopology(cluster *types.Cluster) *apiTypes.ClusterTopology {
	nodes := lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) apiTypes.Node {
		label := fmt.Sprintf("%s (%s)", service.Name, service.Version)
		isRedis := strings.Contains(service.Name, "redis")

		nodeType := apiTypes.ServiceVersion
		if isRedis {
			nodeType = apiTypes.Redis
		}

		serviceName := service.Name
		return apiTypes.Node{
			Id:     fmt.Sprintf("%s (%s)", service.Name, service.Version),
			Label:  &label,
			Type:   nodeType,
			Parent: &serviceName,
		}
	})

	uniqServices := lo.UniqBy(cluster.Services, func(item *types.ServiceSpec) string { return item.Name })
	services := lo.Map(uniqServices, func(service *types.ServiceSpec, _ int) apiTypes.Node {
		label := service.Name
		return apiTypes.Node{
			Id:    service.Name,
			Label: &label,
			Type:  apiTypes.Service,
		}
	})

	gwLabel := "gateway"
	gateway := apiTypes.Node{
		Id:    "gateway",
		Label: &gwLabel,
		Type:  apiTypes.Gateway,
	}

	edges := []apiTypes.Edge{
		{
			Source: "gateway",
			Target: "voting-app-ui (prod)",
		},
		{
			Source: "voting-app-ui (prod)",
			Target: "redis-prod (prod)",
		},
		{
			Source: "gateway",
			Target: "voting-app-ui (dev)",
		},
		{
			Source: "voting-app-ui (dev)",
			Target: "kardinal-db-sidecar (dev)",
		},
		{
			Source: "kardinal-db-sidecar (dev)",
			Target: "redis-prod (prod)",
		},
	}

	allNodes := append(nodes, append(services, gateway)...)
	return &apiTypes.ClusterTopology{
		Nodes: allNodes,
		Edges: lo.Filter(edges, func(edge apiTypes.Edge, _ int) bool {
			_, hasSource := lo.Find(allNodes, func(node apiTypes.Node) bool {
				return edge.Source == node.Id
			})
			_, hasTarget := lo.Find(allNodes, func(node apiTypes.Node) bool {
				return edge.Target == node.Id
			})
			return hasSource && hasTarget
		}),
	}
}
