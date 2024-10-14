package test

import (
	"fmt"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetIngressConfigs() []apitypes.IngressConfig {
	return []apitypes.IngressConfig{
		{
			Ingress: net.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Ingress",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "kontrol-ingress",
					Annotations: map[string]string{
						"kardinal.dev.service/ingress": "true",
					},
				},
				Spec: net.IngressSpec{
					Rules: []net.IngressRule{
						{
							Host:             "app.kardinal.dev",
							IngressRuleValue: net.IngressRuleValue{},
						},
					},
				},
				Status: net.IngressStatus{},
			},
		},
	}
}

func GetServiceConfigs() ([]apitypes.ServiceConfig, []apitypes.DeploymentConfig) {
	deploymentConfigs := []apitypes.DeploymentConfig{}

	// Redis prod service
	allowEmpty := "yes"
	appName := "redis-prod"
	serviceName := appName
	containerImage := "bitnami/redis:6.0.8"
	containerName := "redis-prod"
	version := "prod"
	pluginServiceName := "redis-prod-plugin"
	port := int32(6379)
	portStr := fmt.Sprintf("%d", port)

	redisProdServiceConfig := apitypes.ServiceConfig{
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
				Annotations: map[string]string{
					"kardinal.dev.service/stateful": "true",
					"kardinal.dev.service/plugins":  pluginServiceName,
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
	}

	pluginServiceConfig := apitypes.ServiceConfig{
		Service: v1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: pluginServiceName,
				Annotations: map[string]string{
					"kardinal.dev.service/stateful": "true",
					"kardinal.dev.service/plugin-definition": `
- name: github.com/kardinaldev/redis-db-sidecar-plugin:36ed9a4
  type: stateful
  servicename: redis-prod-plugin
  args:
    mode: "pass-through"
`,
				},
			},
		},
	}

	serviceConfigs := []apitypes.ServiceConfig{redisProdServiceConfig, pluginServiceConfig}

	deploymentConfigs = append(deploymentConfigs, apitypes.DeploymentConfig{
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
									{
										Name:  "ALLOW_EMPTY_PASSWORD",
										Value: allowEmpty,
									},
									{
										Name:  "REDIS_PORT_NUMBER",
										Value: portStr,
									},
								},
								Ports: []v1.ContainerPort{
									{
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
	appName = "voting-app-ui"
	serviceName = appName
	containerImage = "voting-app-ui"
	containerName = "voting-app-ui"
	port = int32(80)
	serviceConfigs = append(serviceConfigs, apitypes.ServiceConfig{
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
				Annotations: map[string]string{
					"kardinal.dev.service/dependencies": "redis-prod:tcp-redis-prod",
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
	})

	deploymentConfigs = append(deploymentConfigs, apitypes.DeploymentConfig{
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
									{
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

	// Voting app LB
	version = "prod"
	appName = "voting-app-lb"
	serviceName = appName
	containerImage = "voting-app-lb"
	containerName = "voting-app-lb"
	port = int32(80)
	serviceConfigs = append(serviceConfigs, apitypes.ServiceConfig{
		Service: v1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName,
				Annotations: map[string]string{
					"kardinal.dev.service/ingress": "true",
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeExternalName,
				Ports: []v1.ServicePort{
					{
						Name:       fmt.Sprintf("tcp-%s", containerName),
						Port:       port,
						Protocol:   v1.ProtocolTCP,
						TargetPort: intstr.FromInt(int(port)),
					},
				},
				Selector: map[string]string{
					"app": "voting-app-ui",
				},
			},
		},
	})

	return serviceConfigs, deploymentConfigs
}
