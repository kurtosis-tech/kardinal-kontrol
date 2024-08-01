package api

import (
	"context"

	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"
	"github.com/sirupsen/logrus"

	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/engine/template"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/topology"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	clusterByTenant         map[string]types.Cluster
	pluginRunnerByTenant    map[string]plugins.PluginRunner
	clusterTopologyByTenant map[string]resolved.ClusterTopology
}

func NewServer() Server {
	return Server{
		clusterByTenant:         make(map[string]types.Cluster),
		pluginRunnerByTenant:    make(map[string]plugins.PluginRunner),
		clusterTopologyByTenant: make(map[string]resolved.ClusterTopology),
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
	logrus.Infof("deploying prod cluster for tenant '%s'", request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs

	err := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidFlowDelete(_ context.Context, request api.PostTenantUuidFlowDeleteRequestObject) (api.PostTenantUuidFlowDeleteResponseObject, error) {
	logrus.Infof("deleting dev flow for tenant '%s'", request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs

	err := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs)
	if err != nil {
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidFlowDelete200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidFlowCreate(_ context.Context, request api.PostTenantUuidFlowCreateRequestObject) (api.PostTenantUuidFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	logrus.Infof("starting new dev flow for service %v on image %v", serviceName, imageLocator)

	serviceConfigs := *request.Body.ServiceConfigs

	err := applyProdDevFlow(sv, request.Uuid, serviceConfigs, serviceName, imageLocator)
	if err != nil {
		logrus.Errorf("an error occured while updating dev flow. error was \n: '%v'", err.Error())
		return nil, err
	}
	resp := "ok"
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	logrus.Infof("getting topology for tenant '%s'", request.Uuid)

	if clusterTopology, found := sv.clusterTopologyByTenant[request.Uuid]; found {
		topo := topology.ClusterTopology(&clusterTopology)
		return api.GetTenantUuidTopology200JSONResponse(*topo), nil
	}

	return nil, nil
}

func (sv *Server) GetTenantUuidClusterResources(_ context.Context, request managerapi.GetTenantUuidClusterResourcesRequestObject) (managerapi.GetTenantUuidClusterResourcesResponseObject, error) {
	namespace := "prod"

	// TODO - this can be removed?
	if cluster, found := sv.clusterByTenant[request.Uuid]; found {
		clusterResources := template.RenderClusterResources(cluster)
		managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
		return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
	}

	if clusterTopology, found := sv.clusterTopologyByTenant[request.Uuid]; found {
		clusterResources := flow.RenderClusterResources(&clusterTopology, namespace)
		managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
		return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig) error {
	clusterTopology, err := engine.GenerateProdOnlyCluster(serviceConfigs)
	if err != nil {
		return err
	}

	sv.pluginRunnerByTenant[tenantUuidStr] = plugins.PluginRunner{}
	sv.clusterTopologyByTenant[tenantUuidStr] = *clusterTopology
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, devServiceName string, devImage string) error {
	flowID := "dev"

	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)
	prodClusterTopology, err := engine.GenerateProdOnlyCluster(serviceConfigs)
	if err != nil {
		return err
	}

	pluginRunner, found := sv.pluginRunnerByTenant[tenantUuidStr]
	if !found {
		pluginRunner = plugins.PluginRunner{}
	}

	logrus.Debugf("calculating cluster topology overlay for tenant %s on flowID %s", tenantUuidStr, flowID)
	devClusterTopology, err := engine.GenerateProdDevCluster(prodClusterTopology, pluginRunner, flowID, devServiceName, devImage)
	if err != nil {
		return err
	}

	sv.clusterTopologyByTenant[tenantUuidStr] = *devClusterTopology
	return nil
}

func newManagerAPIClusterResources(clusterResources types.ClusterResources) managerapitypes.ClusterResources {
	return managerapitypes.ClusterResources{
		Deployments:           &clusterResources.Deployments,
		Services:              &clusterResources.Services,
		VirtualServices:       &clusterResources.VirtualServices,
		DestinationRules:      &clusterResources.DestinationRules,
		Gateway:               &clusterResources.Gateway,
		EnvoyFilters:          &clusterResources.EnvoyFilters,
		AuthorizationPolicies: &clusterResources.AuthorizationPolicies,
	}
}
