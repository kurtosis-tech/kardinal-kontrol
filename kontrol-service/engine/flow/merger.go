package flow

import (
	"github.com/samber/lo"
	v1 "k8s.io/api/networking/v1"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func MergeClusterTopologies(baseTopology resolved.ClusterTopology, clusterTopologies []resolved.ClusterTopology) *resolved.ClusterTopology {
	mergedTopology := &resolved.ClusterTopology{
		FlowID:             "all",
		Services:           deepCopySlice(baseTopology.Services),
		ServiceDependecies: deepCopySlice(baseTopology.ServiceDependecies),
		Ingress:            deepCopySlice(baseTopology.Ingress),
	}
	for _, topology := range clusterTopologies {
		mergedTopology.Services = append(mergedTopology.Services, topology.Services...)
		mergedTopology.ServiceDependecies = append(mergedTopology.ServiceDependecies, topology.ServiceDependecies...)
		mergedTopology.Ingress = append(mergedTopology.Ingress, topology.Ingress...)
	}

	mergedTopology.Services = lo.Uniq(mergedTopology.Services)
	mergedTopology.ServiceDependecies = lo.Uniq(mergedTopology.ServiceDependecies)
	mergedTopology.Ingress = foldAllIngress(mergedTopology.Ingress)

	// fmt.Printf("topology: %s\n", SPrintJSONClusterTopology(mergedTopology))
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
