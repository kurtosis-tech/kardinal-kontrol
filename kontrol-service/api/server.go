package api

import (
	"context"
	"fmt"
	"os"

	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"kardinal.kontrol-service/database"
	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/topology"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
)

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	pluginRunnerByTenant        map[string]*plugins.PluginRunner
	baseClusterTopologyByTenant map[string]resolved.ClusterTopology
	clusterTopologyByTenantFlow map[string]map[string]resolved.ClusterTopology
	db                          *database.Db
}

func NewServer(db *database.Db) Server {
	return Server{
		pluginRunnerByTenant:        make(map[string]*plugins.PluginRunner),
		baseClusterTopologyByTenant: make(map[string]resolved.ClusterTopology),
		clusterTopologyByTenantFlow: make(map[string]map[string]resolved.ClusterTopology),
		db:                          db,
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
	if clusterTopology, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
			finalTopology := flow.MergeClusterTopologies(clusterTopology, lo.Values(allFlows))
			flowHostMapping := finalTopology.GetFlowHostMapping()
			resp := lo.MapToSlice(flowHostMapping, func(flowId string, flowUrls []string) apitypes.Flow {
				return apitypes.Flow{FlowId: flowId, FlowUrls: flowUrls}
			})
			return api.GetTenantUuidFlows200JSONResponse(resp), nil

		}
	}

	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
	return api.GetTenantUuidFlows404JSONResponse{NotFoundJSONResponse: missing}, nil
}

func (sv *Server) PostTenantUuidDeploy(_ context.Context, request api.PostTenantUuidDeployRequestObject) (api.PostTenantUuidDeployResponseObject, error) {
	logrus.Infof("deploying prod cluster for tenant '%s'", request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs

	flowId := "prod"
	err, urls := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs, flowId)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred deploying flow '%v'", flowId)
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.PostTenantUuidDeploy500JSONResponse{errResp}, nil
	}

	resp := apitypes.Flow{FlowId: flowId, FlowUrls: urls}
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) DeleteTenantUuidFlowFlowId(_ context.Context, request api.DeleteTenantUuidFlowFlowIdRequestObject) (api.DeleteTenantUuidFlowFlowIdResponseObject, error) {
	logrus.Infof("deleting dev flow for tenant '%s'", request.Uuid)

	runner, ok := sv.pluginRunnerByTenant[request.Uuid]
	if !ok {
		logrus.Errorf("Plugin Runner for requested tenant not found.")
	}

	if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
		if flowTopology, found := allFlows[request.FlowId]; found {
			logrus.Infof("deleting flow %s", request.FlowId)
			err := flow.DeleteFlow(runner, flowTopology, request.FlowId)
			if err != nil {
				errMsg := fmt.Sprintf("An error occurred deleting flow '%v'", request.FlowId)
				errResp := api.ErrorJSONResponse{
					Error: err.Error(),
					Msg:   &errMsg,
				}
				return api.DeleteTenantUuidFlowFlowId500JSONResponse{errResp}, nil
			}

			delete(allFlows, request.FlowId)
			logrus.Infof("Successfully deleted flow.")
			return api.DeleteTenantUuidFlowFlowId2xxResponse{StatusCode: 200}, nil
		}
		return api.DeleteTenantUuidFlowFlowId2xxResponse{StatusCode: 204}, nil
	}
	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
	return api.DeleteTenantUuidFlowFlowId404JSONResponse{NotFoundJSONResponse: missing}, nil
}

func (sv *Server) PostTenantUuidFlowCreate(_ context.Context, request api.PostTenantUuidFlowCreateRequestObject) (api.PostTenantUuidFlowCreateResponseObject, error) {
	if request.Body == nil || len(*request.Body) == 0 {
		logrus.Errorf("no service config was provided for the new flow")
		os.Exit(1)
	}

	serviceUpdates := *request.Body

	patches := []flow_spec.ServicePatchSpec{}
	for _, serviceUpdate := range serviceUpdates {
		patch := flow_spec.ServicePatchSpec{
			Service: serviceUpdate.ServiceName,
			Image:   serviceUpdate.ImageLocator,
		}
		logrus.Infof("starting new dev flow for service %v on image %v", patch.Service, patch.Image)
		patches = append(patches, patch)
	}

	flowId, flowUrls, err := applyProdDevFlow(sv, request.Uuid, patches)
	if err != nil {
		logrus.Errorf("an error occured while updating dev flow. error was \n: '%v'", err.Error())
		return nil, err
	}
	resp := apitypes.Flow{FlowId: *flowId, FlowUrls: flowUrls}
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	logrus.Infof("getting topology for tenant '%s'", request.Uuid)

	if clusterTopology, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
			allFlowsTopology := lo.Values(allFlows)
			topo := topology.ClusterTopology(&clusterTopology, &allFlowsTopology)
			return api.GetTenantUuidTopology200JSONResponse(*topo), nil
		}
	}

	resourceType := "tenant"
	missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
	return api.GetTenantUuidTopology404JSONResponse{NotFoundJSONResponse: missing}, nil
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
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, flowID string) (error, []string) {
	clusterTopology, err := engine.GenerateProdOnlyCluster(flowID, serviceConfigs)
	if err != nil {
		return err, []string{}
	}

	sv.pluginRunnerByTenant[tenantUuidStr] = plugins.NewPluginRunner()
	sv.baseClusterTopologyByTenant[tenantUuidStr] = *clusterTopology
	sv.clusterTopologyByTenantFlow[tenantUuidStr] = make(map[string]resolved.ClusterTopology)
	flowHostMapping := clusterTopology.GetFlowHostMapping()

	return nil, flowHostMapping[flowID]
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, patches []flow_spec.ServicePatchSpec) (*string, []string, error) {
	randId := GetRandFlowID()
	flowID := fmt.Sprintf("dev-%s", randId)

	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)

	prodClusterTopology, found := sv.baseClusterTopologyByTenant[tenantUuidStr]
	if !found {
		return nil, []string{}, fmt.Errorf("no base cluster topology found for tenant %s, did you deploy the cluster?", tenantUuidStr)
	}

	pluginRunner, found := sv.pluginRunnerByTenant[tenantUuidStr]
	if !found {
		pluginRunner = plugins.NewPluginRunner()
		sv.pluginRunnerByTenant[tenantUuidStr] = pluginRunner
	}

	logrus.Debugf("calculating cluster topology overlay for tenant %s on flowID %s", tenantUuidStr, flowID)

	flowSpec := flow_spec.FlowPatchSpec{
		FlowId:         flowID,
		ServicePatches: patches,
	}
	devClusterTopology, err := engine.GenerateProdDevCluster(&prodClusterTopology, pluginRunner, flowSpec)
	if err != nil {
		return nil, []string{}, err
	}

	sv.clusterTopologyByTenantFlow[tenantUuidStr][flowID] = *devClusterTopology
	flowHostMapping := devClusterTopology.GetFlowHostMapping()

	return &flowID, flowHostMapping[flowID], nil
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
