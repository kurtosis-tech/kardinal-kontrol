package api

import (
	"context"
	"log"

	compose "github.com/compose-spec/compose-go/types"
	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"

	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/engine/template"
	"kardinal.kontrol-service/topology"
	"kardinal.kontrol-service/types"
)

// TODO:find a better way to find the frontend
const (
	frontendServiceName = "voting-app-ui"
	defaultNamespace    = "default"
)

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

func (sv *Server) RegisterExternalAndInternalApi(router api.EchoRouter) {
	externalHandlers := api.NewStrictHandler(sv, nil)
	internalHandlers := managerapi.NewStrictHandler(sv, nil)

	api.RegisterHandlers(router, externalHandlers)
	managerapi.RegisterHandlers(router, internalHandlers)
}

func (sv *Server) PostDeploy(ctx context.Context, request api.PostDeployRequestObject) (api.PostDeployResponseObject, error) {
	log.Printf("Deploying prod cluster")
	project := *request.Body.DockerCompose
	err := applyProdOnlyFlow(sv, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostFlowDelete(ctx context.Context, request api.PostFlowDeleteRequestObject) (api.PostFlowDeleteResponseObject, error) {
	log.Printf("Deleting dev flow")
	project := *request.Body.DockerCompose
	err := applyProdOnlyFlow(sv, project)
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

	project := *request.Body.DockerCompose
	err := applyProdDevFlow(sv, project, serviceName, imageLocator)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTopology(ctx context.Context, request api.GetTopologyRequestObject) (api.GetTopologyResponseObject, error) {
	namespaceParam := request.Params.Namespace
	namespace := defaultNamespace

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if cluster, found := sv.composes[namespace]; found {
		topo := topology.ClusterTopology(&cluster)
		return api.GetTopology200JSONResponse(*topo), nil
	}

	return nil, nil
}

func (sv *Server) GetClusterResources(ctx context.Context, request managerapi.GetClusterResourcesRequestObject) (managerapi.GetClusterResourcesResponseObject, error) {
	log.Printf("Getting cluster resources")
	namespaceParam := request.Params.Namespace
	namespace := defaultNamespace

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if cluster, found := sv.composes[namespace]; found {
		clusterResources := template.RenderClusterResources(cluster)
		managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)

		return managerapi.GetClusterResources200JSONResponse(managerAPIClusterResources), nil
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, project []compose.ServiceConfig) error {
	cluster, err := engine.GenerateProdOnlyCluster(project)
	if err != nil {
		return err
	}

	sv.composes["default"] = *cluster
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, project []compose.ServiceConfig, devServiceName string, devImage string) error {
	cluster, err := engine.GenerateProdDevCluster(project, devServiceName, devImage)
	if err != nil {
		return err
	}

	sv.composes["default"] = *cluster
	return nil
}

func newManagerAPIClusterResources(clusterResources types.ClusterResources) managerapitypes.ClusterResources {
	return managerapitypes.ClusterResources{
		Deployments:      &clusterResources.Deployments,
		Services:         &clusterResources.Services,
		VirtualServices:  &clusterResources.VirtualServices,
		DestinationRules: &clusterResources.DestinationRules,
		Gateway:          &clusterResources.Gateway,
	}
}
