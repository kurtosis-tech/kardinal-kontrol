package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	managerapi "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/server"
	managerapitypes "github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api/api/golang/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/printers"
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
	defaultBaselineFlowId = "baseline"
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
	err := sv.db.Check()
	if err != nil {
		errMsg := "An error occurred checking the database connection"
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.GetHealth500JSONResponse{ErrorJSONResponse: errResp}, nil
	}
	resp := "ok"
	return api.GetHealth200JSONResponse(resp), nil
}

func (sv *Server) GetTenantUuidFlows(_ context.Context, request api.GetTenantUuidFlowsRequestObject) (api.GetTenantUuidFlowsResponseObject, error) {
	clusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.GetTenantUuidFlows404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	finalTopology := flow.MergeClusterTopologies(*clusterTopology, lo.Values(allFlows))
	flowHostMapping := finalTopology.GetFlowHostMapping()
	resp := lo.MapToSlice(flowHostMapping, func(flowId string, entries []resolved.IngressAccessEntry) apitypes.Flow {
		isBaselineFlow := flowId == clusterTopology.Namespace
		return apitypes.Flow{FlowId: flowId, AccessEntry: toApiIngressAccessEntries(entries), IsBaseline: &isBaselineFlow}
	})
	return api.GetTenantUuidFlows200JSONResponse(resp), nil
}

func (sv *Server) PostTenantUuidDeploy(_ context.Context, request api.PostTenantUuidDeployRequestObject) (api.PostTenantUuidDeployResponseObject, error) {
	logrus.Infof("deploying baseline cluster for tenant '%s'", request.Uuid)
	sv.analyticsWrapper.TrackEvent(EVENT_DEPLOY, request.Uuid)
	serviceConfigs := *request.Body.ServiceConfigs

	deploymentConfigs := []apitypes.DeploymentConfig{}
	if request.Body.DeploymentConfigs != nil {
		deploymentConfigs = *request.Body.DeploymentConfigs
	}

	statefulSetConfigs := []apitypes.StatefulSetConfig{}
	if request.Body.StatefulSetConfigs != nil {
		statefulSetConfigs = *request.Body.StatefulSetConfigs
	}

	ingressConfigs := []apitypes.IngressConfig{}
	if request.Body.IngressConfigs != nil {
		ingressConfigs = *request.Body.IngressConfigs
	}

	gatewayConfigs := []apitypes.GatewayConfig{}
	if request.Body.GatewayConfigs != nil {
		gatewayConfigs = *request.Body.GatewayConfigs
	}

	routesConfigs := []apitypes.RouteConfig{}
	if request.Body.RouteConfigs != nil {
		routesConfigs = *request.Body.RouteConfigs
	}

	namespace := *request.Body.Namespace

	if namespace == "" {
		namespace = defaultBaselineFlowId
	}

	flowId := namespace
	entries, err := applyProdOnlyFlow(sv, request.Uuid, serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routesConfigs, namespace, flowId)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred deploying flow '%v'", flowId)
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.PostTenantUuidDeploy500JSONResponse{ErrorJSONResponse: errResp}, nil
	}

	resp := apitypes.Flow{FlowId: flowId, AccessEntry: toApiIngressAccessEntries(entries)}
	return api.PostTenantUuidDeploy200JSONResponse(resp), nil
}

func (sv *Server) DeleteTenantUuidFlowFlowId(_ context.Context, request api.DeleteTenantUuidFlowFlowIdRequestObject) (api.DeleteTenantUuidFlowFlowIdResponseObject, error) {
	logrus.Infof("deleting dev flow for tenant '%s'", request.Uuid)
	sv.analyticsWrapper.TrackEvent(EVENT_FLOW_DELETE, request.Uuid)

	baseClusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		resourceType := "tenant"
		missing := api.NotFoundJSONResponse{ResourceType: resourceType, Id: request.Uuid}
		return api.DeleteTenantUuidFlowFlowId404JSONResponse{NotFoundJSONResponse: missing}, nil
	}

	// the baseline flow ID uses the base cluster topology namespace name
	if request.FlowId == baseClusterTopology.Namespace {
		// We received a request to delete the base topology, so we do that + the flows
		err = deleteTenantTopologies(sv, request.Uuid)
		if err != nil {
			errMsg := "An error occurred deleting the topologies"
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			return api.DeleteTenantUuidFlowFlowId500JSONResponse{errResp}, nil
		}

		logrus.Infof("Successfully deleted topologies.")
		return api.DeleteTenantUuidFlowFlowId2xxResponse{StatusCode: 200}, nil
	}

	if flowTopology, found := allFlows[request.FlowId]; found {
		logrus.Infof("deleting flow %s", request.FlowId)
		pluginRunner := plugins.NewPluginRunner(plugins.NewGitPluginProviderImpl(), request.Uuid, sv.db)
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

	requestTenantUuid := request.Uuid
	requestFlowId := request.Body.FlowId

	flowId, flowIdAlreadyExist, err := sv.checkOrCreateFlowID(requestTenantUuid, requestFlowId)
	if err != nil {
		var apiErrResponse api.PostTenantUuidFlowCreateResponseObject
		errMsg := fmt.Sprintf("An error occurred checking or creating flow ID %s", *request.Body.FlowId)
		if flowIdAlreadyExist {
			errResp := api.RequestErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			apiErrResponse = api.PostTenantUuidFlowCreate400JSONResponse{RequestErrorJSONResponse: errResp}
		} else {
			errResp := api.ErrorJSONResponse{
				Error: err.Error(),
				Msg:   &errMsg,
			}
			apiErrResponse = api.PostTenantUuidFlowCreate500JSONResponse{ErrorJSONResponse: errResp}
		}

		return apiErrResponse, nil
	}

	entries, err := applyProdDevFlow(flowId, sv, request.Uuid, patches, templateSpec)
	if err != nil {
		errMsg := "An error occurred creating flow"
		errResp := api.ErrorJSONResponse{
			Error: err.Error(),
			Msg:   &errMsg,
		}
		return api.PostTenantUuidFlowCreate500JSONResponse{ErrorJSONResponse: errResp}, nil
	}
	resp := apitypes.Flow{FlowId: flowId, AccessEntry: toApiIngressAccessEntries(entries)}
	return api.PostTenantUuidFlowCreate200JSONResponse(resp), nil
}

func (sv *Server) checkOrCreateFlowID(tenantUuid apitypes.Uuid, requestFlowId *string) (string, bool, error) {
	var flowIdAlreadyExist bool

	clusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, tenantUuid)
	if err != nil {
		return "", flowIdAlreadyExist, err
	}

	// Create the flow ID if it was not provided
	if requestFlowId == nil || *requestFlowId == "" {

		// The baseline flow ID is not stored, it's inferred from the namespace name, currently these are the same
		baselineFlowId := clusterTopology.Namespace

		randId := getRandID()

		newFlowId := fmt.Sprintf("%s-%s", baselineFlowId, randId)
		return newFlowId, flowIdAlreadyExist, nil
	}

	// check received flowID from the request
	_, found := lo.FindKeyBy(allFlows, func(key string, value resolved.ClusterTopology) bool {
		return key == *requestFlowId
	})
	if found {
		flowIdAlreadyExist = true
		return "", flowIdAlreadyExist, stacktrace.NewError("flow id '%s' already exists", *requestFlowId)
	}

	return *requestFlowId, flowIdAlreadyExist, nil
}

func (sv *Server) GetTenantUuidTopology(_ context.Context, request api.GetTenantUuidTopologyRequestObject) (api.GetTenantUuidTopologyResponseObject, error) {
	logrus.Infof("getting topology for tenant '%s'", request.Uuid)

	clusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
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
	clusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		return nil, nil
	}

	namespace := clusterTopology.Namespace
	finalTopology := flow.MergeClusterTopologies(*clusterTopology, lo.Values(allFlows))
	clusterResources := flow.RenderClusterResources(finalTopology, namespace)
	managerAPIClusterResources := newManagerAPIClusterResources(clusterResources)
	return managerapi.GetTenantUuidClusterResources200JSONResponse(managerAPIClusterResources), nil
}

func (sv *Server) GetTenantUuidManifest(_ context.Context, request api.GetTenantUuidManifestRequestObject) (api.GetTenantUuidManifestResponseObject, error) {
	logrus.Infof("generating manifest for tenant '%s'", request.Uuid)
	clusterTopology, allFlows, _, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
	if err != nil {
		logrus.WithError(err).Errorf("An error occurred while getting topologys for tenant '%s'", request.Uuid)
		return nil, err
	}

	if clusterTopology != nil {
		namespaceName := clusterTopology.Namespace
		if allFlows != nil {
			finalTopology := flow.MergeClusterTopologies(*clusterTopology, lo.Values(allFlows))
			clusterResources := flow.RenderClusterResources(finalTopology, namespaceName)

			var yamlBuffer bytes.Buffer
			yamlPrinter := printers.YAMLPrinter{}

			// Add namespace
			newNamespace := types.NewNamespaceWithIstioEnabled(namespaceName)

			if err = yamlPrinter.PrintObj(newNamespace, &yamlBuffer); err != nil {
				logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", newNamespace.Name)
				return nil, stacktrace.Propagate(err, "an error occurred printing the cluster topology namespace '%s' in the yaml buffer", namespaceName)
			}

			for _, resource := range clusterResources.Deployments {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing deployment '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.Services {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing service '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.VirtualServices {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing virtual service '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.DestinationRules {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing destination rule '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.EnvoyFilters {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing envoy filter '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.AuthorizationPolicies {
				if err = yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing authorization policy '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.Gateways {
				if err := yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing gateway '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.HTTPRoutes {
				if err := yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing http route '%s' in the yaml buffer", resource.Name)
				}
			}

			for _, resource := range clusterResources.Ingresses {
				if err := yamlPrinter.PrintObj(&resource, &yamlBuffer); err != nil {
					logrus.WithError(err).Errorf("An error occurred printing '%s' in the yaml buffer", resource.Name)
					return nil, stacktrace.Propagate(err, "an error occurred printing ingress '%s' in the yaml buffer", resource.Name)
				}
			}

			response := api.GetTenantUuidManifest200ApplicationxYamlResponse{
				Body:          &yamlBuffer,
				ContentLength: int64(yamlBuffer.Len()),
			}

			return response, nil
		}
	}
	return nil, nil
}

func (sv *Server) GetTenantUuidTemplates(ctx context.Context, request api.GetTenantUuidTemplatesRequestObject) (api.GetTenantUuidTemplatesResponseObject, error) {
	_, _, tenantTemplates, _, _, _, _, _, _, err := getTenantTopologies(sv, request.Uuid)
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

	_, _, tenantTemplates, _, _, _, _, _, _, err := getTenantTopologies(sv, tenantUuid)
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

	_, _, tenantTemplates, _, _, _, _, _, _, err := getTenantTopologies(sv, tenantUuid)
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
func applyProdOnlyFlow(
	sv *Server,
	tenantUuidStr string,
	serviceConfigs []apitypes.ServiceConfig,
	deploymentConfigs []apitypes.DeploymentConfig,
	statefulSetConfigs []apitypes.StatefulSetConfig,
	ingressConfigs []apitypes.IngressConfig,
	gatewayConfigs []apitypes.GatewayConfig,
	routeConfigs []apitypes.RouteConfig,
	namespace string,
	flowID string,
) ([]resolved.IngressAccessEntry, error) {
	clusterTopology, err := engine.GenerateProdOnlyCluster(flowID, serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routeConfigs, namespace)
	if err != nil {
		return nil, err
	}

	tenant, err := sv.db.GetOrCreateTenant(tenantUuidStr)
	if err != nil {
		logrus.Errorf("an error occured while getting the tenant %s\n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}

	clusterTopologyJson, err := json.Marshal(clusterTopology)
	if err != nil {
		logrus.Errorf("an error occured while encoding the cluster topology for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}
	tenant.BaseClusterTopology = clusterTopologyJson

	serviceConfigsJson, err := json.Marshal(serviceConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the service configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}
	tenant.ServiceConfigs = serviceConfigsJson

	ingressConfigsJson, err := json.Marshal(ingressConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the ingress configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}
	tenant.IngressConfigs = ingressConfigsJson

	gatewayConfigsJson, err := json.Marshal(gatewayConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the gateway configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}
	tenant.GatewayConfigs = gatewayConfigsJson

	routeConfigsJson, err := json.Marshal(routeConfigs)
	if err != nil {
		logrus.Errorf("an error occured while encoding the ingress configs for tenant %s, error was \n: '%v'", tenantUuidStr, err.Error())
		return nil, err
	}
	tenant.RouteConfigs = routeConfigsJson

	err = sv.db.SaveTenant(tenant)
	if err != nil {
		logrus.Errorf("an error occured while saving tenant %s. erro was \n: '%v'", tenant.TenantId, err.Error())
		return nil, err
	}

	flowHostMapping := clusterTopology.GetFlowHostMapping()

	return flowHostMapping[flowID], nil
}

// ============================================================================================================
func applyProdDevFlow(
	flowID string,
	sv *Server,
	tenantUuidStr string,
	patches []flow_spec.ServicePatchSpec,
	templateSpec *apitypes.TemplateSpec,
) ([]resolved.IngressAccessEntry, error) {
	logrus.Debugf("generating base cluster topology for tenant %s on flowID %s", tenantUuidStr, flowID)

	baseTopology, _, tenantTemplates, serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routeConfigs, err := getTenantTopologies(sv, tenantUuidStr)
	if err != nil {
		return nil, fmt.Errorf("no base cluster topology found for tenant %s, did you deploy the cluster?", tenantUuidStr)
	}

	baseClusterTopologyMaybeWithTemplateOverrides := *baseTopology
	if templateSpec != nil {
		logrus.Debugf("Using template '%v'", templateSpec.TemplateName)

		template, found := tenantTemplates[templateSpec.TemplateName]
		if !found {
			return nil, fmt.Errorf("template with name '%v' doesn't exist for tenant uuid '%v'", templateSpec.TemplateName, tenantUuidStr)
		}
		serviceConfigs = template.ApplyTemplateOverrides(serviceConfigs, templateSpec)

		// the baseline flow ID uses the base cluster topology namespace name
		baselineFlowID := baseClusterTopologyMaybeWithTemplateOverrides.Namespace

		baseClusterTopologyWithTemplateOverridesPtr, err := engine.GenerateProdOnlyCluster(baselineFlowID, serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routeConfigs, baseTopology.Namespace)
		if err != nil {
			return nil, fmt.Errorf("an error occurred while creating base cluster topology from templates:\n %s", err)
		}
		baseClusterTopologyMaybeWithTemplateOverrides = *baseClusterTopologyWithTemplateOverridesPtr
	}

	logrus.Debugf("calculating cluster topology overlay for tenant %s on flowID %s", tenantUuidStr, flowID)

	flowSpec := flow_spec.FlowPatchSpec{
		FlowId:         flowID,
		ServicePatches: patches,
	}

	pluginRunner := plugins.NewPluginRunner(plugins.NewGitPluginProviderImpl(), tenantUuidStr, sv.db)
	devClusterTopology, err := engine.GenerateProdDevCluster(&baseClusterTopologyMaybeWithTemplateOverrides, baseTopology, pluginRunner, flowSpec)
	if err != nil {
		return nil, err
	}

	devClusterTopologyJson, err := json.Marshal(devClusterTopology)
	if err != nil {
		logrus.Errorf("an error occured while encoding the cluster topology for tenant %s and flow %s, error was \n: '%v'", tenantUuidStr, flowID, err.Error())
		return nil, err
	}

	_, err = sv.db.CreateFlow(tenantUuidStr, flowID, devClusterTopologyJson)
	if err != nil {
		logrus.Errorf("an error occured while creating flow %s. error was \n: '%v'", flowID, err.Error())
		return nil, err
	}

	flowHostMapping := devClusterTopology.GetFlowHostMapping()

	return flowHostMapping[flowID], nil
}

// Returns the following given a tenant ID:
// - Base cluster topology
// - Flows topology
// - Templates
// - Base service configs
// - Base ingress configs
// TOOD: Could return a struct if it becomes too heavy to manipulate the return values.
func getTenantTopologies(sv *Server, tenantUuidStr string) (*resolved.ClusterTopology,
	map[string]resolved.ClusterTopology,
	map[string]templates.Template,
	[]apitypes.ServiceConfig,
	[]apitypes.DeploymentConfig,
	[]apitypes.StatefulSetConfig,
	[]apitypes.IngressConfig,
	[]apitypes.GatewayConfig,
	[]apitypes.RouteConfig,
	error,
) {
	tenant, err := sv.db.GetTenant(tenantUuidStr)
	if err != nil {
		logrus.Errorf("an error occured while getting the tenant %s\n: '%v'", tenantUuidStr, err.Error())
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	if tenant == nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("Cannot find tenant %s", tenantUuidStr)
	}

	flows := map[string]resolved.ClusterTopology{}
	for _, flow := range tenant.Flows {
		var clusterTopology resolved.ClusterTopology
		err := json.Unmarshal(flow.ClusterTopology, &clusterTopology)
		if err != nil {
			logrus.Errorf("An error occurred decoding the cluster topology for flow '%v'", flow.FlowId)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		flows[flow.FlowId] = clusterTopology
	}

	tenantTemplates := map[string]templates.Template{}
	for _, tenantTemplate := range tenant.Templates {
		var template templates.Template
		err := json.Unmarshal(tenantTemplate.Body, &template)
		if err != nil {
			logrus.Errorf("An error occurred decoding the template body for template '%v'", tenantTemplate.Name)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		tenantTemplates[tenantTemplate.Name] = template
	}

	var baseClusterTopology resolved.ClusterTopology
	if tenant.BaseClusterTopology != nil {
		err = json.Unmarshal(tenant.BaseClusterTopology, &baseClusterTopology)
		if err != nil {
			logrus.Errorf("An error occurred decoding the cluster topology for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	} else {
		baseClusterTopology.FlowID = defaultBaselineFlowId
		baseClusterTopology.Namespace = defaultBaselineFlowId
	}

	var serviceConfigs []apitypes.ServiceConfig
	if tenant.ServiceConfigs != nil {
		err = json.Unmarshal(tenant.ServiceConfigs, &serviceConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the service configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	var deploymentConfigs []apitypes.DeploymentConfig
	if tenant.DeploymentConfigs != nil {
		err = json.Unmarshal(tenant.DeploymentConfigs, &deploymentConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the deployment configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	var statefulSetConfigs []apitypes.StatefulSetConfig
	if tenant.StatefulSetConfigs != nil {
		err = json.Unmarshal(tenant.StatefulSetConfigs, &statefulSetConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the stateful set configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	var ingressConfigs []apitypes.IngressConfig
	if tenant.IngressConfigs != nil {
		err = json.Unmarshal(tenant.IngressConfigs, &ingressConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the ingress configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	var gatewayConfigs []apitypes.GatewayConfig
	if tenant.GatewayConfigs != nil {
		err = json.Unmarshal(tenant.GatewayConfigs, &gatewayConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the gateway configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	var routeConfigs []apitypes.RouteConfig
	if tenant.RouteConfigs != nil {
		err = json.Unmarshal(tenant.RouteConfigs, &routeConfigs)
		if err != nil {
			logrus.Errorf("An error occurred decoding the route configs for tenant '%v'", tenantUuidStr)
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	return &baseClusterTopology, flows, tenantTemplates, serviceConfigs, deploymentConfigs, statefulSetConfigs, ingressConfigs, gatewayConfigs, routeConfigs, nil
}

func deleteTenantTopologies(sv *Server, tenantUuidStr string) error {
	tenant, err := sv.db.GetTenant(tenantUuidStr)
	if err != nil {
		logrus.Errorf("an error occured while getting the tenant %s\n: '%v'", tenantUuidStr, err.Error())
		return err
	}

	tenant.BaseClusterTopology = nil
	tenant.ServiceConfigs = nil
	tenant.IngressConfigs = nil

	err = sv.db.SaveTenant(tenant)
	if err != nil {
		logrus.Errorf("an error occured while saving tenant %s. erro was \n: '%v'", tenant.TenantId, err.Error())
		return err
	}

	err = sv.db.DeleteTenantFlows(tenant.TenantId)
	if err != nil {
		logrus.Errorf("an error occured while deleting tenant flows %s. erro was \n: '%v'", tenant.TenantId, err.Error())
		return err
	}

	err = sv.db.DeleteTenantPluginConfigs(tenant.TenantId)
	if err != nil {
		logrus.Errorf("an error occured while deleting tenant plugin configs %s. erro was \n: '%v'", tenant.TenantId, err.Error())
		return err
	}

	return nil
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
		Services:              &clusterResources.Services,
		Deployments:           &clusterResources.Deployments,
		StatefulSets:          &clusterResources.StatefulSets,
		VirtualServices:       &clusterResources.VirtualServices,
		DestinationRules:      &clusterResources.DestinationRules,
		Gateways:              &clusterResources.Gateways,
		HttpRoutes:            &clusterResources.HTTPRoutes,
		Ingresses:             &clusterResources.Ingresses,
		EnvoyFilters:          &clusterResources.EnvoyFilters,
		AuthorizationPolicies: &clusterResources.AuthorizationPolicies,
	}
}

func toApiIngressAccessEntries(entries []resolved.IngressAccessEntry) []apitypes.IngressAccessEntry {
	return lo.Map(entries, func(item resolved.IngressAccessEntry, _ int) apitypes.IngressAccessEntry {
		return apitypes.IngressAccessEntry{
			FlowId:        item.FlowID,
			FlowNamespace: item.FlowNamespace,
			Hostname:      item.Hostname,
			Service:       item.Service,
			Namespace:     item.Namespace,
			Type:          item.Type,
		}
	})
}
