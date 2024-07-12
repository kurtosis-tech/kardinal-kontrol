package types

import (
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
)

type Traffic struct {
	HasMirroring           bool
	MirrorPercentage       uint
	MirrorToVersion        string
	MirrorExternalHostname string
	ExternalHostname       string
	GatewayName            string
}

// TODO: Needs to: 1) Validate/restrict version and name, 2) assume just on port on TCP
// TODO: Remove dup ports and name
type ServiceSpec struct {
	Version    string
	Name       string
	Port       int32
	TargetPort int32
	Config     apitypes.ServiceConfig
}

type NamespaceSpec struct {
	Name string
}

type ServiceDependency struct {
	OriginService      *ServiceSpec
	DestinationService *ServiceSpec
}

type Cluster struct {
	Services            []*ServiceSpec
	ServiceDependencies []*ServiceDependency
	FrontdoorService    []*ServiceSpec
	TrafficSource       Traffic
	Namespace           NamespaceSpec
}
