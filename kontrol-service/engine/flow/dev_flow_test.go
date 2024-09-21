package flow

import (
	"fmt"
	"testing"

	"kardinal.kontrol-service/database"
	"kardinal.kontrol-service/plugins"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
)

const dummyPluginName = "https://github.com/h4ck3rk3y/identity-plugin.git"

func clusterTopologyExample() resolved.ClusterTopology {
	dummySpec := &appsv1.DeploymentSpec{}
	testPlugins := []*resolved.StatefulPlugin{
		{
			Name: dummyPluginName,
		},
	}
	httpProtocol := "HTTP"

	// Create services
	frontendService := resolved.Service{
		ServiceID: "frontend",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:        "http",
					Port:        80,
					AppProtocol: &httpProtocol,
				},
			},
			Selector: map[string]string{
				"app": "frontend",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
		StatefulPlugins: []*resolved.StatefulPlugin{
			{
				Name:        dummyPluginName,
				ServiceName: "free-currency-api",
				Type:        "external",
				Args:        map[string]string{},
			},
		},
	}

	cartService := resolved.Service{
		ServiceID: "cartservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:        "grpc",
					Port:        7070,
					AppProtocol: &httpProtocol,
				},
			},
			Selector: map[string]string{
				"app": "cartservice",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
		StatefulPlugins: []*resolved.StatefulPlugin{
			{
				Name:        dummyPluginName,
				ServiceName: "neon-postgres-db",
				Type:        "external",
				Args:        map[string]string{},
			},
		},
	}

	productCatalogService := resolved.Service{
		ServiceID: "productcatalogservice",
		ServiceSpec: &v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:        "grpc",
					Port:        3550,
					AppProtocol: &httpProtocol,
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
					Name:        "grpc",
					Port:        50051,
					AppProtocol: &httpProtocol,
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
					Name:        "grpc",
					Port:        50051,
					AppProtocol: &httpProtocol,
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
					Name:        "grpc",
					Port:        5050,
					AppProtocol: &httpProtocol,
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
					Name:        "grpc",
					Port:        8080,
					AppProtocol: &httpProtocol,
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

	// external services
	neonService := resolved.Service{
		ServiceID:       "neon-postgres-db",
		ServiceSpec:     nil,
		DeploymentSpec:  nil,
		IsExternal:      true,
		IsStateful:      false, // neon is technically stateful but right now IsExternal and IsStateful are mutually exclusive
		StatefulPlugins: nil,
	}

	freeCurrencyApiService := resolved.Service{
		ServiceID:       "free-currency-api",
		ServiceSpec:     nil,
		DeploymentSpec:  nil,
		IsExternal:      true,
		IsStateful:      false,
		StatefulPlugins: nil,
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
			Service:          &frontendService,
			DependsOnService: &freeCurrencyApiService,
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
			Service:          &cartService,
			DependsOnService: &neonService,
			DependencyPort: &v1.ServicePort{
				Name: "postgres",
				Port: 5432,
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
	ingressRules := []netv1.IngressRule{
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
		ActiveFlowIDs: []string{"main-flow"},
		Ingresses: []netv1.Ingress{
			{
				Spec: netv1.IngressSpec{
					Rules: ingressRules,
				},
			},
		},
	}

	// Create cluster topology
	clusterTopology := resolved.ClusterTopology{
		FlowID:  "test-prod",
		Ingress: &ingress,
		Services: []*resolved.Service{
			&frontendService,
			&cartService,
			&productCatalogService,
			&paymentService,
			&shippingService,
			&checkoutService,
			&recommendationService,
			&redisService,
			&freeCurrencyApiService,
			&neonService,
		},
		ServiceDependencies: serviceDependencies,
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

func assertSameStateLessServices(t *testing.T, originalCluster *resolved.ClusterTopology, devCluster *resolved.ClusterTopology, services []string) {
	for _, serviceID := range services {
		originalService := getServiceRef(originalCluster, serviceID)
		devService := getServiceRef(devCluster, serviceID)
		require.False(t, originalService.IsStateful)
		require.Equal(t, originalService.IsStateful, devService.IsStateful)
		require.Equal(t, originalService, devService)
		require.Equal(t, originalService.Version, devService.Version)
	}
}

func assertDuplicatedStateLessServices(t *testing.T, originalCluster *resolved.ClusterTopology, devCluster *resolved.ClusterTopology, services []string) {
	for _, serviceID := range services {
		originalService := getServiceRef(originalCluster, serviceID)
		devService := getServiceRef(devCluster, serviceID)
		require.Equal(t, false, originalService.IsStateful)
		require.Equal(t, originalService.IsStateful, devService.IsStateful)
		require.NotEqual(t, originalService, devService)
		require.NotEqual(t, originalService.Version, devService.Version)
	}
}

func assertStatefulServices(t *testing.T, originalCluster *resolved.ClusterTopology, devCluster *resolved.ClusterTopology, services []string) {
	for _, serviceID := range services {
		originalService := getServiceRef(originalCluster, serviceID)
		devService := getServiceRef(devCluster, serviceID)
		require.True(t, originalService.IsStateful)
		require.Equal(t, originalService.IsStateful, devService.IsStateful)
		require.NotEqual(t, originalService, devService)
		require.NotEqual(t, originalService.Version, devService.Version)
	}
}

func getPluginRunner(t *testing.T) (*plugins.PluginRunner, func() error) {
	db, cleanUpDbFunc, err := database.NewSQLiteDB()
	require.NoError(t, err)
	err = db.Clear()
	require.NoError(t, err)
	err = db.AutoMigrate(&database.Tenant{}, &database.Flow{}, &database.PluginConfig{})
	require.NoError(t, err)
	_, err = db.GetOrCreateTenant("tenant-test")
	require.NoError(t, err)
	pluginRunner := plugins.NewPluginRunner(
		plugins.NewMockGitPluginProvider(plugins.MockGitHub),
		"tenant-test",
		db,
	)
	return pluginRunner, cleanUpDbFunc
}

func TestTopologyToGraph(t *testing.T) {
	cluster := clusterTopologyExample()
	g := topologyToGraph(&cluster)
	targetService, err := cluster.GetService("checkoutservice")
	require.Nil(t, err)

	resultGraph := findAllDownstreamStatefulPaths(targetService, g, &cluster)
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
	pluginRunner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	flowSpec := flow_spec.FlowPatch{
		FlowId: "dev-flow-1",
		ServicePatches: []flow_spec.ServicePatch{
			{
				Service:        "checkoutservice",
				DeploymentSpec: checkoutservice.DeploymentSpec,
			},
		},
	}

	devCluster, err := CreateDevFlow(pluginRunner, cluster, cluster, flowSpec)
	require.NoError(t, err)

	devCheckoutservice := getServiceRef(devCluster, "checkoutservice")
	require.NotEqual(t, devCheckoutservice, checkoutservice)
	require.NotEqual(t, devCheckoutservice.Version, checkoutservice.Version)

	statefulServices := []string{
		"paymentservice",
		"shippingservice",
		"redis",
	}
	assertStatefulServices(t, &cluster, devCluster, statefulServices)

	statelessServices := []string{
		"frontend",
		"productcatalogservice",
		"recommendationservice",
	}
	assertSameStateLessServices(t, &cluster, devCluster, statelessServices)

	dupStatelessServices := []string{
		"cartservice",
	}
	assertDuplicatedStateLessServices(t, &cluster, devCluster, dupStatelessServices)

	require.Equal(t, len(cluster.Services), len(devCluster.Services))
	require.Equal(t, len(cluster.ServiceDependencies), len(devCluster.ServiceDependencies))

	for _, deps := range devCluster.ServiceDependencies {
		require.Equal(t, true, lo.Contains(devCluster.Services, deps.Service))
		require.Equal(t, true, lo.Contains(devCluster.Services, deps.DependsOnService))
	}
}

func TestFlowMerging(t *testing.T) {
	cluster := clusterTopologyExample()
	checkoutservice := getServiceRef(&cluster, "checkoutservice")
	pluginRunner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	flowSpec := flow_spec.FlowPatch{
		FlowId: "dev-flow-1",
		ServicePatches: []flow_spec.ServicePatch{
			{
				Service:        "checkoutservice",
				DeploymentSpec: checkoutservice.DeploymentSpec,
			},
		},
	}

	devCluster, err := CreateDevFlow(pluginRunner, cluster, cluster, flowSpec)
	require.NoError(t, err)
	require.Equal(t, len(cluster.Services), len(devCluster.Services))
	require.Equal(t, len(cluster.ServiceDependencies), len(devCluster.ServiceDependencies))

	mergedTopology := MergeClusterTopologies(cluster, []resolved.ClusterTopology{*devCluster})

	extraModifiedServices := []string{
		"checkoutservice",
		"paymentservice",
		"shippingservice",
		"cartservice",
		"redis",
		"neon-postgres-db",
	}
	require.Equal(t, len(cluster.Services)+len(extraModifiedServices), len(mergedTopology.Services))

	nunExtraDeps := 9
	require.Equal(t, len(cluster.ServiceDependencies)+nunExtraDeps, len(mergedTopology.ServiceDependencies))
}

func TestExternalServicesFlowOnDependentService(t *testing.T) {
	cluster := clusterTopologyExample()

	cartservice, err := cluster.GetService("cartservice")
	require.NoError(t, err)
	pluginRunner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	flowSpec := flow_spec.FlowPatch{
		FlowId: "dev-flow-1",
		ServicePatches: []flow_spec.ServicePatch{
			{
				Service:        "cartservice",
				DeploymentSpec: cartservice.DeploymentSpec,
			},
		},
	}

	newClusterTopology, err := CreateDevFlow(pluginRunner, cluster, cluster, flowSpec)
	require.NoError(t, err)

	// the topology should have the same amount of services
	require.Equal(t, len(cluster.Services), len(newClusterTopology.Services))

	// but dev versions of cart service, redis (stateful service), neon postgres db (external service)
	expectedDevServices := []string{
		"cartservice",
		"neon-postgres-db",
		"redis",
	}
	assertDevVersionsFound(t, newClusterTopology, "dev-flow-1", expectedDevServices)
}

func TestExternalServicesCreateDevFlowOnNotDependentService(t *testing.T) {
	cluster := clusterTopologyExample()

	frontend, err := cluster.GetService("frontend")
	require.NoError(t, err)
	pluginRunner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	flowSpec := flow_spec.FlowPatch{
		FlowId: "dev-flow-1",
		ServicePatches: []flow_spec.ServicePatch{
			{
				Service:        "frontend",
				DeploymentSpec: frontend.DeploymentSpec,
			},
		},
	}

	newClusterTopology, err := CreateDevFlow(pluginRunner, cluster, cluster, flowSpec)
	require.NoError(t, err)

	// topology should have same amount of services
	require.Equal(t, len(cluster.Services), len(newClusterTopology.Services))

	// and dev versions of external service
	expectedDevServices := []string{
		"neon-postgres-db",
		"free-currency-api",
		"frontend",
		"cartservice",
		"shippingservice",
		"paymentservice",
	}
	assertDevVersionsFound(t, newClusterTopology, "dev-flow-1", expectedDevServices)
}

func assertDevVersionsFound(t *testing.T, devCluster *resolved.ClusterTopology, flowId string, serviceIds []string) {
	for _, serviceId := range serviceIds {
		_, found := lo.Find(devCluster.Services, func(service *resolved.Service) bool {
			return service.ServiceID == serviceId && service.Version == flowId
		})
		require.True(t, found)
	}
}
