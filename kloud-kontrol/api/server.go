package api

import (
	"context"
	"log"

	api "kardinal/cli-kontrol-api/api/golang/server"
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
	for _, service := range project {
		log.Printf("ContainerName: %s", service.ContainerName)
	}

	resp := "ok"
	return api.PostDevFlow200JSONResponse(resp), nil
}
