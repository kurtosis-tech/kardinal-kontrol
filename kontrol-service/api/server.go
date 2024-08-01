package api

import (
	"context"
	"fmt"

	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/topology"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	pluginRunnerByTenant        map[string]plugins.PluginRunner
	baseClusterTopologyByTenant map[string]resolved.ClusterTopology
	clusterTopologyByTenantFlow map[string]map[string]resolved.ClusterTopology
}

func NewServer() Server {
	return Server{
		pluginRunnerByTenant:        make(map[string]plugins.PluginRunner),
		baseClusterTopologyByTenant: make(map[string]resolved.ClusterTopology),
		clusterTopologyByTenantFlow: make(map[string]map[string]resolved.ClusterTopology),
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

func (sv *Server) GetTenantUuidFlows(_ context.Context, request api.GetTenantUuidFlowsRequestObject) (api.GetTenantUuidFlowsResponseObject, error) {
	if _, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
			flows := lo.Map(lo.Keys(allFlows), func(value string, _ int) apitypes.DevFlow { return apitypes.DevFlow{&value} })
			return api.GetTenantUuidFlows200JSONResponse(flows), nil
		}
	}
	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: &resourceType, Id: &request.Uuid}
	return api.GetTenantUuidFlows404JSONResponse{missing}, nil
}

func (sv *Server) PostTenantUuidDeploy(_ context.Context, request api.PostTenantUuidDeployRequestObject) (api.PostTenantUuidDeployResponseObject, error) {
	logrus.Infof("deploying prod cluster for tenant '%s'", request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs

	flowId := "prod"
	err := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs, flowId)
	if err != nil {
		return nil, err
	}
	resp := apitypes.DevFlow{DevFlowId: &flowId}
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidFlowFlowIdDelete(_ context.Context, request api.PostTenantUuidFlowFlowIdDeleteRequestObject) (api.PostTenantUuidFlowFlowIdDeleteResponseObject, error) {
	logrus.Infof("deleting dev flow for tenant '%s'", request.Uuid)
	if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
		if _, found := allFlows[request.FlowId]; found {
			logrus.Infof("deleting flow %s", request.FlowId)
			delete(allFlows, request.FlowId)
			return api.PostTenantUuidFlowFlowIdDelete2xxResponse{StatusCode: 200}, nil
		}
		return api.PostTenantUuidFlowFlowIdDelete2xxResponse{StatusCode: 204}, nil
	}
	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: &resourceType, Id: &request.Uuid}
	return api.PostTenantUuidFlowFlowIdDelete404JSONResponse{missing}, nil
}

func (sv *Server) PostTenantUuidFlowCreate(_ context.Context, request api.PostTenantUuidFlowCreateRequestObject) (api.PostTenantUuidFlowCreateResponseObject, error) {
	serviceName := *request.Body.ServiceName
	imageLocator := *request.Body.ImageLocator
	logrus.Infof("starting new dev flow for service %v on image %v", serviceName, imageLocator)

	serviceConfigs := *request.Body.ServiceConfigs

	flowId, err := applyProdDevFlow(sv, request.Uuid, serviceConfigs, serviceName, imageLocator)
	if err != nil {
		logrus.Errorf("an error occured while updating dev flow. error was \n: '%v'", err.Error())
		return nil, err
	}
	resp := apitypes.DevFlow{DevFlowId: flowId}
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	logrus.Infof("getting topology for tenant '%s'", request.Uuid)

	if clusterTopology, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		topo := topology.ClusterTopology(&clusterTopology)
		return api.GetTenantUuidTopology200JSONResponse(*topo), nil
	}

	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: &resourceType, Id: &request.Uuid}
	return api.GetTenantUuidTopology404JSONResponse{missing}, nil
}

func (sv *Server) GetTenantUuidClusterResources(_ context.Context, request managerapi.GetTenantUuidClusterResourcesRequestObject) (managerapi.GetTenantUuidClusterResourcesResponseObject, error) {
	namespace := "prod"

	if clusterTopology, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
			finalTopology := flow.MergeClusterTopologies(clusterTopology, lo.Values(allFlows))
			clusterResources := flow.RenderClusterResources(finalTopology, namespace)
			managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
			return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
		}
	}

	return nil, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, flowID string) error {
	clusterTopology, err := engine.GenerateProdOnlyCluster(flowID, serviceConfigs)
	if err != nil {
		return err
	}

	sv.pluginRunnerByTenant[tenantUuidStr] = plugins.PluginRunner{}
	sv.baseClusterTopologyByTenant[tenantUuidStr] = *clusterTopology
	sv.clusterTopologyByTenantFlow[tenantUuidStr] = make(map[string]resolved.ClusterTopology)
	return nil
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, devServiceName string, devImage string) (*string, error) {
	randId := GetRandFlowID()
	flowID := fmt.Sprintf("dev-%s", randId)

	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)

	prodClusterTopology, found := sv.baseClusterTopologyByTenant[tenantUuidStr]
	if !found {
		return nil, fmt.Errorf("no base cluster topology found for tenant %s, did you deploy the cluster?", tenantUuidStr)
	}

	pluginRunner, found := sv.pluginRunnerByTenant[tenantUuidStr]
	if !found {
		pluginRunner = plugins.PluginRunner{}
	}

	logrus.Debugf("calculating cluster topology overlay for tenant %s on flowID %s", tenantUuidStr, flowID)
	devClusterTopology, err := engine.GenerateProdDevCluster(&prodClusterTopology, pluginRunner, flowID, devServiceName, devImage)
	if err != nil {
		return nil, err
	}

	sv.clusterTopologyByTenantFlow[tenantUuidStr][flowID] = *devClusterTopology
	return &flowID, nil
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
