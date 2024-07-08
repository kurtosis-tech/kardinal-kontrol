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

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	clusterByTenant map[string]types.Cluster
}

func NewServer() Server {
	return Server{
		clusterByTenant: make(map[string]types.Cluster),
	}
}

func (sv *Server) RegisterExternalAndInternalApi(router api.EchoRouter) {
	externalHandlers := api.NewStrictHandler(sv, nil)
	internalHandlers := managerapi.NewStrictHandler(sv, nil)

	api.RegisterHandlers(router, externalHandlers)
	managerapi.RegisterHandlers(router, internalHandlers)
}

func (sv *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	resp := "ok"
	return api.GetHealth200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidDeploy(_ context.Context, request api.PostTenantUuidDeployRequestObject) (api.PostTenantUuidDeployResponseObject, error) {
	log.Printf("Deploying prod cluster")
	project := *request.Body.DockerCompose

	err := applyProdOnlyFlow(sv, request.Uuid, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidFlowDelete(_ context.Context, request api.PostTenantUuidFlowDeleteRequestObject) (api.PostTenantUuidFlowDeleteResponseObject, error) {
	log.Printf("Deleting dev flow")
	project := *request.Body.DockerCompose

	err := applyProdOnlyFlow(sv, request.Uuid, project)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidFlowDelete200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidFlowCreate(_ context.Context, request api.PostTenantUuidFlowCreateRequestObject) (api.PostTenantUuidFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	log.Printf("Starting new dev flow for service %v on image %v", serviceName, imageLocator)

	project := *request.Body.DockerCompose

	err := applyProdDevFlow(sv, request.Uuid, project, serviceName, imageLocator)
	if err != nil {
		log.Printf("an error occured while updating dev flow. error was \n: '%v'", err.Error())
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	log.Printf("Getting topology for tenant '%s'", request.Uuid)

	if cluster, found := sv.clusterByTenant[request.Uuid]; found {
		topo := topology.ClusterTopology(&cluster)
		return api.GetTenantUuidTopology200JSONResponse(*topo), nil
	}

	return nil, nil
}

func (sv *Server) GetTenantUuidClusterResources(_ context.Context, request managerapi.GetTenantUuidClusterResourcesRequestObject) (managerapi.GetTenantUuidClusterResourcesResponseObject, error) {
	log.Printf("Getting cluster resources for tenant '%s'", request.Uuid)

	if cluster, found := sv.clusterByTenant[request.Uuid]; found {
		clusterResources := template.RenderClusterResources(cluster)
		managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
		return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, project []compose.ServiceConfig) error {
	cluster, err := engine.GenerateProdOnlyCluster(project)
	if err != nil {
		return err
	}

	sv.clusterByTenant[tenantUuidStr] = *cluster
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, project []compose.ServiceConfig, devServiceName string, devImage string) error {
	cluster, err := engine.GenerateProdDevCluster(project, devServiceName, devImage)
	if err != nil {
		return err
	}

	sv.clusterByTenant[tenantUuidStr] = *cluster
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
