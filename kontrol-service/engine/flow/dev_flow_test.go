package flow

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func clusterTopologyExample() resolved.ClusterTopology {
	dummySpec := &appsv1.DeploymentSpec{}
	testPlugins := []*resolved.StatefulPlugin{
		{
			Name: "https://github.com/h4ck3rk3y/identity-plugin",
		},
	}

	// Create services
	frontendService := resolved.Service{
		ServiceID: "frontend",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
			Selector: map[string]string{
				"app": "frontend",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
	}

	cartService := resolved.Service{
		ServiceID: "cartservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 7070,
				},
			},
			Selector: map[string]string{
				"app": "cartservice",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
	}

	productCatalogService := resolved.Service{
		ServiceID: "productcatalogservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 3550,
				},
			},
			Selector: map[string]string{
				"app": "productcatalogservice",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
	}

	paymentService := resolved.Service{
		ServiceID: "paymentservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 50051,
				},
			},
			Selector: map[string]string{
				"app": "paymentservice",
			},
		},
		DeploymentSpec:  dummySpec,
		IsExternal:      false,
		IsStateful:      true,
		StatefulPlugins: testPlugins,
	}

	shippingService := resolved.Service{
		ServiceID: "shippingservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 50051,
				},
			},
			Selector: map[string]string{
				"app": "shippingservice",
			},
		},
		DeploymentSpec:  dummySpec,
		IsExternal:      false,
		IsStateful:      true,
		StatefulPlugins: testPlugins,
	}

	checkoutService := resolved.Service{
		ServiceID: "checkoutservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 5050,
				},
			},
			Selector: map[string]string{
				"app": "checkoutservice",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
	}

	recommendationService := resolved.Service{
		ServiceID: "recommendationservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "grpc",
					Port: 8080,
				},
			},
			Selector: map[string]string{
				"app": "recommendationservice",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
	}

	redisService := resolved.Service{
		ServiceID: "redis",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "redis",
					Port: 6379,
				},
			},
			Selector: map[string]string{
				"app": "redis",
			},
		},
		DeploymentSpec:  dummySpec,
		IsExternal:      false,
		IsStateful:      true,
		StatefulPlugins: testPlugins,
	}

	// Create service dependencies
	serviceDependencies := []resolved.ServiceDependency{
		{
			Service:          &frontendService,
			DependsOnService: &recommendationService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 8080,
			},
		},
		{
			Service:          &frontendService,
			DependsOnService: &productCatalogService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 3550,
			},
		},
		{
			Service:          &frontendService,
			DependsOnService: &cartService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 7070,
			},
		},
		{
			Service:          &frontendService,
			DependsOnService: &shippingService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 50051,
			},
		},
		{
			Service:          &frontendService,
			DependsOnService: &checkoutService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 5050,
			},
		},
		{
			Service:          &checkoutService,
			DependsOnService: &productCatalogService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 3550,
			},
		},
		{
			Service:          &checkoutService,
			DependsOnService: &cartService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 7070,
			},
		},
		{
			Service:          &checkoutService,
			DependsOnService: &shippingService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 50051,
			},
		},
		{
			Service:          &checkoutService,
			DependsOnService: &paymentService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 50051,
			},
		},
		{
			Service:          &cartService,
			DependsOnService: &redisService,
			DependencyPort: &v1.ServicePort{
				Name: "redis",
				Port: 6379,
			},
		},
		{
			Service:          &recommendationService,
			DependsOnService: &productCatalogService,
			DependencyPort: &v1.ServicePort{
				Name: "grpc",
				Port: 3550,
			},
		},
	}

	// Create ingress rules
	ingressRules := []*netv1.IngressRule{
		{
			Host: "online-boutique.com",
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: []netv1.HTTPIngressPath{
						{
							Path:    "/",
							Backend: netv1.IngressBackend{},
						},
					},
				},
			},
		},
	}

	// Create ingress
	ingress := resolved.Ingress{
		IngressID:    "main-ingress",
		IngressRules: ingressRules,
	}

	// Create cluster topology
	clusterTopology := resolved.ClusterTopology{
		Ingress: ingress,
		Services: []*resolved.Service{
			&frontendService,
			&cartService,
			&productCatalogService,
			&paymentService,
			&shippingService,
			&checkoutService,
			&recommendationService,
			&redisService,
		},
		ServiceDependecies: serviceDependencies,
	}

	// Use the clusterTopology as needed
	return clusterTopology
}

func getServiceRef(cluster *resolved.ClusterTopology, serviceID string) *resolved.Service {
	service, found := lo.Find(cluster.Services, func(item *resolved.Service) bool { return item.ServiceID == serviceID })
	if !found {
		panic(fmt.Sprintf("service with UUID %s not found", serviceID))
	}
	return service
}

func assertStateLessServices(t *testing.T, originalCluster *resolved.ClusterTopology, devCluster *resolved.ClusterTopology, services []string) {
	for _, serviceID := range services {
		originalService := getServiceRef(originalCluster, serviceID)
		devService := getServiceRef(devCluster, serviceID)
		require.Equal(t, false, originalService.IsStateful)
		require.Equal(t, originalService.IsStateful, devService.IsStateful)
		require.Equal(t, originalService, devService)
		require.Equal(t, originalService.Version, devService.Version)
	}
}

func assertStatefulServices(t *testing.T, originalCluster *resolved.ClusterTopology, devCluster *resolved.ClusterTopology, services []string) {
	for _, serviceID := range services {
		originalService := getServiceRef(originalCluster, serviceID)
		devService := getServiceRef(devCluster, serviceID)
		require.Equal(t, true, originalService.IsStateful)
		require.Equal(t, originalService.IsStateful, devService.IsStateful)
		require.NotEqual(t, originalService, devService)
		require.NotEqual(t, originalService.Version, devService.Version)
	}
}

func TestTopologyToGraph(t *testing.T) {
	cluster := clusterTopologyExample()
	g := topologyToGraph(cluster)
	targetService, found := FindServiceByID(cluster, "checkoutservice")
	require.Equal(t, found, true)

	resultGraph := findAllDownstreamStatefulPaths(targetService, g, cluster)
	fmt.Println("Paths:")
	for _, paths := range resultGraph {
		fmt.Println("Segs:")
		for _, dep := range paths {
			fmt.Println(dep)
		}
	}

	paymentservice := getServiceRef(&cluster, "paymentservice")
	shippingservice := getServiceRef(&cluster, "shippingservice")
	cartservice := getServiceRef(&cluster, "cartservice")
	redis := getServiceRef(&cluster, "redis")

	expected := [][]*resolved.Service{
		{targetService, paymentservice},
		{targetService, shippingservice},
		{targetService, cartservice, redis},
	}

	require.Equal(t, expected, resultGraph)
}

func TestDeepCopy(t *testing.T) {
	cluster := clusterTopologyExample()
	paymentService := getServiceRef(&cluster, "paymentservice")
	devCluster := DeepCopyClusterTopology(&cluster)
	devPaymentService := getServiceRef(devCluster, "paymentservice")
	require.Equal(t, true, paymentService.IsStateful)
	require.Equal(t, paymentService.IsStateful, devPaymentService.IsStateful)
	require.Equal(t, false, devPaymentService == paymentService)

	devPaymentService.Version = "test"
	require.NotEqual(t, devPaymentService.Version, paymentService.Version)
	paymentService.Version = "test"
	require.Equal(t, devPaymentService.Version, paymentService.Version)
}

func TestDeepCopyService(t *testing.T) {
	cluster := clusterTopologyExample()
	paymentService := getServiceRef(&cluster, "paymentservice")
	devPaymentService := DeepCopyService(paymentService)
	require.Equal(t, true, paymentService.IsStateful)
	require.Equal(t, paymentService.IsStateful, devPaymentService.IsStateful)
	require.Equal(t, false, devPaymentService == paymentService)
}

func TestDevFlowImmutability(t *testing.T) {
	cluster := clusterTopologyExample()
	checkoutservice := getServiceRef(&cluster, "checkoutservice")
	paymentservice := getServiceRef(&cluster, "paymentservice")
	devCluster, err := CreateDevFlow(plugins.PluginRunner{}, "dev-flow-1", "checkoutservice", *checkoutservice.DeploymentSpec, cluster)
	require.NoError(t, err)

	devPaymentService := getServiceRef(devCluster, "paymentservice")
	require.NotEqual(t, devPaymentService, paymentservice)
	require.NotEqual(t, devPaymentService.Version, paymentservice.Version)

	statefulServices := []string{
		"paymentservice",
		"shippingservice",
		"redis",
	}
	assertStatefulServices(t, &cluster, devCluster, statefulServices)

	statelessServices := []string{
		"frontend",
		"cartservice",
		"productcatalogservice",
		"checkoutservice",
		"recommendationservice",
	}
	assertStateLessServices(t, &cluster, devCluster, statelessServices)

	require.Equal(t, len(cluster.Services), len(devCluster.Services))
	require.Equal(t, len(cluster.ServiceDependecies), len(devCluster.ServiceDependecies))

	for _, deps := range devCluster.ServiceDependecies {
		require.Equal(t, true, lo.Contains(devCluster.Services, deps.Service))
		require.Equal(t, true, lo.Contains(devCluster.Services, deps.DependsOnService))
	}
}
