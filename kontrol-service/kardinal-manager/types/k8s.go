package types

import (
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type ClusterResources struct {
	Services         []v1.Service                  `json:"services"`
	Deployments      []apps.Deployment             `json:"deployments"`
	VirtualServices  []istioclient.VirtualService  `json:"virtualServices"`
	DestinationRules []istioclient.DestinationRule `json:"destinationRules"`
	Gateway          istioclient.Gateway           `json:"gateway"`
}
