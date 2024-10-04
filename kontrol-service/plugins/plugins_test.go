package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kardinal.kontrol-service/database"
	kardinal "kardinal.kontrol-service/types/kardinal"
)

const (
	simplePlugin   = "https://github.com/fake-org/kardinal-simple-plugin-example.git"
	complexPlugin  = "https://github.com/fake-org/kardinal-slightly-more-complex-plugin-example.git"
	identityPlugin = "https://github.com/fake-org/kardinal-identity-plugin-example.git"
	redisPlugin    = "https://github.com/fake-org/kardinal-redis-sidecar-plugin-example.git"
	flowUuid       = "test-flow-uuid"
)

var serviceSpecs = []corev1.ServiceSpec{}

var deploymentSpecs = []appv1.DeploymentSpec{
	{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "helloworld",
			},
		},
		Replicas: int32Ptr(1),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "helloworld",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "helloworld",
						Image: "karthequian/helloworld:latest",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 80,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "REDIS",
								Value: "ip_addr",
							},
						},
					},
				},
			},
		},
	},
	{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "helloworld-version2",
			},
		},
		Replicas: int32Ptr(4),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "helloworld-version2",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "helloworld",
						Image: "karthequian/helloworld:latest",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 80,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "REDIS",
								Value: "ip_addr",
							},
							{
								Name:  "FOO", // adding extra env var to compare differences
								Value: "bar",
							},
						},
					},
				},
			},
		},
	},
}

var workloadSpec1 = kardinal.NewDeploymentWorkloadSpec(deploymentSpecs[0])
var workloadSpec2 = kardinal.NewDeploymentWorkloadSpec(deploymentSpecs[1])

var workloadSpecs = []*kardinal.WorkloadSpec{
	&workloadSpec1,
	&workloadSpec2,
}

func getPluginRunner(t *testing.T) (*PluginRunner, func() error) {
	db, cleanUpDbFunc, err := database.NewSQLiteDB()
	require.NoError(t, err)
	err = db.Clear()
	require.NoError(t, err)
	err = db.AutoMigrate(&database.Tenant{}, &database.Flow{}, &database.PluginConfig{})
	require.NoError(t, err)
	_, err = db.GetOrCreateTenant("tenant-test")
	require.NoError(t, err)
	pluginRunner := NewPluginRunner(
		NewMockGitPluginProvider(MockGitHub),
		"tenant-test",
		db,
	)
	return pluginRunner, cleanUpDbFunc
}

func TestSimplePlugin(t *testing.T) {
	runner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	arguments := map[string]string{
		"text_to_replace": "helloworld",
	}

	updatedDeploymentSpecs, configMap, err := runner.CreateFlow(simplePlugin, serviceSpecs, workloadSpecs, flowUuid, arguments)
	require.NoError(t, err)

	for idx, updatedDeploymentSpec := range updatedDeploymentSpecs {
		// Check if the deployment spec was updated correctly
		require.Equal(t, "the-text-has-been-replaced", updatedDeploymentSpec.GetTemplateSpec().Containers[0].Name)
		require.Equal(t, workloadSpecs[idx].GetTemplateSpec().Containers[0].Env, updatedDeploymentSpec.GetTemplateSpec().Containers[0].Env)
	}

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Equal(t, "helloworld", configMapData["original_text"])

	err = runner.DeleteFlow(simplePlugin, flowUuid)
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, err = runner.getConfigForFlow(flowUuid)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no config map found")
}

func TestIdentityPlugin(t *testing.T) {
	runner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	updatedServiceSpec, configMap, err := runner.CreateFlow(identityPlugin, serviceSpecs, workloadSpecs, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Check if the deployment spec was updated correctly
	require.Equal(t, workloadSpecs, updatedServiceSpec)

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{}, configMapData)

	err = runner.DeleteFlow(identityPlugin, flowUuid)
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, err = runner.getConfigForFlow(flowUuid)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no config map found")
}

func TestComplexPlugin(t *testing.T) {
	runner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	updatedDeploymentSpecs, configMap, err := runner.CreateFlow(complexPlugin, serviceSpecs, workloadSpecs, flowUuid, map[string]string{})
	require.NoError(t, err)

	for _, updatedDeploymentSpec := range updatedDeploymentSpecs {
		// Check if the deployment spec was updated correctly
		require.NotEqual(t, "ip_addr", updatedDeploymentSpec.GetTemplateSpec().Containers[0].Env[0].Value)
		require.Regexp(t, `\b(?:\d{1,3}\.){3}\d{1,3}\b`, updatedDeploymentSpec.GetTemplateSpec().Containers[0].Env[0].Value)
	}

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Equal(t, "ip_addr", configMapData["original_value"])

	err = runner.DeleteFlow(complexPlugin, flowUuid)
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, err = runner.getConfigForFlow(flowUuid)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no config map found")
}

func TestRedisPluginTest(t *testing.T) {
	runner, cleanUpDbFunc := getPluginRunner(t)
	defer cleanUpDbFunc()

	updatedDeploymentSpecs, configMap, err := runner.CreateFlow(redisPlugin, serviceSpecs, workloadSpecs, flowUuid, map[string]string{})
	require.NoError(t, err)

	for _, updatedDeploymentSpec := range updatedDeploymentSpecs {
		// Check if the deployment spec was updated correctly
		require.Equal(t, "kurtosistech/redis-proxy-overlay:latest", updatedDeploymentSpec.GetTemplateSpec().Containers[0].Image)
	}

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Empty(t, configMapData)

	err = runner.DeleteFlow(complexPlugin, flowUuid)
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, err = runner.getConfigForFlow(flowUuid)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no config map found")
}

func int32Ptr(i int32) *int32 { return &i }
