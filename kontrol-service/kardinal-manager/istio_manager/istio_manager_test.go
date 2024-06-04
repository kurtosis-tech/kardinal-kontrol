package istio_manager

import (
	"context"
	"github.com/stretchr/testify/require"
	istio "istio.io/api/networking/v1alpha3"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"testing"
)

func TestIstIoManager(t *testing.T) {
	ctx := context.Background()
	istioManager, err := getIstioManagerForTesting()
	require.NoError(t, err)

	virtualServices, err := istioManager.GetVirtualServices(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, virtualServices)
}

// This test is to demonstrate using the IstIoManager to accomplish certain workflows
// assumes
// - default k8s namespace contains the services from the sample bookinfo application: https://istio.io/latest/docs/examples/bookinfo/
// - a destination rule for reviews service has been preconfigured with one version of reviews
// - a virtual service for reviews service has been preconfigured with one routing rule
func TestIstioManagerWorkflows(t *testing.T) {
	ctx := context.Background()
	istioManager, err := getIstioManagerForTesting()
	require.NoError(t, err)

	// verify that there exists a destination rule for the "reviews" service that only sends traffic to v1
	reviewsDestinationRule, err := istioManager.GetDestinationRule(ctx, "reviews")
	require.NoError(t, err)
	require.NotEmpty(t, reviewsDestinationRule)
	require.NotEmpty(t, reviewsDestinationRule.Spec.Subsets)
	require.Equal(t, reviewsDestinationRule.Spec.Subsets[0].Name, "v1")

	// verify that there exists a virtual service for the "reviews" service
	reviewsVirtualService, err := istioManager.GetVirtualService(ctx, "reviews")
	require.NoError(t, err)
	require.NotEmpty(t, reviewsVirtualService)
	require.NotEmpty(t, reviewsVirtualService.Spec.Http)
	// TODO: may want to implement types in house to manage some of this stuff but for now jus use objects directly
	require.Equal(t, reviewsVirtualService.Spec.Http[0].Route[0].Destination.Host, "reviews")
	require.Equal(t, reviewsVirtualService.Spec.Http[0].Route[0].Destination.Subset, "v1")

	// register a new version of the reviews service
	v2subset := &istio.Subset{
		Name: "v2",
		Labels: map[string]string{
			"version": "v2",
		},
		TrafficPolicy: nil,
	}
	err = istioManager.AddSubset(ctx, "reviews", v2subset)
	require.NoError(t, err)

	// add a routing rule that splits traffic between v1 and v2
	splitTraffic5050Rule := &istio.HTTPRoute{
		Route: []*istio.HTTPRouteDestination{
			{
				Destination: &istio.Destination{
					Host:   "reviews",
					Subset: "v1",
					Port:   nil,
				},
				Weight: 50,
			},
			{
				Destination: &istio.Destination{
					Host:   "reviews",
					Subset: "v2",
					Port:   nil,
				},
				Weight: 50,
			},
		},
	}
	// can consider adjusting the AddRoutingRule api to only take in params we care about to make the api easier to use but again for now, KISS till we know more about usecases
	err = istioManager.AddRoutingRule(ctx, "reviews", splitTraffic5050Rule)
}

func getIstioManagerForTesting() (*IstIoManager, error) {
	// for now pick up local k8s config and err if it doesn't exist
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	istioManager, err := CreateIstIoManager(k8sConfig)
	if err != nil {
		return nil, err
	}
	return istioManager, nil
}
