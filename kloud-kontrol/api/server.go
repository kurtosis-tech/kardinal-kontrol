package api

import (
	"context"
	"log"

	compose "github.com/compose-spec/compose-go/types"
	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"

	"k8s.io/client-go/rest"
	"kardinal.kloud-kontrol/engine"
	"kardinal.kloud-kontrol/engine/template"
	"kardinal.kloud-kontrol/topology"
	"kardinal.kloud-kontrol/types"
)

// TODO:find a better way to find the frontend
const frontendServiceName = "voting-app-ui"

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	composes map[string]types.Cluster
}

func NewServer() Server {
	return Server{
		composes: make(map[string]types.Cluster),
	}
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
	err := applyProdOnlyFlow(sv, restConn, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostFlowDelete(ctx context.Context, request api.PostFlowDeleteRequestObject) (api.PostFlowDeleteResponseObject, error) {
	log.Printf("Deleting dev flow")
	restConn := engine.ConnectToCluster()
	project := *request.Body.DockerCompose
	err := applyProdOnlyFlow(sv, restConn, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowDelete200JSONResponse(resp), nil
}

func (sv *Server) PostFlowCreate(ctx context.Context, request api.PostFlowCreateRequestObject) (api.PostFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	restConn := engine.ConnectToCluster()
	project := *request.Body.DockerCompose
	err := applyProdDevFlow(sv, restConn, project, serviceName, imageLocator)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTopology(ctx context.Context, request api.GetTopologyRequestObject) (api.GetTopologyResponseObject, error) {
	namespaceParam := request.Params.Namespace
	namespace := "default"

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if cluster, found := sv.composes[namespace]; found {
		topo := topology.ClusterTopology(&cluster)
		return api.GetTopology200JSONResponse(*topo), nil
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, restConn *rest.Config, project []compose.ServiceConfig) error {
	cluster, err := engine.GenerateProdOnlyCluster(project)
	if err != nil {
		return err
	}

	sv.composes["default"] = *cluster
	clusterResources := template.RenderClusterResources(*cluster)
	engine.ApplyClusterResources(restConn, &clusterResources)
	engine.CleanUpClusterResources(restConn, &clusterResources)
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, restConn *rest.Config, project []compose.ServiceConfig, devServiceName string, devImage string) error {
	cluster, err := engine.GenerateProdDevCluster(project, devServiceName, devImage)
	if err != nil {
		return err
	}

	sv.composes["default"] = *cluster
	clusterResources := template.RenderClusterResources(*cluster)
	engine.ApplyClusterResources(restConn, &clusterResources)
	engine.CleanUpClusterResources(restConn, &clusterResources)
	return nil
}
