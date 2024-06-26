package types

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ClusterResources struct {
	Services         []corev1.Service           `json:"services"`
	Deployments      []appsv1.Deployment        `json:"deployments"`
	VirtualServices  []v1alpha3.VirtualService  `json:"virtualServices"`
	DestinationRules []v1alpha3.DestinationRule `json:"destinationRules"`
	Gateway          v1alpha3.Gateway           `json:"gateway"`
}
