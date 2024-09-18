package types

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	namespaceApiVersion = "v1"
	namespaceKind       = "Namespace"
	istioLabel          = "istio-injection"
	enabledIstioValue   = "enabled"
)

type ClusterResources struct {
	Services              []corev1.Service                      `json:"services"`
	Deployments           []appsv1.Deployment                   `json:"deployments"`
	VirtualServices       []v1alpha3.VirtualService             `json:"virtualServices"`
	DestinationRules      []v1alpha3.DestinationRule            `json:"destinationRules"`
	EnvoyFilters          []v1alpha3.EnvoyFilter                `json:"envoy_filters"`
	AuthorizationPolicies []securityv1beta1.AuthorizationPolicy `json:"authorization_policies"`
	Gateways              []gateway.Gateway                     `json:"gateways"`
	HTTPRoutes            []gateway.HTTPRoute                   `json:"http_routes"`
}

func NewNamespaceWithIstioEnabled(namespaceName string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: namespaceApiVersion,
			Kind:       namespaceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				istioLabel: enabledIstioValue,
			},
		},
	}
}
