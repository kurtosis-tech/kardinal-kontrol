package fetcher

import (
	"context"
	"github.com/stretchr/testify/require"
	"kardinal.kontrol/kardinal-manager/cluster_manager"
	"testing"
	"time"
)

// This test can be executed and use Minikube dashboard and Kiali Dashboard to see the changes between prod apply and devInProd apply
func TestVotingAppDemoProdAndDevCase(t *testing.T) {
	clusterManager, err := cluster_manager.CreateClusterManager()
	require.NoError(t, err)

	prodOnlyDemoConfigEndpoint := "https://gist.githubusercontent.com/leoporoli/d9afda02795f18abef04fa74afe3b555/raw/ac5123344a4cf2da26b747d69fb8ad6185a03723/prod-only-demo.json"

	prodFetcher := NewFetcher(clusterManager, prodOnlyDemoConfigEndpoint)

	ctx := context.Background()

	err = prodFetcher.fetchAndApply(ctx)
	require.NoError(t, err)

	// Sleep to check the Cluster topology in Minikube and Kiali, prod topology should be created in voting-app namespace
	time.Sleep(2 * time.Minute)

	devInProdEndpoint := "https://gist.githubusercontent.com/leoporoli/565e55949c976d25eaedfa7433dd8a0e/raw/4b252105fcc5ab8b07e4a8bb183428253c304268/dev-in-prod-demo.json"

	devInProdFetcher := NewFetcher(clusterManager, devInProdEndpoint)

	err = devInProdFetcher.fetchAndApply(ctx)
	require.NoError(t, err)

	// Sleep to check the Cluster topology in Minikube and Kiali, dev topology should be added in voting-app namespace
	time.Sleep(2 * time.Minute)

	//Executing prodFetcher again to remove the Dev resources
	err = prodFetcher.fetchAndApply(ctx)
	require.NoError(t, err)

	// Now you can check that dev components has been removed from the cluster
}
