package types

import (
	composetypes "github.com/compose-spec/compose-go/types"
)

type Traffic struct {
	HasMirroring     bool
	MirrorPercentage uint
	MirrorToVersion  string
	ExternalHostname string
	GatewayName      string
}

// TODO: Needs to: 1) Validate/restrict version and name, 2) assume just on port on TCP
// TODO: Remove dup ports and name
type ServiceSpec struct {
	Version    string
	Name       string
	Port       int32
	TargetPort int32
	Config     composetypes.ServiceConfig
}

type NamespaceSpec struct {
	Name string
}
