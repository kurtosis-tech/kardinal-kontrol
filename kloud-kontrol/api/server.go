package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	compose "github.com/compose-spec/compose-go/types"
	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	apiTypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"

	"k8s.io/client-go/rest"
	"kardinal.kloud-kontrol/engine"
	"kardinal.kloud-kontrol/engine/template"
	"kardinal.kloud-kontrol/topology"
	"kardinal.kloud-kontrol/types"

	"github.com/samber/lo"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	composes map[string][]compose.ServiceConfig
}

func NewServer() Server {
	return Server{}
}

func RegisterHandlers(router api.EchoRouter, si api.ServerInterface) {
	api.RegisterHandlers(router, si)
}

func NewStrictHandler(si api.StrictServerInterface) api.ServerInterface {
	return api.NewStrictHandler(si, nil)
}

func (sv *Server) PostDeploy(ctx context.Context, request api.PostDeployRequestObject) (api.PostDeployResponseObject, error) {
	log.Printf("Deploying prod cluster")
	restConn := engine.ConnectToCluster()
	project := *request.Body.DockerCompose
	err := applyProdOnlyFlow(restConn, project)
	if err != nil {
		return nil, err
	}
	sv.composes["default"] = project
	resp := "ok"
	return api.PostDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostFlowDelete(ctx context.Context, request api.PostFlowDeleteRequestObject) (api.PostFlowDeleteResponseObject, error) {
	log.Printf("Deleting dev flow")
	restConn := engine.ConnectToCluster()
	project := *request.Body.DockerCompose
	err := applyProdOnlyFlow(restConn, project)
	if err != nil {
		return nil, err
	}
	sv.composes["default"] = project
	resp := "ok"
	return api.PostFlowDelete200JSONResponse(resp), nil
}

func (sv *Server) PostFlowCreate(ctx context.Context, request api.PostFlowCreateRequestObject) (api.PostFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	restConn := engine.ConnectToCluster()
	project := *request.Body.DockerCompose
	err := applyProdDevFlow(restConn, project, serviceName, imageLocator)
	if err != nil {
		return nil, err
	}
	sv.composes["default"] = project
	resp := "ok"
	return api.PostFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTopology(ctx context.Context, request api.GetTopologyRequestObject) (api.GetTopologyResponseObject, error) {
	namespaceParam := request.Params.Namespace
	namespace := "default"

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if targetCompose, found := sv.composes[namespace]; found {
		topo := topology.ComposeToTopology(&targetCompose)
		return api.GetTopology200JSONResponse(*topo), nil
	}

	redisServiceName := "redis-prod"
	redisServiceVersion := "bitnami/redis:6.0.8"
	redisServiceID := "azure-vote-back"

	votingAppServiceName := "voting-app-ui"
	votingAppServiceVersion := "voting-app-ui"
	votingAppServiceID := "azure-vote-front"

	topo := apiTypes.Topology{
		Graph: &apiTypes.Graph{
			Nodes: &[]apiTypes.Node{
				{
					Id:             &redisServiceID,
					ServiceName:    &redisServiceName,
					ServiceVersion: &redisServiceVersion,
					TalksTo:        nil,
				},
				{
					Id:             &votingAppServiceID,
					ServiceName:    &votingAppServiceName,
					ServiceVersion: &votingAppServiceVersion,
					TalksTo:        &[]string{redisServiceID},
				},
			},
		},
	}

	log.Printf("Received %v as namespace", namespace)
	return api.GetTopology200JSONResponse(topo), nil
}

// ============================================================================================================
func applyProdOnlyFlow(restConn *rest.Config, project []compose.ServiceConfig) error {
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
	return nil
}
