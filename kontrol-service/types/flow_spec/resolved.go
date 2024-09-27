package flow_spec

type FlowPatch struct {
	FlowId         string
	ServicePatches []ServicePatch
}

type ServicePatch struct {
	Service string
	Image   string
}
