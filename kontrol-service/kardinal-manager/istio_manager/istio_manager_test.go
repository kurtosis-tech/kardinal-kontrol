package istio_manager

import (
	"context"
	"github.com/stretchr/testify/require"
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
// assume
// - default k8s namespace contains the services from the sample bookinfo application: https://istio.io/latest/docs/examples/bookinfo/
// - a destination rule for reviews service has been preconfigured
// - a virtual service for reviews service has been preconfigured
func TestIstioManagerWorkflows(t *testing.T) {
	ctx := context.Background()
	istioManager, err := getIstioManagerForTesting()
	require.NoError(t, err)

	// verify that there exists a destination rule for the "reviews" service that only sends traffic to v1

	// verify that there exists a virtual service for the "reviews" service

	// add a subset to the reviews destination rule that to register v2 of the reviews service

	// add a routing rule that
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
