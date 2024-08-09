package flow_spec

import v1 "k8s.io/api/apps/v1"

type FlowPatch struct {
	FlowId         string
	ServicePatches []ServicePatch
}

type ServicePatch struct {
	Service        string
	DeploymentSpec *v1.DeploymentSpec
}
