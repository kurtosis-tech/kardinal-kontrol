package fetcher

import (
	"context"
	"github.com/stretchr/testify/require"
	"kardinal.kontrol/kardinal-manager/kubernetes_client"
	"testing"
)

// This test can be executed and use Minikube dashboard and Kiali Dashboard to see the changes between prod apply and devInProd apply
func TestVotingAppDemoProdAndDevCase(t *testing.T) {
	kubernetesClient, err := kubernetes_client.CreateKubernetesClient()
	require.NoError(t, err)

	prodOnlyDemoConfigEndpoint := "https://gist.githubusercontent.com/leoporoli/d9afda02795f18abef04fa74afe3b555/raw/ac5123344a4cf2da26b747d69fb8ad6185a03723/prod-only-demo.json"

	prodFetcher := NewFetcher(kubernetesClient, prodOnlyDemoConfigEndpoint)

	ctx := context.Background()

	err = prodFetcher.fetchAndApply(ctx)
	require.NoError(t, err)

	//time.Sleep(2 * time.Minute)

	devInProdEndpoint := "https://gist.githubusercontent.com/leoporoli/565e55949c976d25eaedfa7433dd8a0e/raw/4b252105fcc5ab8b07e4a8bb183428253c304268/dev-in-prod-demo.json"

	devInProdFetcher := NewFetcher(kubernetesClient, devInProdEndpoint)

	err = devInProdFetcher.fetchAndApply(ctx)
	require.NoError(t, err)

	//time.Sleep(2 * time.Minute)

	//Executing prodFetcher again to remove the Dev resources
	err = prodFetcher.fetchAndApply(ctx)
	require.NoError(t, err)
}
