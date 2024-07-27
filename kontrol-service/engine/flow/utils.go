package flow

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// Hack the to deep copy the topology. oh go! :(
// Plase, avoid using this function directly. Instead implement a direct and specific wrapper function
// DeepCopyDeploymentSpec or DeepCopyService. If you use it, make sure to the dst argument as a pointer
// and test it properly.
func unsafeDeepCopy(src, dst interface{}) {
	bytes, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, dst)
}

func deepCopySlice[T any](orig []T) []T {
	cpy := make([]T, len(orig))
	copy(cpy, orig)
	return cpy
}

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}

func DeepCopyDeploymentSpec(src *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	dst := &appsv1.DeploymentSpec{}
	unsafeDeepCopy(src, dst)
	return dst
}

func DeepCopyService(src *resolved.Service) *resolved.Service {
	dst := &resolved.Service{}
	unsafeDeepCopy(src, dst)
	return dst
}

func DeepCopyClusterTopology(src *resolved.ClusterTopology) *resolved.ClusterTopology {
	dst := &resolved.ClusterTopology{}
	unsafeDeepCopy(src, dst)
	return dst
}
