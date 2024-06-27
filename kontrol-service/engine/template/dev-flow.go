package template

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/intstr"

	composetypes "github.com/compose-spec/compose-go/types"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kardinal.kontrol-service/types"
)

func RenderClusterResources(cluster types.Cluster) types.ClusterResources {
	backendServices := lo.Filter(cluster.Services, func(service *types.ServiceSpec, _ int) bool { return lo.Contains(cluster.FrontdoorService, service) })
	backendVSs := lo.Map(backendServices, func(service *types.ServiceSpec, _ int) istioclient.VirtualService {
		return BackendVirtualService(service, cluster.Namespace, cluster.TrafficSource)
	})
	frontendVS := lo.Map(cluster.FrontdoorService, func(service *types.ServiceSpec, _ int) istioclient.VirtualService {
		return FrontendVirtualService(service, cluster.Namespace, cluster.TrafficSource)
	})

	return types.ClusterResources{
		Services: lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) v1.Service {
			return Service(service, cluster.Namespace)
		}),

		Deployments: lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) apps.Deployment {
			return Deployment(service, cluster.Namespace)
		}),

		Gateway: Gateway(cluster.Namespace, cluster.TrafficSource),

		VirtualServices: append(backendVSs, frontendVS...),

		DestinationRules: []istioclient.DestinationRule{
			FrontendDestinationRule(cluster.FrontdoorService, cluster.Namespace, cluster.TrafficSource),
		},
	}
}

// Define the Service
func Service(serviceSpec *types.ServiceSpec, namespaceSpec types.NamespaceSpec) v1.Service {
	return v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceSpec.Name,
			Namespace: namespaceSpec.Name,
			Labels: map[string]string{
				"app": serviceSpec.Name,
				// "version": serviceSpec.Version,
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

func Deployment(serviceSpec *types.ServiceSpec, namespaceSpec types.NamespaceSpec) apps.Deployment {
	vol25pct := intstr.FromString("25%")
	numReplicas := int32(1)
	configPorts := serviceSpec.Config.Ports

	containerPorts := lo.Map(configPorts, func(port composetypes.ServicePortConfig, _ int) v1.ContainerPort {
		return v1.ContainerPort{
			Name:          fmt.Sprintf("%s-%d", port.Protocol, port.Target),
			ContainerPort: int32(port.Target),
			Protocol:      v1.Protocol(strings.ToUpper(port.Protocol)),
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
							Name:            serviceSpec.Config.ContainerName,
							Image:           serviceSpec.Config.Image,
							ImagePullPolicy: "IfNotPresent",
							Env:             envVars,
							Ports:           containerPorts,
						},
					},
				},
			},
		},
	}
}

// TODO: split prod and and dev flows
func Gateway(namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.Gateway {
	extHosts := []string{
		traffic.ExternalHostname,
	}
	if traffic.HasMirroring {
		extHosts = append(extHosts, traffic.MirrorExternalHostname)
	}

	return istioclient.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gateway",
			Namespace: namespaceSpec.Name,
			Labels: map[string]string{
				"app":     "gateway",
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
					Hosts: extHosts,
				},
			},
		},
	}
}

// Define the VirtualService
func FrontendVirtualService(serviceSpec *types.ServiceSpec, namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.VirtualService {
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

	// TODO: Remove hardcoded prod, too fragile :(
	extHost := traffic.ExternalHostname
	if serviceSpec.Version != "prod" {
		extHost = traffic.MirrorExternalHostname
	}

	// TODO: Remove hardcoded prod, too fragile :(
	if traffic.HasMirroring && serviceSpec.Version == "prod" {
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
			Name:      fmt.Sprintf("%s-%s", serviceSpec.Name, serviceSpec.Version),
			Namespace: namespaceSpec.Name,
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{
				extHost,
			},
			Gateways: []string{
				traffic.GatewayName,
			},
			Http: []*v1alpha3.HTTPRoute{&mainRoute},
		},
	}
}

// Define the VirtualService
func BackendVirtualService(serviceSpec *types.ServiceSpec, namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.VirtualService {
	mainRoute := v1alpha3.TCPRoute{
		Match: []*v1alpha3.L4MatchAttributes{{
			Port: uint32(serviceSpec.Port),
		}},
		Route: []*v1alpha3.RouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: serviceSpec.Name,
					Port: &v1alpha3.PortSelector{
						Number: uint32(serviceSpec.Port),
					},
				},
				Weight: 100,
			},
		},
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
				serviceSpec.Name,
			},
			Tcp: []*v1alpha3.TCPRoute{&mainRoute},
		},
	}
}

func FrontendDestinationRule(services []*types.ServiceSpec, namespaceSpec types.NamespaceSpec, traffic types.Traffic) istioclient.DestinationRule {
	name := "frontendRule"

	if len(services) > 0 && services[0] != nil {
		name = services[0].Name
	}
	subsets := lo.Map(services, func(service *types.ServiceSpec, _ int) *v1alpha3.Subset {
		return &v1alpha3.Subset{
			Name: service.Version,
			Labels: map[string]string{
				"version": service.Version,
			},
		}
	})

	if traffic.HasMirroring {
		isMirrorVersionAlreadySet := lo.ContainsBy(subsets, func(item *v1alpha3.Subset) bool {
			return item.Name == traffic.MirrorToVersion
		})
		if !isMirrorVersionAlreadySet {
			devSubset := v1alpha3.Subset{
				Name: traffic.MirrorToVersion,
				Labels: map[string]string{
					"version": traffic.MirrorToVersion,
				},
			}
			subsets = append(subsets, &devSubset)
		}
	}
	return istioclient.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespaceSpec.Name,
		},
		Spec: v1alpha3.DestinationRule{
			Host:    name,
			Subsets: subsets,
		},
	}
}

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
