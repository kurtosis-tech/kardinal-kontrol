package api

import (
	"context"
	compose "github.com/compose-spec/compose-go/types"
	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"log"

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
	composesByTenant map[string]map[string]types.Cluster
}

func NewServer() Server {
	return Server{
		composesByTenant: make(map[string]map[string]types.Cluster),
	}
}

func (sv *Server) RegisterExternalAndInternalApi(router api.EchoRouter) {
	externalHandlers := api.NewStrictHandler(sv, nil)
	internalHandlers := managerapi.NewStrictHandler(sv, nil)

	api.RegisterHandlers(router, externalHandlers)
	managerapi.RegisterHandlers(router, internalHandlers)
}

func (sv *Server) PostDeploy(_ context.Context, request api.PostDeployRequestObject) (api.PostDeployResponseObject, error) {
	log.Printf("Deploying prod cluster")
	project := *request.Body.DockerCompose

	tenantUuidStr := request.Params.Tenant
	if tenantUuidStr == "" {
		return nil, stacktrace.NewError("tenant parameter is required")
	}

	err := applyProdOnlyFlow(sv, tenantUuidStr, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostFlowDelete(_ context.Context, request api.PostFlowDeleteRequestObject) (api.PostFlowDeleteResponseObject, error) {
	log.Printf("Deleting dev flow")
	project := *request.Body.DockerCompose

	tenantUuidStr := request.Params.Tenant
	if tenantUuidStr == "" {
		return nil, stacktrace.NewError("tenant parameter is required")
	}

	err := applyProdOnlyFlow(sv, tenantUuidStr, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowDelete200JSONResponse(resp), nil
}

func (sv *Server) PostFlowCreate(_ context.Context, request api.PostFlowCreateRequestObject) (api.PostFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	project := *request.Body.DockerCompose

	tenantUuidStr := request.Params.Tenant
	if tenantUuidStr == "" {
		return nil, stacktrace.NewError("tenant parameter is required")
	}

	err := applyProdDevFlow(sv, tenantUuidStr, project, serviceName, imageLocator)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTopology(_ context.Context, request api.GetTopologyRequestObject) (api.GetTopologyResponseObject, error) {
	tenantUuidStr := request.Params.Tenant
	if tenantUuidStr == "" {
		return nil, stacktrace.NewError("tenant parameter is required")
	}

	log.Printf("Getting topology for tenant '%s'", tenantUuidStr)

	namespaceParam := request.Params.Namespace
	namespace := defaultNamespace

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if composesByNamespace, found := sv.composesByTenant[tenantUuidStr]; found {
		if cluster, found := composesByNamespace[namespace]; found {
			topo := topology.ClusterTopology(&cluster)
			return api.GetTopology200JSONResponse(*topo), nil
		}
	}

	return nil, nil
}

func (sv *Server) GetClusterResources(_ context.Context, request managerapi.GetClusterResourcesRequestObject) (managerapi.GetClusterResourcesResponseObject, error) {
	tenantUuidStr := request.Params.Tenant
	if tenantUuidStr == "" {
		return nil, stacktrace.NewError("tenant parameter is required")
	}

	log.Printf("Getting cluster resources for tenant '%s'", tenantUuidStr)

	namespaceParam := request.Params.Namespace
	namespace := defaultNamespace

	if namespaceParam != nil && len(*namespaceParam) != 0 {
		namespace = *namespaceParam
	}

	if composesByNamespace, found := sv.composesByTenant[tenantUuidStr]; found {
		if cluster, found := composesByNamespace[namespace]; found {
			clusterResources := template.RenderClusterResources(cluster)
			managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)

			return managerapi.GetClusterResources200JSONResponse(managerAPIClusterResources), nil
		}
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, project []compose.ServiceConfig) error {
	cluster, err := engine.GenerateProdOnlyCluster(project)
	if err != nil {
		return err
	}

	if _, found := sv.composesByTenant[tenantUuidStr]; !found {
		composesByNamespace := make(map[string]types.Cluster)
		sv.composesByTenant[tenantUuidStr] = composesByNamespace
	}

	sv.composesByTenant[tenantUuidStr][defaultNamespace] = *cluster
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, project []compose.ServiceConfig, devServiceName string, devImage string) error {
	cluster, err := engine.GenerateProdDevCluster(project, devServiceName, devImage)
	if err != nil {
		return err
	}

	if _, found := sv.composesByTenant[tenantUuidStr]; !found {
		composesByNamespace := make(map[string]types.Cluster)
		sv.composesByTenant[tenantUuidStr] = composesByNamespace
	}

	sv.composesByTenant[tenantUuidStr][defaultNamespace] = *cluster
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
