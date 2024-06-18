package template

import (
	"istio.io/api/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Define the Service
var redisService = &v1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "redis-prod",
		Namespace: "voting-app",
		Labels: map[string]string{
			"app":     "redis-prod",
			"version": "v1",
		},
	},
	Spec: v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:       "tcp-redis",
				Port:       6379,
				Protocol:   v1.ProtocolTCP,
				TargetPort: intstr.FromInt(6379),
			},
		},
		Selector: map[string]string{
			"app": "redis-prod",
		},
	},
}

var vol25pct = intstr.FromString("25%")

// Define the Deployment
var redisDeployment = &apps.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "redis-prod-v1",
		Namespace: "voting-app",
		Labels: map[string]string{
			"app":     "redis-prod",
			"version": "v1",
		},
	},
	Spec: apps.DeploymentSpec{
		Replicas: int32Ptr(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":     "redis-prod",
				"version": "v1",
			},
		},
		Strategy: apps.DeploymentStrategy{
			Type: apps.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &apps.RollingUpdateDeployment{
				MaxSurge:       &vol25pct,
				MaxUnavailable: &vol25pct,
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "true",
				},
				Labels: map[string]string{
					"app":     "redis-prod",
					"version": "v1",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "redis-prod",
						Image: "bitnami/redis:6.0.8",
						Env: []v1.EnvVar{
							{
								Name:  "ALLOW_EMPTY_PASSWORD",
								Value: "yes",
							},
							{
								Name:  "REDIS_PORT_NUMBER",
								Value: "6379",
							},
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("250m"),
								v1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
						Ports: []v1.ContainerPort{
							{
								ContainerPort: 6379,
								Name:          "redis",
							},
						},
					},
				},
			},
		},
	},
}

// Define the Gateway
var votingAppGateway = &istioclient.Gateway{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.istio.io/v1alpha3",
		Kind:       "Gateway",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "voting-app",
		Namespace: "voting-app",
		Labels: map[string]string{
			"app":     "voting-app",
			"version": "v1",
		},
	},
	Spec: v1alpha3.Gateway{
		Selector: map[string]string{
			"istio": "ingressgateway",
		},
		Servers: []*v1alpha3.Server{
			{
				Port: &v1alpha3.Port{
					Number:   80,
					Name:     "http",
					Protocol: "HTTP",
				},
				Hosts: []string{
					"voting-app.localhost",
					"dev.voting-app.localhost",
				},
			},
		},
	},
}

// Define the VirtualService
var votingAppVirtualService = &istioclient.VirtualService{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.istio.io/v1alpha3",
		Kind:       "VirtualService",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "voting-app-ui",
		Namespace: "voting-app",
	},
	Spec: v1alpha3.VirtualService{
		Hosts: []string{
			"voting-app.localhost",
		},
		Gateways: []string{
			"voting-app",
		},
		Http: []*v1alpha3.HTTPRoute{
			{
				Route: []*v1alpha3.HTTPRouteDestination{
					{
						Destination: &v1alpha3.Destination{
							Host:   "voting-app-ui",
							Subset: "v1",
						},
						Weight: 100,
					},
				},
				Mirror: &v1alpha3.Destination{
					Host:   "voting-app-ui",
					Subset: "v2",
				},
				MirrorPercentage: &v1alpha3.Percent{
					Value: 10,
				},
			},
		},
	},
}

// Define the DestinationRule
var votingAppDestinationRule = &istioclient.DestinationRule{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.istio.io/v1alpha3",
		Kind:       "DestinationRule",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "voting-app-ui",
		Namespace: "voting-app",
		Labels: map[string]string{
			"type": "dev",
		},
	},
	Spec: v1alpha3.DestinationRule{
		Host: "voting-app-ui",
		Subsets: []*v1alpha3.Subset{
			{
				Name: "v1",
				Labels: map[string]string{
					"version": "v1",
				},
			},
			{
				Name: "v2",
				Labels: map[string]string{
					"version": "v2",
				},
			},
		},
	},
}

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
