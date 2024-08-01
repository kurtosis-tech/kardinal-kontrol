package types

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ClusterResources struct {
	Services              []corev1.Service                      `json:"services"`
	Deployments           []appsv1.Deployment                   `json:"deployments"`
	VirtualServices       []v1alpha3.VirtualService             `json:"virtualServices"`
	DestinationRules      []v1alpha3.DestinationRule            `json:"destinationRules"`
	Gateway               v1alpha3.Gateway                      `json:"gateway"`
	EnvoyFilters          []v1alpha3.EnvoyFilter                `json:"envoy_filters"`
	AuthorizationPolicies []securityv1beta1.AuthorizationPolicy `json:"authorization_policies"`
}
