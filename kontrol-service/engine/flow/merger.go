package flow

import (
	"encoding/json"

	"github.com/samber/lo"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func MergeClusterTopologies(baseTopology resolved.ClusterTopology, clusterTopologies []resolved.ClusterTopology) (mergedTopology *resolved.ClusterTopology) {
	mergedTopology = &resolved.ClusterTopology{
		FlowID:              "all",
		Services:            deepCopySlice(baseTopology.Services),
		ServiceDependencies: deepCopySlice(baseTopology.ServiceDependencies),
		Ingress:             DeepCopyIngress(baseTopology.Ingress),
		GatewayAndRoutes:    DeepCopyGatewayAndRoutes(baseTopology.GatewayAndRoutes),
		Namespace:           baseTopology.Namespace,
	}
	for _, topology := range clusterTopologies {
		mergedTopology.Services = append(mergedTopology.Services, topology.Services...)
		mergedTopology.ServiceDependencies = append(mergedTopology.ServiceDependencies, topology.ServiceDependencies...)
		mergedTopology.Ingress.ActiveFlowIDs = append(mergedTopology.Ingress.ActiveFlowIDs, topology.Ingress.ActiveFlowIDs...)
		mergedTopology.GatewayAndRoutes.ActiveFlowIDs = append(mergedTopology.GatewayAndRoutes.ActiveFlowIDs, topology.GatewayAndRoutes.ActiveFlowIDs...)
	}
	mergedTopology.Ingress.ActiveFlowIDs = lo.Uniq(mergedTopology.Ingress.ActiveFlowIDs)
	mergedTopology.GatewayAndRoutes.ActiveFlowIDs = lo.Uniq(mergedTopology.GatewayAndRoutes.ActiveFlowIDs)

	// TODO improve the filtering method, we could implement the `Service.Equal` method to compare and filter the services
	// TODO and inside this method we could use the k8s service marshall method (https://pkg.go.dev/k8s.io/api/core/v1#Service.Marsha) and also the same for other k8s fields
	// TODO it should be faster
	mergedTopology.Services = lo.UniqBy(mergedTopology.Services, MustGetMarshalledKey[*resolved.Service])
	mergedTopology.ServiceDependencies = lo.UniqBy(mergedTopology.ServiceDependencies, MustGetMarshalledKey[resolved.ServiceDependency])

	return mergedTopology
}

func MustGetMarshalledKey[T any](resource T) string {
	bytes, err := json.Marshal(resource)
	if err != nil {
		panic("Failed to marshal resource")
	}
	return string(bytes)
}
