package flow_spec

type FlowPatchSpec struct {
	FlowId         string
	ServicePatches []ServicePatchSpec
}

type ServicePatchSpec struct {
	Service               string
	Image                 string
	EnvVarOverrides       map[string]string
	SecretEnvVarOverrides map[string]string
}
