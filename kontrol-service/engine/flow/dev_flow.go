package flow

import (
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/samber/lo"
	"istio.io/api/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func createDevFlow(serviceID string, deploymentSpec appsv1.DeploymentSpec, topology resolved.ClusterTopology) types.ClusterResources {
	// TODO: Implement
	return types.ClusterResources{}
}

func topologyToGraph(topology resolved.ClusterTopology) graph.Graph[string, resolved.Service] {
	serviceHash := func(service resolved.Service) string {
		return service.ServiceID
	}
	graph := graph.New(serviceHash)

	for _, service := range topology.Services {
		graph.AddVertex(service)
	}

	for _, dependency := range topology.ServiceDependecies {
		graph.AddEdge(dependency.Service.ServiceID, dependency.DependsOnService.ServiceID)
	}

	return graph
}

func RenderClusterResources(clusterTopology *resolved.ClusterTopology, namespace string) types.ClusterResources {

	virtualServices := []istioclient.VirtualService{}
	for _, service := range clusterTopology.Services {
		var gateway *string
		var extHost *string
		if clusterTopology.IsIngressDestination(&service) {
			gateway = &clusterTopology.Ingress.IngressID
			extHost = clusterTopology.Ingress.GetHost()
		}
		virtualService := getVirtualService(&service, namespace, gateway, extHost)
		virtualServices = append(virtualServices, *virtualService)
	}

	return types.ClusterResources{
		Services: lo.Map(clusterTopology.Services, func(service resolved.Service, _ int) v1.Service {
			return *getService(&service, namespace)
		}),

		Deployments: lo.Map(clusterTopology.Services, func(service resolved.Service, _ int) appsv1.Deployment {
			return *getDeployment(&service, namespace)
		}),

		Gateway: *getGateway(&clusterTopology.Ingress, namespace),

		VirtualServices: virtualServices,

		DestinationRules: []istioclient.DestinationRule{},
	}
}

func getTCPRoute(service *resolved.Service, servicePort *v1.ServicePort) *v1alpha3.TCPRoute {
	return &v1alpha3.TCPRoute{
		Match: []*v1alpha3.L4MatchAttributes{{
			Port: uint32(servicePort.Port),
		}},
		Route: []*v1alpha3.RouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: service.ServiceID,
					Port: &v1alpha3.PortSelector{
						Number: uint32(servicePort.Port),
					},
				},
				Weight: 100,
			},
		},
	}
}

func getHTTPRoute(service *resolved.Service, servicePort *v1.ServicePort) *v1alpha3.HTTPRoute {
	return &v1alpha3.HTTPRoute{
		Route: []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   service.ServiceID,
					Subset: service.Version,
				},
				Weight: 100,
			},
		},
	}
}

func getVirtualService(service *resolved.Service, namespace string, gateway *string, extHost *string) *istioclient.VirtualService {
	// TODO: Support for multiple ports
	servicePort := &service.ServiceSpec.Ports[0]

	virtualServiceSpec := v1alpha3.VirtualService{}

	if servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP" {
		virtualServiceSpec.Http = []*v1alpha3.HTTPRoute{getHTTPRoute(service, servicePort)}
	} else {
		virtualServiceSpec.Tcp = []*v1alpha3.TCPRoute{getTCPRoute(service, servicePort)}
	}

	if gateway != nil {
		virtualServiceSpec.Gateways = []string{*gateway}
	}

	if extHost != nil {
		virtualServiceSpec.Hosts = []string{*extHost}
	} else {
		virtualServiceSpec.Hosts = []string{service.ServiceID}
	}

	return &istioclient.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.ServiceID, service.Version),
			Namespace: namespace,
		},
		Spec: virtualServiceSpec,
	}
}

func getService(service *resolved.Service, namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ServiceID,
			Namespace: namespace,
			Labels: map[string]string{
				"app": service.ServiceID,
			},
		},
		Spec: *service.ServiceSpec,
	}
}

func getDeployment(service *resolved.Service, namespace string) *appsv1.Deployment {

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.ServiceID, service.Version),
			Namespace: namespace,
			Labels: map[string]string{
				"app":     service.ServiceID,
				"version": service.Version,
			},
		},
		Spec: *service.DeploymentSpec,
	}

	numReplicas := int32(1)
	deployment.Spec.Replicas = int32Ptr(numReplicas)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}
	vol25pct := intstr.FromString("25%")
	deployment.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       &vol25pct,
			MaxUnavailable: &vol25pct,
		},
	}
	deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Annotations: map[string]string{
			"sidecar.istio.io/inject": "true",
		},
		Labels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}

	return &deployment
}

func getGateway(ingress *resolved.Ingress, namespace string) *istioclient.Gateway {
	extHosts := []string{}
	ingressHost := ingress.GetHost()
	if ingressHost != nil {
		extHosts = append(extHosts, *ingressHost)
	}

	return &istioclient.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingress.IngressID,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     ingress.IngressID,
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

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
