package topology

import (
	"fmt"
	"testing"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kardinal.kontrol-service/engine"
)

const (
	redisServiceName    = "redis-prod"
	redisServiceVersion = "6.0.8"
	redisServiceID      = "node-1"

	votingAppServiceName    = "voting-app-ui"
	votingAppServiceVersion = "latest"
	votingAppServiceID      = "node-2"
)

func TestServiceConfigsToTopology(t *testing.T) {
	testServiceConfigs := []apitypes.ServiceConfig{}

	// Redis prod service
	allowEmpty := "yes"
	appName := "azure-vote-back"
	serviceName := appName
	containerImage := "bitnami/redis:6.0.8"
	containerName := "redis-prod"
	version := "prod"
	port := int32(6379)
	portStr := fmt.Sprintf("%d", port)
	testServiceConfigs = append(testServiceConfigs, apitypes.ServiceConfig{
		Service: v1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: "",
				Labels: map[string]string{
					"app": appName,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       fmt.Sprintf("tcp-%s", containerName),
						Port:       port,
						Protocol:   v1.ProtocolTCP,
						TargetPort: intstr.FromInt(int(port)),
					},
				},
				Selector: map[string]string{
					"app": appName,
				},
			},
		},
		Deployment: apps.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", containerName, version),
				Namespace: "",
				Labels: map[string]string{
					"app":     appName,
					"version": version,
				},
			},
			Spec: apps.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":     appName,
						"version": version,
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":     appName,
							"version": version,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:            containerName,
								Image:           containerImage,
								ImagePullPolicy: "IfNotPresent",
								Env: []v1.EnvVar{
									v1.EnvVar{
										Name:  "ALLOW_EMPTY_PASSWORD",
										Value: allowEmpty,
									},
									v1.EnvVar{
										Name:  "REDIS_PORT_NUMBER",
										Value: portStr,
									},
								},
								Ports: []v1.ContainerPort{
									v1.ContainerPort{
										Name:          fmt.Sprintf("tcp-%d", port),
										ContainerPort: port,
										Protocol:      v1.ProtocolTCP,
									},
								},
							},
						},
					},
				},
			},
		},
	})

	// Voting app UI
	version = "prod"
	appName = "azure-vote-front"
	serviceName = appName
	containerImage = "voting-app-ui"
	containerName = "voting-app-ui"
	port = int32(80)
	testServiceConfigs = append(testServiceConfigs, apitypes.ServiceConfig{
		Service: v1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: "",
				Labels: map[string]string{
					"app": appName,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       fmt.Sprintf("tcp-%s", containerName),
						Port:       port,
						Protocol:   v1.ProtocolTCP,
						TargetPort: intstr.FromInt(int(port)),
					},
				},
				Selector: map[string]string{
					"app": appName,
				},
			},
		},
		Deployment: apps.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", containerName, version),
				Namespace: "",
				Labels: map[string]string{
					"app":     appName,
					"version": version,
				},
			},
			Spec: apps.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":     appName,
						"version": version,
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":     appName,
							"version": version,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:            containerName,
								Image:           containerImage,
								ImagePullPolicy: "IfNotPresent",
								Ports: []v1.ContainerPort{
									v1.ContainerPort{
										Name:          fmt.Sprintf("tcp-%d", port),
										ContainerPort: port,
										Protocol:      v1.ProtocolTCP,
									},
								},
							},
						},
					},
				},
			},
		},
	})

	cluster, err := engine.GenerateProdOnlyCluster(testServiceConfigs)
	if err != nil {
		t.Errorf("Error generating cluster: %s", err)
	}
	topo := ClusterTopology(cluster)
	require.NotNil(t, topo)
}
