package flow_spec

import kardinal "kardinal.kontrol-service/types/kardinal"

type FlowPatch struct {
	FlowId         string
	ServicePatches []ServicePatch
}

type ServicePatch struct {
	Service      string
	WorkloadSpec *kardinal.WorkloadSpec
}
