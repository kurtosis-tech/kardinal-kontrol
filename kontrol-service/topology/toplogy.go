package topology

import (
	"fmt"

	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/samber/lo"

	"kardinal.kontrol-service/types"
)

func ClusterTopology(cluster *types.Cluster) *apiTypes.ClusterTopology {
	nodes := lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) apiTypes.Node {
		label := fmt.Sprintf("%s (%s)", service.Name, service.Version)
		return apiTypes.Node{
			Id:    fmt.Sprintf("%s (%s)", service.Name, service.Version),
			Label: &label,
		}
	})

	gwLabel := "gateway"
	gateway := apiTypes.Node{
		Id:    "gateway",
		Label: &gwLabel,
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

	return &apiTypes.ClusterTopology{
		Nodes: append(nodes, gateway),
		Edges: lo.Filter(edges, func(edge apiTypes.Edge, _ int) bool {
			_, hasSource := lo.Find(cluster.Services, func(service *types.ServiceSpec) bool {
				return edge.Source == fmt.Sprintf("%s (%s)", service.Name, service.Version)
			})
			_, hasTarget := lo.Find(cluster.Services, func(service *types.ServiceSpec) bool {
				return edge.Target == fmt.Sprintf("%s (%s)", service.Name, service.Version)
			})
			return hasSource && hasTarget
		}),
	}
}
