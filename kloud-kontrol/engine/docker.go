package engine

import (
	"errors"
	"fmt"
	"log"

	compose "github.com/compose-spec/compose-go/types"

	"github.com/samber/lo"
	"kardinal.kloud-kontrol/types"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

func GenerateProdOnlyCluster(project []compose.ServiceConfig) (*types.Cluster, error) {
	serviceSpecs := lo.Map(project, func(service compose.ServiceConfig, _ int) *types.ServiceSpec {
		version := "prod"
		return &types.ServiceSpec{
			Version:    version,
			Name:       service.ContainerName,
			Port:       int32(service.Ports[0].Target),
			TargetPort: int32(service.Ports[0].Target),
			Config:     service,
		}
	})

	frontendService := lo.Filter(serviceSpecs, func(service *types.ServiceSpec, _ int) bool { return service.Name == frontendServiceName })
	if len(frontendService) == 0 {
		log.Fatalf("Frontend service not found")
		return nil, errors.New("Frontend service not found")
	}

	cluster := types.Cluster{
		Services:            serviceSpecs,
		ServiceDependencies: []*types.ServiceDependency{},
		FrontdoorService:    frontendService,
		TrafficSource: types.Traffic{
			HasMirroring:           false,
			MirrorPercentage:       0,
			MirrorToVersion:        "",
			MirrorExternalHostname: "",
			ExternalHostname:       "prod.app.localhost",
			GatewayName:            "gateway",
		},
		Namespace: types.NamespaceSpec{Name: "prod"},
	}

	return &cluster, nil
}

func GenerateProdDevCluster(project []compose.ServiceConfig, devServiceName string, devImage string) (*types.Cluster, error) {
	var devServiceSpec types.ServiceSpec
	devService, found := lo.Find(project, func(service compose.ServiceConfig) bool { return service.ContainerName == devServiceName })
	if !found {
		log.Fatalf("Frontend service not found")
		return nil, errors.New("Frontend service not found")
	} else {
		devService.Image = devImage
		devService.Environment = lo.MapEntries(devService.Environment, func(key string, value *string) (string, *string) {
			if key == "REDIS" {
				proxyUrl := "kardinal-db-sidecar"
				return key, &proxyUrl
			}
			return key, value
		})
		version := "dev"
		devServiceSpec = types.ServiceSpec{
			Version:    version,
			Name:       devService.ContainerName,
			Port:       int32(devService.Ports[0].Target),
			TargetPort: int32(devService.Ports[0].Target),
			Config:     devService,
		}
	}

	serviceSpecsDev := lo.Map(project, func(service compose.ServiceConfig, _ int) *types.ServiceSpec {
		version := "prod"
		return &types.ServiceSpec{
			Version:    version,
			Name:       service.ContainerName,
			Port:       int32(service.Ports[0].Target),
			TargetPort: int32(service.Ports[0].Target),
			Config:     service,
		}
	})

	redisPort := int32(6379)
	redisPortStr := fmt.Sprintf("%d", redisPort)
	redisProdAddr := fmt.Sprintf("redis-prod:%d", redisPort)
	redisProxyOverlay := types.ServiceSpec{
		Version:    "dev",
		Name:       "kardinal-db-sidecar",
		Port:       redisPort,
		TargetPort: redisPort,
		Config: compose.ServiceConfig{
			ContainerName: "kardinal-db-sidecar",
			Image:         "kurtosistech/redis-proxy-overlay:latest",
			Environment: compose.MappingWithEquals{
				"REDIS_ADDR": &redisProdAddr,
				"PORT":       &redisPortStr,
			},
			Ports: []compose.ServicePortConfig{{
				Protocol: "tcp",
				Target:   uint32(redisPort),
			}},
		},
	}

	allServiceSpecs := append(serviceSpecsDev, &devServiceSpec, &redisProxyOverlay)

	frontendServiceDev := lo.Filter(allServiceSpecs, func(service *types.ServiceSpec, _ int) bool { return service.Name == frontendServiceName })
	if len(frontendServiceDev) == 0 {
		log.Fatalf("Frontend service not found")
		return nil, errors.New("Frontend service not found")
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
