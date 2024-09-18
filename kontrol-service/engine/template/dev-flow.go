package template

import (
	"fmt"
	"sort"

	"github.com/samber/lo"
	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/intstr"

	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"

	"kardinal.kontrol-service/types"
)

func RenderClusterResources(cluster types.Cluster) types.ClusterResources {
	backendVSs := lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) istioclient.VirtualService {
		return BackendVirtualService(service, cluster.Namespace, cluster.TrafficSource)
	})

	return types.ClusterResources{
		Services: lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) v1.Service {
			return Service(service, cluster.Namespace)
		}),

		Deployments: lo.Map(cluster.Services, func(service *types.ServiceSpec, _ int) apps.Deployment {
			return Deployment(service, cluster.Namespace)
		}),

		// Gateways: Gateway(cluster.Namespace, cluster.TrafficSource),
		Gateways: lo.Map(cluster.TrafficSource.Gateways, func(item *gateway.Gateway, _ int) gateway.Gateway { return *item }),

		VirtualServices: backendVSs,

		DestinationRules: []istioclient.DestinationRule{},
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
	serviceContainer := serviceSpec.Config.Deployment.Spec.Template.Spec.Containers[0]
	containerPorts := serviceContainer.Ports
	envVars := serviceContainer.Env

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
							Name:            serviceContainer.Name,
							Image:           serviceContainer.Image,
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

// Helper function to create a sorted slice from a map with string key and any value
func mapWithStringKeyToSortedSlice[V any, R any](in map[string]V, iterate func(key string, value V) R) []R {
	result := make([]R, 0, len(in))

	inKeys := make([]string, 0)
	for inKey := range in {
		inKeys = append(inKeys, inKey)
	}
	sort.Strings(inKeys)
	for _, mapKey := range inKeys {
		mapValue := in[mapKey]
		result = append(result, iterate(mapKey, mapValue))
	}
	return result
}
