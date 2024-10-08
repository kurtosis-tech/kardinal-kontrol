package topology

import (
	"fmt"
	"testing"

	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

func TestServiceConfigsToTopology(t *testing.T) {
	testServiceConfigs := []apitypes.ServiceConfig{}
	testDeploymentConfigs := []apitypes.DeploymentConfig{}

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
	})

	testDeploymentConfigs = append(testDeploymentConfigs, apitypes.DeploymentConfig{
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
										Name:          fmt.Sprintf("tcp-%s", portStr),
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
				Annotations: map[string]string{
					"kardinal.dev.service/dependencies": "azure-vote-back:tcp-redis-prod",
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

	testDeploymentConfigs = append(testDeploymentConfigs, apitypes.DeploymentConfig{
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

	testIngressConfigs := []apitypes.IngressConfig{
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
							Host: "app.kardinal.dev",
							IngressRuleValue: net.IngressRuleValue{
								HTTP: &net.HTTPIngressRuleValue{
									Paths: []net.HTTPIngressPath{
										{
											Path: "/",
											Backend: net.IngressBackend{
												Service: &net.IngressServiceBackend{
													Name: "azure-vote-front",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Status: net.IngressStatus{},
			},
		},
	}

	testGatewayConfigs := []apitypes.GatewayConfig{}
	testRouteConfigs := []apitypes.RouteConfig{}
	testStatefulSetConfigs := []apitypes.StatefulSetConfig{}
	clusterTopology, err := engine.GenerateProdOnlyCluster("prod", testServiceConfigs, testDeploymentConfigs, testStatefulSetConfigs, testIngressConfigs, testGatewayConfigs, testRouteConfigs, "prod")
	if err != nil {
		t.Errorf("Error generating cluster: %s", err)
		return
	}
	clusterTopologyFlowA := deepcopy.Copy(*clusterTopology).(resolved.ClusterTopology)
	flowID := "A"
	clusterTopologyFlowA.FlowID = flowID
	for _, service := range clusterTopologyFlowA.Services {
		service.Version = flowID
		if service == nil {
			panic("Service is nil")
		}
		if len(service.WorkloadSpec.GetTemplateSpec().Containers) == 0 {
			panic("DeploymentSpec is empty: " + service.ServiceID)
		}
		image := service.WorkloadSpec.GetTemplateSpec().Containers[0].Image
		service.WorkloadSpec.GetTemplateSpec().Containers[0].Image = fmt.Sprintf("%s.a", image)
	}

	clusterTopologyFlowB := deepcopy.Copy(*clusterTopology).(resolved.ClusterTopology)
	flowID = "B"
	clusterTopologyFlowB.FlowID = flowID
	for _, service := range clusterTopologyFlowB.Services {
		service.Version = flowID
		image := service.WorkloadSpec.GetTemplateSpec().Containers[0].Image
		service.WorkloadSpec.GetTemplateSpec().Containers[0].Image = fmt.Sprintf("%s.b", image)
	}
	allFlows := []resolved.ClusterTopology{}
	allFlows = append(allFlows, clusterTopologyFlowA, clusterTopologyFlowB)
	topo := ClusterTopology(clusterTopology, &allFlows)
	require.NotNil(t, topo)

	expectedAzureVoteBackImageProd := "bitnami/redis:6.0.8"
	expectedAzureVoteBackImageA := "bitnami/redis:6.0.8.a"
	expectedAzureVoteBackImageB := "bitnami/redis:6.0.8.b"
	expectedAzurVoteFrontImageProd := "voting-app-ui"
	expectedAzurVoteFrontImageA := "voting-app-ui.a"
	expectedAzurVoteFrontImageB := "voting-app-ui.b"

	require.Equal(t,
		[]apiTypes.Node{
			{
				Id:    "azure-vote-back",
				Label: "azure-vote-back",
				Type:  apiTypes.Service,
				Versions: &[]apiTypes.NodeVersion{
					{
						FlowId:     "prod",
						ImageTag:   &expectedAzureVoteBackImageProd,
						IsBaseline: true,
					},
					{
						FlowId:     "A",
						ImageTag:   &expectedAzureVoteBackImageA,
						IsBaseline: false,
					},
					{
						FlowId:     "B",
						ImageTag:   &expectedAzureVoteBackImageB,
						IsBaseline: false,
					},
				},
			},
			{
				Id:    "azure-vote-front",
				Label: "azure-vote-front",
				Type:  apiTypes.Service,
				Versions: &[]apiTypes.NodeVersion{
					{
						FlowId:     "prod",
						ImageTag:   &expectedAzurVoteFrontImageProd,
						IsBaseline: true,
					},
					{
						FlowId:     "A",
						ImageTag:   &expectedAzurVoteFrontImageA,
						IsBaseline: false,
					},
					{
						FlowId:     "B",
						ImageTag:   &expectedAzurVoteFrontImageB,
						IsBaseline: false,
					},
				},
			},
			{
				Id:       "kontrol-ingress",
				Label:    "kontrol-ingress",
				Type:     apiTypes.Gateway,
				Versions: &[]apiTypes.NodeVersion{},
			},
		},
		topo.Nodes)

	require.Equal(t,
		[]apiTypes.Edge{
			{
				Source: "azure-vote-front",
				Target: "azure-vote-back",
			},
			{
				Source: "kontrol-ingress",
				Target: "azure-vote-front",
			},
		},
		topo.Edges,
	)
}
