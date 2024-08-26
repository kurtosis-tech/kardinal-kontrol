package flow

import (
	"encoding/json"
	"github.com/samber/lo"
	v1 "k8s.io/api/networking/v1"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func MergeClusterTopologies(baseTopology resolved.ClusterTopology, clusterTopologies []resolved.ClusterTopology) (mergedTopology *resolved.ClusterTopology) {
	mergedTopology = &resolved.ClusterTopology{
		FlowID:              "all",
		Services:            deepCopySlice(baseTopology.Services),
		ServiceDependencies: deepCopySlice(baseTopology.ServiceDependencies),
		Ingresses:           deepCopySlice(baseTopology.Ingresses),
	}
	for _, topology := range clusterTopologies {
		mergedTopology.Services = append(mergedTopology.Services, topology.Services...)
		mergedTopology.ServiceDependencies = append(mergedTopology.ServiceDependencies, topology.ServiceDependencies...)
		mergedTopology.Ingresses = append(mergedTopology.Ingresses, topology.Ingresses...)
	}

	//TODO improve the filtering method, we could implement the `Service.Equal` method to compare and filter the services
	//TODO and inside this method we could use the k8s service marshall method (https://pkg.go.dev/k8s.io/api/core/v1#Service.Marsha) and also the same for other k8s fields
	//TODO it should be faster
	mergedTopology.Services = lo.UniqBy(mergedTopology.Services, MustGetMarshalledKey[*resolved.Service])
	mergedTopology.ServiceDependencies = lo.UniqBy(mergedTopology.ServiceDependencies, MustGetMarshalledKey[resolved.ServiceDependency])
	mergedTopology.Ingresses = foldAllIngress(mergedTopology.Ingresses)

	return mergedTopology
}

func foldAllIngress(ingresses []*resolved.Ingress) []*resolved.Ingress {
	groups := lo.GroupBy(ingresses, func(item *resolved.Ingress) string { return item.IngressID })
	return lo.MapToSlice(groups, func(key string, value []*resolved.Ingress) *resolved.Ingress {
		merged := resolved.Ingress{
			IngressID:     key,
			ActiveFlowIDs: lo.Uniq(lo.FlatMap(value, func(ingress *resolved.Ingress, _ int) []string { return ingress.ActiveFlowIDs })),
			IngressRules:  lo.Uniq(lo.FlatMap(value, func(ingress *resolved.Ingress, _ int) []*v1.IngressRule { return ingress.IngressRules })),
			ServiceSpec:   value[0].ServiceSpec,
		}
		return &merged
	})
}

func MustGetMarshalledKey[T any](resource T) string {
	bytes, err := json.Marshal(resource)
	if err != nil {
		panic("Failed to marshal resource")
	}
	return string(bytes)
}
