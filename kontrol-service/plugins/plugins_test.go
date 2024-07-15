package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	simplePlugin  = "https://github.com/h4ck3rk3y/a-test-plugin"
	complexPlugin = "https://github.com/h4ck3rk3y/slightly-more-complex-plugin"
	redisPlugin   = "https://github.com/h4ck3rk3y/redis-sidecar-plugin"
	flowUuid      = "test-flow-uuid"
)

var serviceSpec = corev1.ServiceSpec{}

var deploymentSpec = appv1.DeploymentSpec{
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
}

func TestSimplePlugin(t *testing.T) {
	t.Skip("skipping this as this pulls from the internet")
	runner := NewPluginRunner()

	arguments := map[string]string{
		"text_to_replace": "helloworld",
	}

	updatedDeploymentSpec, configMap, err := runner.CreateFlow(simplePlugin, serviceSpec, deploymentSpec, flowUuid, arguments)
	require.NoError(t, err)

	// Check if the deployment spec was updated correctly
	require.Equal(t, "the-text-has-been-replaced", updatedDeploymentSpec.Template.Labels["app"])
	require.Equal(t, "the-text-has-been-replaced", updatedDeploymentSpec.Selector.MatchLabels["app"])
	require.Equal(t, "the-text-has-been-replaced", updatedDeploymentSpec.Template.Spec.Containers[0].Name)

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Equal(t, "helloworld", configMapData["original_text"])

	err = runner.DeleteFlow(simplePlugin, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, exists := runner.memory[flowUuid]
	require.False(t, exists)
}

func TestComplexPlugin(t *testing.T) {
	t.Skip("skipping this as this pulls from the internet")
	runner := NewPluginRunner()

	updatedServiceSpec, configMap, err := runner.CreateFlow(complexPlugin, serviceSpec, deploymentSpec, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Check if the deployment spec was updated correctly
	require.NotEqual(t, "ip_addr", updatedServiceSpec.Template.Spec.Containers[0].Env[0].Value)
	require.Regexp(t, `\b(?:\d{1,3}\.){3}\d{1,3}\b`, updatedServiceSpec.Template.Spec.Containers[0].Env[0].Value)

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Equal(t, "ip_addr", configMapData["original_value"])

	err = runner.DeleteFlow(complexPlugin, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, exists := runner.memory[flowUuid]
	require.False(t, exists)
}

func TestRedisPluginTest(t *testing.T) {
	t.Skip("skipping this as this pulls from the internet")
	runner := NewPluginRunner()

	updatedServiceSpec, configMap, err := runner.CreateFlow(redisPlugin, serviceSpec, deploymentSpec, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Check if the deployment spec was updated correctly
	require.Equal(t, "kurtosistech/redis-proxy-overlay:latest", updatedServiceSpec.Template.Spec.Containers[0].Image)

	// Verify the config map
	var configMapData map[string]interface{}
	err = json.Unmarshal([]byte(configMap), &configMapData)
	require.NoError(t, err)
	require.Empty(t, configMapData)

	err = runner.DeleteFlow(complexPlugin, flowUuid, map[string]string{})
	require.NoError(t, err)

	// Verify that the flow UUID was removed from memory
	_, exists := runner.memory[flowUuid]
	require.False(t, exists)
}

func int32Ptr(i int32) *int32 { return &i }