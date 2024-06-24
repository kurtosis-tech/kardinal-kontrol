package topology

import (
	compose "github.com/compose-spec/compose-go/types"
	"github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
)

func ComposeToTopology(services *[]compose.ServiceConfig) *types.Topology {
	var nodes []types.Node

	for _, service := range *services {
		serviceName := service.ContainerName
		serviceVersion := service.Image
		serviceID := service.Name
		talksTo := service.GetDependencies()

		node := types.Node{
			Id:             &serviceID,
			ServiceName:    &serviceName,
			ServiceVersion: &serviceVersion,
			TalksTo:        &talksTo,
		}

		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		emptyTopology := types.Topology{}
		return &emptyTopology
	}

	topology := types.Topology{
		Graph: &types.Graph{
			Nodes: &nodes,
		},
	}

	return &topology
}
