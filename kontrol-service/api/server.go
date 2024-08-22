package api

import (
	"context"
	"encoding/json"
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
	db               *database.Db
	analyticsWrapper *AnalyticsWrapper
}

func NewServer(db *database.Db, analyticsWrapper *AnalyticsWrapper) Server {
	return Server{
		db:               db,
		analyticsWrapper: analyticsWrapper,
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
	clusterTopology, allFlows, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.GetTenantUuidFlows404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	finalTopology := flow.MergeClusterTopologies(*clusterTopology, lo.Values(allFlows))
	flowHostMapping := finalTopology.GetFlowHostMapping()
	resp := lo.MapToSlice(flowHostMapping, func(flowId string, flowUrls []string) apitypes.Flow {
		return apitypes.Flow{FlowId: flowId, FlowUrls: flowUrls}
	})
	return api.GetTenantUuidFlows200JSONResponse(resp), nil
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

	_, allFlows, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.DeleteTenantUuidFlowFlowId404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	if flowTopology, found := allFlows[request.FlowId]; found {
		logrus.Infof("deleting flow %s", request.FlowId)
		pluginRunner := plugins.NewPluginRunner(request.Uuid, sv.db)
		err := flow.DeleteFlow(pluginRunner, flowTopology, request.FlowId)
		if err != nil {
			errMsg := fmt.Sprintf("An error occurred deleting flow '%v'", request.FlowId)
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.DeleteTenantUuidFlowFlowId500JSONResponse{errResp}, nil
		}

		err = sv.db.DeleteFlow(request.Uuid, request.FlowId)
		if err != nil {
			errMsg := fmt.Sprintf("An error occurred deleting flow '%v' from the database", request.FlowId)
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.DeleteTenantUuidFlowFlowId500JSONResponse{errResp}, nil
		}

		logrus.Infof("Successfully deleted flow.")
		return api.DeleteTenantUuidFlowFlowId2xxResponse{StatusCode: 200}, nil
	}

	return api.DeleteTenantUuidFlowFlowId2xxResponse{StatusCode: 204}, nil
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
	resp := apitypes.Flow{FlowId: *flowId, FlowUrls: flowUrls}
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	logrus.Infof("getting topology for tenant '%s'", request.Uuid)

	clusterTopology, allFlows, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.GetTenantUuidTopology404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	allFlowsTopology := lo.Values(allFlows)
	topo := topology.ClusterTopology(clusterTopology, &allFlowsTopology)
	return api.GetTenantUuidTopology200JSONResponse(*topo), nil
}

func (sv *Server) GetTenantUuidClusterResources(_ context.Context, request managerapi.GetTenantUuidClusterResourcesRequestObject) (managerapi.GetTenantUuidClusterResourcesResponseObject, error) {
	clusterTopology, allFlows, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		return nil, nil
	}

	namespace := clusterTopology.Namespace
	finalTopology := flow.MergeClusterTopologies(*clusterTopology, lo.Values(allFlows))
	clusterResources := flow.RenderClusterResources(finalTopology, namespace)
	managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
	return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
}

func (sv *Server) GetTenantUuidTemplates(ctx context.Context, request api.GetTenantUuidTemplatesRequestObject) (api.GetTenantUuidTemplatesResponseObject, error) {
	_, _, tenantTemplates, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.GetTenantUuidTemplates404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	var allTemplatesForTenant []templates.Template

	for _, template := range tenantTemplates {
		allTemplatesForTenant = append(allTemplatesForTenant, template)
	}

	return api.GetTenantUuidTemplates200JSONResponse(newClIAPITemplates(allTemplatesForTenant)), nil
}

func (sv *Server) DeleteTenantUuidTemplatesTemplateName(_ context.Context, request api.DeleteTenantUuidTemplatesTemplateNameRequestObject) (api.DeleteTenantUuidTemplatesTemplateNameResponseObject, error) {
	tenantUuid := request.Uuid
	templateName := request.TemplateName

	_, _, tenantTemplates, _, _, err := getTenantTopologies(sv, tenantUuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.DeleteTenantUuidTemplatesTemplateName404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	if _, exists := tenantTemplates[templateName]; exists {
		err = sv.db.DeleteTemplate(tenantUuid, templateName)
		if err != nil {
			errMsg := fmt.Sprintf("An error occurred deleting template '%v' from the database", templateName)
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.DeleteTenantUuidTemplatesTemplateName500JSONResponse{errResp}, nil
		}

		return api.DeleteTenantUuidTemplatesTemplateName2xxResponse{StatusCode: 202}, nil
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

	_, _, tenantTemplates, _, _, err := getTenantTopologies(sv, tenantUuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.PostTenantUuidTemplatesCreate404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	template := templates.NewTemplate(templateOverrides, templateDescriptionPtr, templateName, templateId)
	templateJson, err := json.Marshal(template)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred encoding template %s, error was \n: %v", templateName, err.Error())
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.PostTenantUuidTemplatesCreate500JSONResponse{errResp}, nil
	}

	dbTemplate := &database.Template{
		Name:     templateName,
		Body:     templateJson,
		TenantId: tenantUuid,
	}

	if _, found := tenantTemplates[templateName]; found {
		logrus.Infof("Template with name '%v' exists; will be overwritten", templateName)
		err = sv.db.SaveTemplate(dbTemplate)
		if err != nil {
			errMsg := fmt.Sprintf("An error occurred updating template '%v' from the database", templateName)
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.PostTenantUuidTemplatesCreate500JSONResponse{errResp}, nil
		}
	} else {
		_, err = sv.db.CreateTemplate(tenantUuid, templateName, templateJson)
		if err != nil {
			errMsg := fmt.Sprintf("An error occurred updating template '%v' from the database", templateName)
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.PostTenantUuidTemplatesCreate500JSONResponse{errResp}, nil
		}
	}

	return api.PostTenantUuidTemplatesCreate200JSONResponse{
		Description: template.Description,
		Name:        template.Name,
		TemplateId:  template.Id,
	}, nil
}

// ============================================================================================================
func applyProdOnlyFlow(sv *Server, tenantUuidStr string, serviceConfigs []apitypes.ServiceConfig, ingressConfigs []apitypes.IngressConfig, namespace string, flowID string) (error, []string) {
	clusterTopology, err := engine.GenerateProdOnlyCluster(flowID, serviceConfigs, ingressConfigs, namespace)
	if err != nil {
		return err, []string{}
	}

<<<<<<< HEAD
	// TODO there is an issue here where one of these get updated and failure happens
	// Perhaps have a super map / something that accounts for this
	// we need to keep this in consistent state
	sv.pluginRunnerByTenant[tenantUuidStr] = plugins.NewPluginRunner(plugins.NewGitPluginProviderImpl())
	sv.baseClusterTopologyByTenant[tenantUuidStr] = *clusterTopology
	sv.clusterTopologyByTenantFlow[tenantUuidStr] = make(map[string]resolved.ClusterTopology)
	sv.serviceConfigsByTenant[tenantUuidStr] = serviceConfigs
	sv.ingressConfigsByTenant[tenantUuidStr] = ingressConfigs
=======
	tenant, err := sv.db.GetOrCreateTenant(tenantUuidStr)
	if err != nil {
		logrus.Errorf("an error occured while getting the tenant %s\n: '%v'", tenantUuidStr, err.Error())
		return err, nil
	}

	clusterTopologyJson, err := json.Marshal(clusterTopology)
	if err != nil {
		logrus.Errorf("an error occured while encoding the cluster topology for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return err, nil
	}
	tenant.BaseClusterTopology = clusterTopologyJson

	serviceConfigsJson, err := json.Marshal(serviceConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the service configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return err, nil
	}
	tenant.ServiceConfigs = serviceConfigsJson

	ingressConfigsJson, err := json.Marshal(ingressConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the ingress configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return err, nil
	}
	tenant.IngressConfigs = ingressConfigsJson

	err = sv.db.SaveTenant(tenant)
	if err != nil {
		logrus.Errorf("an error occured while saving tenant %s. erro was \n: '%v'", tenant.TenantId, err.Error())
		return err, nil
	}

>>>>>>> main
	flowHostMapping := clusterTopology.GetFlowHostMapping()

	return nil, flowHostMapping[flowID]
}

// ============================================================================================================
func applyProdDevFlow(sv *Server, tenantUuidStr string, patches []flow_spec.ServicePatchSpec, templateSpec *apitypes.TemplateSpec) (*string, []string, error) {
	randId := getRandFlowID()
	flowID := fmt.Sprintf("dev-%s", randId)

	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)

	baseTopology, _, tenantTemplates, serviceConfigs, ingressConfigs, err := getTenantTopologies(sv, tenantUuidStr)
	if err != nil {
		return nil, []string{}, fmt.Errorf("no base cluster topology found for tenant %s, did you deploy the cluster?", tenantUuidStr)
	}

	baseClusterTopologyMaybeWithTemplateOverrides := *baseTopology
	if templateSpec != nil {
		logrus.Debugf("Using template '%v'", templateSpec.TemplateName)

		template, found := tenantTemplates[templateSpec.TemplateName]
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

<<<<<<< HEAD
	pluginRunner, found := sv.pluginRunnerByTenant[tenantUuidStr]
	if !found {
		pluginRunner = plugins.NewPluginRunner(plugins.NewGitPluginProviderImpl())
		sv.pluginRunnerByTenant[tenantUuidStr] = pluginRunner
	}

=======
>>>>>>> main
	logrus.Debugf("calculating cluster topology overlay for tenant %s on flowID %s", tenantUuidStr, flowID)

	flowSpec := flow_spec.FlowPatchSpec{
		FlowId:         flowID,
		ServicePatches: patches,
	}

	pluginRunner := plugins.NewPluginRunner(tenantUuidStr, sv.db)
	devClusterTopology, err := engine.GenerateProdDevCluster(&baseClusterTopologyMaybeWithTemplateOverrides, baseTopology, pluginRunner, flowSpec)
	if err != nil {
		return nil, []string{}, err
	}

	devClusterTopologyJson, err := json.Marshal(devClusterTopology)
	if err != nil {
		logrus.Errorf("an error occured while encoding the cluster topology for tenant %s and flow %s, error was \n: '%v'", tenantUuidStr, flowID, err.Error())
		return nil, []string{}, err
	}

	_, err = sv.db.CreateFlow(tenantUuidStr, flowID, devClusterTopologyJson)
	if err != nil {
		logrus.Errorf("an error occured while creating flow %s. error was \n: '%v'", flowID, err.Error())
		return nil, []string{}, err
	}

	flowHostMapping := devClusterTopology.GetFlowHostMapping()

	return &flowID, flowHostMapping[flowID], nil
}

// Returns the following given a tenant ID:
// - Base cluster topology
// - Flows topology
// - Templates
// - Base service configs
// - Base ingress configs
// TOOD: Could return a struct if it becomes too heavy to manipulate the return values.
func getTenantTopologies(sv *Server, tenantUuidStr string) (*resolved.ClusterTopology, map[string]resolved.ClusterTopology, map[string]templates.Template, []apitypes.ServiceConfig, []apitypes.IngressConfig, error) {
	tenant, err := sv.db.GetOrCreateTenant(tenantUuidStr)
	if err != nil {
		logrus.Errorf("an error occured while getting the tenant %s\n: '%v'", tenantUuidStr, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	var clusterTopology resolved.ClusterTopology
	err = json.Unmarshal(tenant.BaseClusterTopology, &clusterTopology)
	if err != nil {
		logrus.Errorf("An error occurred decoding the cluster topology for tenant '%v'", tenant.TenantId)
		return nil, nil, nil, nil, nil, err
	}

	flows := map[string]resolved.ClusterTopology{}
	for _, flow := range tenant.Flows {
		var clusterTopology resolved.ClusterTopology
		err := json.Unmarshal(flow.ClusterTopology, &clusterTopology)
		if err != nil {
			logrus.Errorf("An error occurred decoding the cluster topology for flow '%v'", flow.FlowId)
			return nil, nil, nil, nil, nil, err
		}
		flows[flow.FlowId] = clusterTopology
	}

	tenantTemplates := map[string]templates.Template{}
	for _, tenantTemplate := range tenant.Templates {
		var template templates.Template
		err := json.Unmarshal(tenantTemplate.Body, &template)
		if err != nil {
			logrus.Errorf("An error occurred decoding the template body for template '%v'", tenantTemplate.Name)
			return nil, nil, nil, nil, nil, err
		}
		tenantTemplates[tenantTemplate.Name] = template
	}

	var baseClusterTopology resolved.ClusterTopology
	err = json.Unmarshal(tenant.BaseClusterTopology, &baseClusterTopology)
	if err != nil {
		logrus.Errorf("An error occurred decoding the cluster topology for tenant '%v'", tenantUuidStr)
		return nil, nil, nil, nil, nil, err
	}

	var serviceConfigs []apitypes.ServiceConfig
	err = json.Unmarshal(tenant.ServiceConfigs, &serviceConfigs)
	if err != nil {
		logrus.Errorf("An error occurred decoding the service configs for tenant '%v'", tenantUuidStr)
		return nil, nil, nil, nil, nil, err
	}

	var ingressConfigs []apitypes.IngressConfig
	err = json.Unmarshal(tenant.IngressConfigs, &ingressConfigs)
	if err != nil {
		logrus.Errorf("An error occurred decoding the ingress configs for tenant '%v'", tenantUuidStr)
		return nil, nil, nil, nil, nil, err
	}
	return &baseClusterTopology, flows, tenantTemplates, serviceConfigs, ingressConfigs, nil
}

func newClIAPITemplates(templates []templates.Template) []apitypes.Template {
	var apiTypeTemplates []apitypes.Template
	for _, template := range templates {
		apiTypeTemplate := apitypes.Template{
			Description: template.Description,
			Name:        template.Name,
			TemplateId:  template.Id,
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
