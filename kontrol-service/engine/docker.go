package engine

import (
	"errors"
	"fmt"
	"log"
	"strings"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mohae/deepcopy"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

func GenerateProdOnlyCluster(serviceConfigs []apitypes.ServiceConfig) (*resolved.ClusterTopology, error) {

	version := "prod"
	clusterTopology, err := generateClusterTopology(serviceConfigs, version)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occured generating the cluster topology from the service configs")
	}

	return clusterTopology, nil
}

func GenerateProdDevCluster(serviceConfigs []apitypes.ServiceConfig, devServiceName string, devImage string) (*types.Cluster, error) {
	var devServiceSpec types.ServiceSpec
	var updateErr error
	serviceConfig, found := lo.Find(serviceConfigs, func(serviceConfig apitypes.ServiceConfig) bool {
		return serviceConfig.Deployment.Spec.Template.Spec.Containers[0].Name == devServiceName
	})
	if !found {
		log.Fatalf("Dev service %s not found", devServiceName)
		return nil, errors.New("Dev service not found")
	} else {
		devServiceConfig := deepcopy.Copy(serviceConfig).(apitypes.ServiceConfig)
		devServiceContainer := devServiceConfig.Deployment.Spec.Template.Spec.Containers[0]
		devServiceContainer.Image = devImage

		neonApiKey := ""
		projectID := ""
		mainBranchId := ""
		for index, envVar := range devServiceContainer.Env {
			switch envVar.Name {
			case "NEON_API_KEY":
				neonApiKey = envVar.Value
			case "NEON_PROJECT_ID":
				projectID = envVar.Value
			case "NEON_MAIN_BRANCH_ID":
				mainBranchId = envVar.Value
			case "REDIS":
				proxyUrl := "kardinal-db-sidecar"
				devServiceContainer.Env[index].Value = proxyUrl
			case "POSTGRES":
				if neonApiKey == "" || projectID == "" || mainBranchId == "" {
					log.Println("Saw postgres env var but at least one of NEON_API_KEY, NEON_PROJECT_ID, NEON_MAIN_BRANCH were empty")
					break
				}

				newHost, err := createNeonBranch(neonApiKey, projectID, mainBranchId)
				if err != nil {
					updateErr = fmt.Errorf("error creating Neon branch: %v", err)
					log.Printf("an error occurred while creating neon branch. Error was:\n '%v'", updateErr.Error())
					break
				}

				updatedConnString, err := updateConnectionString(envVar.Value, newHost)
				if err != nil {
					updateErr = fmt.Errorf("error updating connection string: %v", err)
					log.Printf("an error occurred while creating updating the connection string. Error was:\n '%v'", updateErr.Error())
					break
				}

				log.Printf("neon branching succeeded, new connection string with host '%v' will be used", newHost)

				devServiceContainer.Env[index].Value = updatedConnString
			}
		}

		devServiceConfig.Deployment.Spec.Template.Spec.Containers[0] = devServiceContainer

		version := "dev"
		devServiceSpec = types.ServiceSpec{
			Version:    version,
			Name:       devServiceContainer.Name,
			Port:       int32(devServiceConfig.Service.Spec.Ports[0].TargetPort.IntValue()),
			TargetPort: int32(devServiceConfig.Service.Spec.Ports[0].TargetPort.IntValue()),
			Config:     devServiceConfig,
		}
	}

	if updateErr != nil {
		log.Printf("an error occurred while updating the postgres string. Error was :\n '%v'", updateErr.Error())
		return nil, updateErr
	}

	serviceSpecsDev := lo.Map(serviceConfigs, func(serviceConfig apitypes.ServiceConfig, _ int) *types.ServiceSpec {
		version := "prod"
		return &types.ServiceSpec{
			Version:    version,
			Name:       serviceConfig.Deployment.Spec.Template.Spec.Containers[0].Name,
			Port:       int32(serviceConfig.Service.Spec.Ports[0].TargetPort.IntValue()),
			TargetPort: int32(serviceConfig.Service.Spec.Ports[0].TargetPort.IntValue()),
			Config:     serviceConfig,
		}
	})

	redisPort := int32(6379)
	redisPortStr := fmt.Sprintf("%d", redisPort)
	redisProdAddr := fmt.Sprintf("redis-prod:%d", redisPort)
	appName := "kardinal-db-sidecar"
	serviceName := appName
	containerName := appName
	containerImage := "kurtosistech/redis-proxy-overlay:latest"
	version := "dev"
	redisProxyOverlay := types.ServiceSpec{
		Version:    version,
		Name:       serviceName,
		Port:       redisPort,
		TargetPort: redisPort,
		Config: apitypes.ServiceConfig{
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
							Port:       redisPort,
							Protocol:   v1.ProtocolTCP,
							TargetPort: intstr.FromInt(int(redisPort)),
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
										{
											Name:  "REDIS_ADDR",
											Value: redisProdAddr,
										},
										{
											Name:  "PORT",
											Value: redisPortStr,
										},
									},
									Ports: []v1.ContainerPort{
										{
											Name:          fmt.Sprintf("tcp-%d", redisPort),
											ContainerPort: redisPort,
											Protocol:      v1.ProtocolTCP,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	allServiceSpecs := append(serviceSpecsDev, &devServiceSpec, &redisProxyOverlay)

	frontendServiceDev := lo.Filter(allServiceSpecs, func(service *types.ServiceSpec, _ int) bool { return service.Name == frontendServiceName })
	if len(frontendServiceDev) == 0 {
		log.Println("Frontend service not found")
	}

	clusterDev := types.Cluster{
		Services:            allServiceSpecs,
		ServiceDependencies: []*types.ServiceDependency{},
		FrontdoorService:    frontendServiceDev,
		TrafficSource: types.Traffic{
			HasMirroring:           true,
			MirrorPercentage:       10,
			MirrorToVersion:        "dev",
			MirrorExternalHostname: "dev.app.localhost",
			ExternalHostname:       "prod.app.localhost",
			GatewayName:            "gateway",
		},
		Namespace: types.NamespaceSpec{Name: "prod"},
	}
	return &clusterDev, nil
}

func generateClusterTopology(serviceConfigs []apitypes.ServiceConfig, version string) (*resolved.ClusterTopology, error) {
	clusterTopology := resolved.ClusterTopology{}

	clusterTopologyServices := []resolved.Service{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		deployment := serviceConfig.Deployment
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		// Ingress?
		isIngress, ok := serviceAnnotations["kardinal.dev.service/ingress"]
		if ok && isIngress == "true" {
			clusterTopology.Ingress = resolved.Ingress{
				IngressID:   service.ObjectMeta.Name,
				ServiceSpec: &service.Spec,
			}
			host, ok := serviceAnnotations["kardinal.dev.service/host"]
			if ok {
				clusterTopology.Ingress.IngressRules = []*net.IngressRule{
					&net.IngressRule{
						Host: host,
					},
				}
			}
			continue
		}

		// Service
		clusterTopologyService := resolved.Service{
			ServiceID:      service.GetObjectMeta().GetName(),
			Version:        version,
			ServiceSpec:    &service.Spec,
			DeploymentSpec: &deployment.Spec,
		}
		isStateful, ok := serviceAnnotations["kardinal.dev.service/stateful"]
		if ok && isStateful == "true" {
			clusterTopologyService.IsStateful = true
		}
		isExternal, ok := serviceAnnotations["kardinal.dev.service/external"]
		if ok && isExternal == "true" {
			clusterTopologyService.IsExternal = true
		}

		// Service plugin?
		plugins, ok := serviceAnnotations["kardinal.dev.service/plugins"]
		if ok {
			var statefulPlugins []resolved.StatefulPlugin
			err := yaml.Unmarshal([]byte(plugins), &statefulPlugins)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred parsing the plugins for service %s", service.GetObjectMeta().GetName())
			}
			serviceStatefulPlugins := make([]*resolved.StatefulPlugin, len(statefulPlugins))
			for index := range statefulPlugins {
				serviceStatefulPlugins[index] = &statefulPlugins[index]
			}
			clusterTopologyService.StatefulPlugins = serviceStatefulPlugins
		}

		clusterTopologyServices = append(clusterTopologyServices, clusterTopologyService)
	}

	clusterTopology.Services = clusterTopologyServices

	clusterTopologyServiceDependencies := []resolved.ServiceDependency{}
	for _, serviceConfig := range serviceConfigs {
		service := serviceConfig.Service
		serviceAnnotations := service.GetObjectMeta().GetAnnotations()

		if service.GetObjectMeta().GetName() == clusterTopology.Ingress.IngressID {
			continue
		}

		clusterTopologyService, err := clusterTopology.GetService(service.GetObjectMeta().GetName())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred finding service %s in the list of services", service.GetObjectMeta().GetName())
		}

		// Service dependencies?
		deps, ok := serviceAnnotations["kardinal.dev.service/dependencies"]
		if ok {
			serviceAndPorts := strings.Split(deps, ",")
			for _, serviceAndPort := range serviceAndPorts {
				serviceAndPortParts := strings.Split(serviceAndPort, ":")
				depService, depServicePort, err := clusterTopology.GetServiceAndPort(serviceAndPortParts[0], serviceAndPortParts[1])
				if err != nil {
					return nil, stacktrace.Propagate(err, "An error occurred finding the service dependency for service %s and port %s", serviceAndPortParts[0], serviceAndPortParts[1])
				}

				serviceDependency := resolved.ServiceDependency{
					Service:          clusterTopologyService,
					DependsOnService: depService,
					DependencyPort:   depServicePort,
				}

				clusterTopologyServiceDependencies = append(clusterTopologyServiceDependencies, serviceDependency)
			}
		}
	}

	clusterTopology.ServiceDependecies = clusterTopologyServiceDependencies

	return &clusterTopology, nil
}
