package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	api "kardinal/cli-kontrol-api/api/golang/server"

	compose "github.com/compose-spec/compose-go/types"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"kardinal.kloud-kontrol/engine"
	"kardinal.kloud-kontrol/engine/template"
	"kardinal.kloud-kontrol/types"

	"github.com/samber/lo"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func RegisterHandlers(router api.EchoRouter, si api.ServerInterface) {
	api.RegisterHandlers(router, si)
}

func NewStrictHandler(si api.StrictServerInterface) api.ServerInterface {
	return api.NewStrictHandler(si, nil)
}

func (Server) PostDeploy(ctx context.Context, request api.PostDeployRequestObject) (api.PostDeployResponseObject, error) {
	restConn := engine.ConnectToCluster()
	err := applyProdOnlyFlow(restConn, *request.Body.DockerCompose)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostDeploy200JSONResponse(resp), nil
}

func (Server) PostFlowDelete(ctx context.Context, request api.PostFlowDeleteRequestObject) (api.PostFlowDeleteResponseObject, error) {
	restConn := engine.ConnectToCluster()
	err := applyProdOnlyFlow(restConn, *request.Body.DockerCompose)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowDelete200JSONResponse(resp), nil
}

// (POST /dev-flow)
func (Server) PostFlowCreate(ctx context.Context, request api.PostFlowCreateRequestObject) (api.PostFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	restConn := engine.ConnectToCluster()
	err := applyProdDevFlow(restConn, *request.Body.DockerCompose, serviceName, imageLocator)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowCreate200JSONResponse(resp), nil
}

// ============================================================================================================
func applyProdOnlyFlow(restConn *rest.Config, project []compose.ServiceConfig) error {
	lo.ForEach(project, func(service compose.ServiceConfig, _ int) {
		serviceStr, _ := yaml.Marshal(service)
		fmt.Println(string(serviceStr))
	})

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
		return errors.New("Frontend service not found")
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

	clusterResources := template.RenderClusterResources(cluster)
	engine.ApplyClusterResources(restConn, &clusterResources)
	engine.CleanUpClusterResources(restConn, &clusterResources)
	return nil
}

// ============================================================================================================
func applyProdDevFlow(restConn *rest.Config, project []compose.ServiceConfig, devServiceName string, devImage string) error {
	var devServiceSpec types.ServiceSpec
	devService, found := lo.Find(project, func(service compose.ServiceConfig) bool { return service.ContainerName == devServiceName })
	if !found {
		log.Fatalf("Frontend service not found")
		return errors.New("Frontend service not found")
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
		return errors.New("Frontend service not found")
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

	clusterDevResources := template.RenderClusterResources(clusterDev)
	engine.ApplyClusterResources(restConn, &clusterDevResources)
	engine.CleanUpClusterResources(restConn, &clusterDevResources)
	fmt.Println(clusterDevResources)
	return nil
}
