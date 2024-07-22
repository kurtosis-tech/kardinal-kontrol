package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
	"kardinal.kontrol-service/test"
)

func TestServiceConfigsToClusterTopology(t *testing.T) {
	testServiceConfigs := test.GetServiceConfigs()

	cluster, err := GenerateClusterTopology(testServiceConfigs)
	if err != nil {
		t.Errorf("Error generating cluster: %s", err)
	}

	redisProdService := cluster.Services[0]
	require.Equal(t, redisProdService.ServiceID, "redis-prod")
	require.Equal(t, redisProdService.IsExternal, false)
	require.Equal(t, redisProdService.IsStateful, true)
	statefulPlugin := redisProdService.StatefulPlugins[0]
	require.Equal(t, statefulPlugin.Name, "github.com/kardinaldev/redis-db-sidecar-plugin:36ed9a4")
	require.Equal(t, *redisProdService.ServiceSpec, testServiceConfigs[0].Service.Spec)
	require.Equal(t, *redisProdService.DeploymentSpec, testServiceConfigs[0].Deployment.Spec)

	votingAppUIService := cluster.Services[1]
	require.Equal(t, votingAppUIService.ServiceID, "voting-app-ui")
	require.Equal(t, votingAppUIService.IsExternal, false)
	require.Equal(t, votingAppUIService.IsStateful, false)
	require.Equal(t, *votingAppUIService.ServiceSpec, testServiceConfigs[1].Service.Spec)
	require.Equal(t, *votingAppUIService.DeploymentSpec, testServiceConfigs[1].Deployment.Spec)

	dependency := cluster.ServiceDependecies[0]
	require.Equal(t, *dependency.Service, votingAppUIService)
	require.Equal(t, *dependency.DependsOnService, redisProdService)
	require.Equal(t, *dependency.DependencyPort, testServiceConfigs[0].Service.Spec.Ports[0])

	ingressService := cluster.Ingress
	require.Equal(t, ingressService.IngressUUID, "voting-app-lb")
}
