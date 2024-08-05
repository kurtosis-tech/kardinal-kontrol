package flow

import (
	"github.com/samber/lo"
	v1 "k8s.io/api/networking/v1"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func MergeClusterTopologies(baseTopology resolved.ClusterTopology, clusterTopologies []resolved.ClusterTopology) *resolved.ClusterTopology {
	mergedTopology := &resolved.ClusterTopology{
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

	mergedTopology.Services = lo.Uniq(mergedTopology.Services)
	mergedTopology.ServiceDependencies = lo.Uniq(mergedTopology.ServiceDependencies)
	mergedTopology.Ingresses = foldAllIngress(mergedTopology.Ingresses)

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
