package template

import (
	"fmt"

	"github.com/samber/lo"
	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/intstr"

	composetypes "github.com/compose-spec/compose-go/types"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kardinal.kloud-kontrol/types"
)

// Define the Service
func Service(serviceSpec types.ServiceSpec, namespaceSpec types.NamespaceSpec) v1.Service {
	return v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceSpec.Name,
			Namespace: namespaceSpec.Name,
			Labels: map[string]string{
				"app":     serviceSpec.Name,
				"version": serviceSpec.Version,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       fmt.Sprintf("tcp-%s", serviceSpec.Name),
					Port:       serviceSpec.Port,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(serviceSpec.TargetPort)),
				},
			},
			Selector: map[string]string{
				"app": serviceSpec.Name,
			},
		},
	}
}

func Deployment(serviceSpec types.ServiceSpec, namespaceSpec types.NamespaceSpec) apps.Deployment {
	vol25pct := intstr.FromString("25%")
	numReplicas := int32(1)
	configPorts := serviceSpec.Config.Ports

	containerPorts := lo.Map(configPorts, func(port composetypes.ServicePortConfig, _ int) v1.ContainerPort {
		return v1.ContainerPort{
			Name:          fmt.Sprintf("%s-%d", port.Protocol, port.Target),
			ContainerPort: int32(port.Target),
			Protocol:      v1.Protocol(port.Protocol),
		}
	})

	envVars := lo.MapToSlice(serviceSpec.Config.Environment, func(key string, value *string) v1.EnvVar {
		return v1.EnvVar{
			Name:  key,
			Value: *value,
		}
	})

	return apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", serviceSpec.Name, serviceSpec.Version),
			Namespace: namespaceSpec.Name,
			Labels: map[string]string{
				"app":     serviceSpec.Name,
				"version": serviceSpec.Version,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(numReplicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     serviceSpec.Name,
					"version": serviceSpec.Version,
				},
			},
			Strategy: apps.DeploymentStrategy{
				Type: apps.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &apps.RollingUpdateDeployment{
					MaxSurge:       &vol25pct,
					MaxUnavailable: &vol25pct,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: map[string]string{
						"app":     serviceSpec.Name,
						"version": serviceSpec.Version,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  serviceSpec.Config.ContainerName,
							Image: serviceSpec.Config.Image,
							Env:   envVars,
							Ports: containerPorts,
						},
					},
				},
			},
		},
	}
}

// TODO: split prod and and dev flows
func Gateway() istioclient.Gateway {
	return istioclient.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "voting-app",
			Namespace: "voting-app",
			Labels: map[string]string{
				"app":     "voting-app",
				"version": "v1",
			},
		},
		Spec: v1alpha3.Gateway{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: []string{
						"voting-app.localhost",
						"dev.voting-app.localhost",
					},
				},
			},
		},
	}
}

// Define the VirtualService
func FrontendVirtualService(serviceSpec types.ServiceSpec, namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.VirtualService {
	mainRoute := v1alpha3.HTTPRoute{
		Route: []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   serviceSpec.Name,
					Subset: serviceSpec.Version,
				},
				Weight: 100,
			},
		},
	}

	if traffic.HasMirroring {
		mainRoute.Mirror = &v1alpha3.Destination{
			Host:   serviceSpec.Name,
			Subset: traffic.MirrorToVersion,
		}
		mainRoute.MirrorPercentage = &v1alpha3.Percent{
			Value: 10,
		}

	}

	return istioclient.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceSpec.Name,
			Namespace: namespaceSpec.Name,
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{
				traffic.ExternalHostname,
			},
			Gateways: []string{
				traffic.GatewayName,
			},
			Http: []*v1alpha3.HTTPRoute{&mainRoute},
		},
	}
}

func FrontendDestinationRule(serviceSpec types.ServiceSpec, namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.DestinationRule {
	subsets := []*v1alpha3.Subset{
		{
			Name: "v1",
			Labels: map[string]string{
				"version": "v1",
			},
		},
	}

	if traffic.HasMirroring {
		devSubset := v1alpha3.Subset{
			Name: traffic.MirrorToVersion,
			Labels: map[string]string{
				"version": traffic.MirrorToVersion,
			},
		}
		subsets = append(subsets, &devSubset)
	}
	return istioclient.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceSpec.Name,
			Namespace: namespaceSpec.Name,
		},
		Spec: v1alpha3.DestinationRule{
			Host:    serviceSpec.Name,
			Subsets: subsets,
		},
	}
}

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
