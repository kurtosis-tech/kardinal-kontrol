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
	"kardinal.kontrol-service/database"
	"kardinal.kontrol-service/engine"
	"kardinal.kontrol-service/engine/flow"
	"kardinal.kontrol-service/plugins"
	"kardinal.kontrol-service/topology"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
	"kardinal.kontrol-service/types/flow_spec"
	"kardinal.kontrol-service/types/templates"
)

const (
	prodFlowId = "prod"
)

// optional code omitted
var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	pluginRunnerByTenant        map[string]*plugins.PluginRunner
	baseClusterTopologyByTenant map[string]resolved.ClusterTopology
	clusterTopologyByTenantFlow map[string]map[string]resolved.ClusterTopology
	templatesByNameAndTenant    map[string]map[string]templates.Template
	serviceConfigsByTenant      map[string][]apitypes.ServiceConfig // New field
	ingressConfigsByTenant      map[string][]apitypes.IngressConfig // New field
	flowTemplateMapping         map[string]string
	db                          *database.Db
	analyticsWrapper            *AnalyticsWrapper
}

func NewServer(db *database.Db, analyticsWrapper *AnalyticsWrapper) Server {
	return Server{
		pluginRunnerByTenant:        make(map[string]*plugins.PluginRunner),
		baseClusterTopologyByTenant: make(map[string]resolved.ClusterTopology),
		clusterTopologyByTenantFlow: make(map[string]map[string]resolved.ClusterTopology),
		templatesByNameAndTenant:    make(map[string]map[string]templates.Template),
		serviceConfigsByTenant:      make(map[string][]apitypes.ServiceConfig),
		ingressConfigsByTenant:      make(map[string][]apitypes.IngressConfig),
		flowTemplateMapping:         make(map[string]string),
		db:                          db,
		analyticsWrapper:            analyticsWrapper,
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
				templateName, found := sv.flowTemplateMapping[flowId]
				if !found {
					templateName = "default"
				}
				return apitypes.Flow{FlowId: flowId, FlowUrls: flowUrls, TemplateName: &templateName}
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
	sv.analyticsWrapper.TrackEvent(EVENT_DEPLOY, request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs
	ingressConfigs := *request.Body.IngressConfigs
	namespace := *request.Body.Namespace

	if namespace == "" {
		namespace = "prod"
	}

	flowId := "prod"
	sv.flowTemplateMapping["prod"] = "default"
	err, urls := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs, ingressConfigs, namespace, flowId)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred deploying flow '%v'", prodFlowId)
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.PostTenantUuidDeploy500JSONResponse{errResp}, nil
	}

	resp := apitypes.Flow{FlowId: prodFlowId, FlowUrls: urls}
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) DeleteTenantUuidFlowFlowId(_ context.Context, request api.DeleteTenantUuidFlowFlowIdRequestObject) (api.DeleteTenantUuidFlowFlowIdResponseObject, error) {
	logrus.Infof("deleting dev flow for tenant '%s'", request.Uuid)
	sv.analyticsWrapper.TrackEvent(EVENT_FLOW_DELETE, request.Uuid)

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
	sv.analyticsWrapper.TrackEvent(EVENT_FLOW_CREATE, request.Uuid)
	serviceUpdates := request.Body.FlowSpec
	templateSpec := request.Body.TemplateSpec

	patches := []flow_spec.ServicePatchSpec{}
	for _, serviceUpdate := range serviceUpdates {
		patch := flow_spec.ServicePatchSpec{
			Service: serviceUpdate.ServiceName,
			Image:   serviceUpdate.ImageLocator,
		}
		logrus.Infof("starting new dev flow for service %v on image %v", patch.Service, patch.Image)
		patches = append(patches, patch)
	}

	flowId, flowUrls, err := applyProdDevFlow(sv, request.Uuid, patches, templateSpec)
	if err != nil {
		logrus.Errorf("an error occured while updating dev flow. error was \n: '%v'", err.Error())
		return nil, err
	}

	sv.flowTemplateMapping[*flowId] = "default"
	if templateSpec != nil {
		sv.flowTemplateMapping[*flowId] = templateSpec.TemplateName
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
	if clusterTopology, found := sv.baseClusterTopologyByTenant[request.Uuid]; found {
		namespace := clusterTopology.Namespace
		if allFlows, found := sv.clusterTopologyByTenantFlow[request.Uuid]; found {
			finalTopology := flow.MergeClusterTopologies(clusterTopology, lo.Values(allFlows))
			clusterResources := flow.RenderClusterResources(finalTopology, namespace)
			managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
			return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
		}
	}

	return nil, nil
}

func (sv *Server) GetTenantUuidTemplates(ctx context.Context, request api.GetTenantUuidTemplatesRequestObject) (api.GetTenantUuidTemplatesResponseObject, error) {
	var allTemplatesForTenant []templates.Template

	for _, template := range sv.templatesByNameAndTenant[request.Uuid] {
		allTemplatesForTenant = append(allTemplatesForTenant, template)
	}

	return api.GetTenantUuidTemplates200JSONResponse(newClIAPITemplates(allTemplatesForTenant)), nil
}

func (sv *Server) DeleteTenantUuidTemplatesTemplateName(_ context.Context, request api.DeleteTenantUuidTemplatesTemplateNameRequestObject) (api.DeleteTenantUuidTemplatesTemplateNameResponseObject, error) {
	tenantUuid := request.Uuid
	templateName := request.TemplateName

	if templatesForTenant, ok := sv.templatesByNameAndTenant[tenantUuid]; ok {
		if _, exists := templatesForTenant[templateName]; exists {
			delete(templatesForTenant, templateName)

			if len(templatesForTenant) == 0 {
				delete(sv.templatesByNameAndTenant, tenantUuid)
			}

			return api.DeleteTenantUuidTemplatesTemplateName2xxResponse{StatusCode: 202}, nil
		}
	}

	resourceType := "template"
	missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: templateName}
	return api.DeleteTenantUuidTemplatesTemplateName404JSONResponse{NotFoundJSONResponse: missing}, nil
}

func (sv *Server) PostTenantUuidTemplatesCreate(_ context.Context, request api.PostTenantUuidTemplatesCreateRequestObject) (api.PostTenantUuidTemplatesCreateResponseObject, error) {
	tenantUuid := request.Uuid
	templateName := request.Body.Name
	templateDescriptionPtr := request.Body.Description
	templateOverrides := request.Body.Service
	templateId := getRandTemplateID()

	// Initialize the map for the tenant if it doesn't exist
	if _, ok := sv.templatesByNameAndTenant[tenantUuid]; !ok {
		sv.templatesByNameAndTenant[tenantUuid] = make(map[string]templates.Template)
	}

	if _, found := sv.templatesByNameAndTenant[tenantUuid][templateName]; found {
		logrus.Infof("Template with name '%v' exists; will be overwritten", templateName)
	}

	template := templates.NewTemplate(templateOverrides, templateDescriptionPtr, templateName, templateId)

	sv.templatesByNameAndTenant[tenantUuid][templateName] = template

	return api.PostTenantUuidTemplatesCreate200JSONResponse{
		Description: template.GetDescription(),
		Name:        template.GetName(),
		TemplateId:  template.GetID(),
	}, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, ingressConfigs []apitypes.IngressConfig, namespace string, flowID string) (error, []string) {
	clusterTopology, err := engine.GenerateProdOnlyCluster(flowID, serviceConfigs, ingressConfigs, namespace)
	if err != nil {
		return err, []string{}
	}

	// TODO there is an issue here where one of these get updated and failure happens
	// Perhaps have a super map / something that accounts for this
	// we need to keep this in consistent state
	sv.pluginRunnerByTenant[tenantUuidStr] = plugins.NewPluginRunner()
	sv.baseClusterTopologyByTenant[tenantUuidStr] = *clusterTopology
	sv.clusterTopologyByTenantFlow[tenantUuidStr] = make(map[string]resolved.ClusterTopology)
	sv.serviceConfigsByTenant[tenantUuidStr] = serviceConfigs
	sv.ingressConfigsByTenant[tenantUuidStr] = ingressConfigs
	flowHostMapping := clusterTopology.GetFlowHostMapping()

	return nil, flowHostMapping[flowID]
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, patches []flow_spec.ServicePatchSpec, templateSpec *apitypes.TemplateSpec) (*string, []string, error) {
	randId := getRandFlowID()
	flowID := fmt.Sprintf("dev-%s", randId)

	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)

	baseTopology, found := sv.baseClusterTopologyByTenant[tenantUuidStr]
	if !found {
		return nil, []string{}, fmt.Errorf("no base cluster topology found for tenant %s, did you deploy the cluster?", tenantUuidStr)
	}

	baseClusterTopologyMaybeWithTemplateOverrides := baseTopology
	if templateSpec != nil {
		logrus.Debugf("Using template '%v'", templateSpec.TemplateName)
		serviceConfigs, found := sv.serviceConfigsByTenant[tenantUuidStr]
		if !found {
			return nil, []string{}, fmt.Errorf("no service configs found for tenant %s, did you deploy the cluster?", tenantUuidStr)
		}
		ingressConfigs, found := sv.ingressConfigsByTenant[tenantUuidStr]
		if !found {
			return nil, []string{}, fmt.Errorf("no ingress configs found for tenant %s, did you deploy the cluster?", tenantUuidStr)
		}

		template, found := sv.templatesByNameAndTenant[tenantUuidStr][templateSpec.TemplateName]
		if !found {
			return nil, []string{}, fmt.Errorf("template with name '%v' doesn't exist for tenant uuid '%v'", templateSpec.TemplateName, tenantUuidStr)
		}
		serviceConfigs = template.ApplyTemplateOverrides(serviceConfigs, templateSpec)
		baseClusterTopologyWithTemplateOverridesPtr, err := engine.GenerateProdOnlyCluster(prodFlowId, serviceConfigs, ingressConfigs, baseTopology.Namespace)
		if err != nil {
			return nil, []string{}, fmt.Errorf("an error occurred while creating base cluster topology from templates:\n %s", err)
		}
		baseClusterTopologyMaybeWithTemplateOverrides = *baseClusterTopologyWithTemplateOverridesPtr
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
	devClusterTopology, err := engine.GenerateProdDevCluster(&baseClusterTopologyMaybeWithTemplateOverrides, &baseTopology, pluginRunner, flowSpec)
	if err != nil {
		return nil, []string{}, err
	}

	sv.clusterTopologyByTenantFlow[tenantUuidStr][flowID] = *devClusterTopology
	flowHostMapping := devClusterTopology.GetFlowHostMapping()

	return &flowID, flowHostMapping[flowID], nil
}

func newClIAPITemplates(templates []templates.Template) []apitypes.Template {
	var apiTypeTemplates []apitypes.Template
	for _, template := range templates {
		apiTypeTemplate := apitypes.Template{
			Description: template.GetDescription(),
			Name:        template.GetName(),
			TemplateId:  template.GetID(),
		}
		apiTypeTemplates = append(apiTypeTemplates, apiTypeTemplate)
	}
	return apiTypeTemplates
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
