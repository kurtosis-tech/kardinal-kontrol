package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	api "kardinal/cli-kontrol-api/api/golang/server"

	compose "github.com/compose-spec/compose-go/types"
	"gopkg.in/yaml.v2"
	"kardinal.kloud-kontrol/engine"
	"kardinal.kloud-kontrol/engine/template"
	"kardinal.kloud-kontrol/types"

	"github.com/samber/lo"
)

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

// (POST /dev-flow)
func (Server) PostDevFlow(ctx context.Context, request api.PostDevFlowRequestObject) (api.PostDevFlowResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	project := *request.Body.DockerCompose
	lo.ForEach(project, func(service compose.ServiceConfig, _ int) {
		serviceStr, _ := yaml.Marshal(service)
		fmt.Println(string(serviceStr))
	})

	serviceSpecs := lo.Map(project, func(service compose.ServiceConfig, _ int) types.ServiceSpec {
		return types.ServiceSpec{
			Version:    "prod",
			Name:       service.ContainerName,
			Port:       int32(service.Ports[0].Target),
			TargetPort: int32(service.Ports[0].Target),
			Config:     service,
		}
	})

	frontendService, found := lo.Find(serviceSpecs, func(service types.ServiceSpec) bool { return service.Name == "voting-app-ui" })
	if !found {
		log.Fatalf("Frontend service not found")
		return nil, errors.New("Frontend service not found")
	}

	cluster := types.Cluster{
		Services:            serviceSpecs,
		ServiceDependencies: []types.ServiceDependency{},
		FrontdoorService:    &frontendService,
		TrafficSource: types.Traffic{
			HasMirroring:     false,
			MirrorPercentage: 0,
			MirrorToVersion:  "",
			ExternalHostname: "prod.app.localhost",
			GatewayName:      "gateway",
		},
		Namespace: types.NamespaceSpec{Name: "prod"},
	}

	cluserResources := template.RenderClusterResources(cluster)
	engine.ApplyClusterResources(&cluserResources)

	resp := "ok"
	return api.PostDevFlow200JSONResponse(resp), nil
}
