package flow

import (
	"fmt"
	"kardinal.kontrol-service/constants"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"istio.io/api/networking/v1alpha3"
	securityapi "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"
)

// RenderClusterResources returns a cluster resource for a given topology
// Perhaps we can make this throw an error if the # of extHosts != # of versions
// This assumes that there is a dev version of the ext host as well
func RenderClusterResources(clusterTopology *resolved.ClusterTopology, namespace string) types.ClusterResources {
	virtualServices := []istioclient.VirtualService{}
	destinationRules := []istioclient.DestinationRule{}
	envoyFilters := []istioclient.EnvoyFilter{}

	authorizationPolicies := []securityv1beta1.AuthorizationPolicy{}

	servicesAgainstVersions := map[string][]string{}
	serviceList := []v1.Service{}

	allActiveFlows := lo.FlatMap(clusterTopology.Ingresses, func(item *resolved.Ingress, _ int) []string { return item.ActiveFlowIDs })
	var versionsAgainstExtHost map[string]string

	groupedServices := lo.GroupBy(clusterTopology.Services, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		logrus.Infof("Rendering service with id: '%v'.", serviceID)
		servicesAgainstVersions[serviceID] = lo.Map(services, func(item *resolved.Service, _ int) string { return item.Version })
		if len(services) > 0 {
			// TODO: this assumes service specs didn't change. May we need a new version to ClusterTopology data structure

			// ServiceSpec is nil for external services - don't process anything bc theres nothing to add to the cluster
			if services[0].ServiceSpec == nil {
				continue
			}
			serviceList = append(serviceList, *getService(services[0], namespace))

			var gateway *string
			var extHost *string
			if ingressService, found := clusterTopology.GetIngressForService(services[0]); found {
				logrus.Infof("The service has an ingress")
				gateway = &ingressService.IngressID
				extHost = ingressService.GetHost()

				// TODO: either update getEnvoyFilters or merge maps in case there is more than one ingress
				versionsAgainstExtHost = lo.SliceToMap(allActiveFlows, func(item string) (string, string) { return item, resolved.ReplaceOrAddSubdomain(*extHost, item) })
			}

			virtualService, destinationRule := getVirtualService(serviceID, services, namespace, gateway, extHost)
			virtualServices = append(virtualServices, *virtualService)
			if destinationRule != nil {
				destinationRules = append(destinationRules, *destinationRule)
			}
			logrus.Infof("adding filters and authorization policies for service '%s'", serviceID)

			envoyFiltersForService := getEnvoyFilters(services[0], namespace)
			envoyFilters = append(envoyFilters, envoyFiltersForService...)

			authorizationPolicy := getAuthorizationPolicy(services[0], namespace)
			if authorizationPolicy != nil {
				authorizationPolicies = append(authorizationPolicies, *authorizationPolicy)
			}
		}

	}

	logrus.Infof("have total of %d envoy filters", len(envoyFilters))

	sharedServiceBackupVersions := map[string][]string{}
	for _, service := range clusterTopology.Services {
		if service.IsShared && service.Version == constants.SharedVersionVersionString {
			logrus.Infof("Found original version '%v' for service '%v'", service.OriginalVersionIfShared, service.ServiceID)
			sharedServiceBackupVersions[service.ServiceID] = append(sharedServiceBackupVersions[service.ServiceID], service.OriginalVersionIfShared)
		}
	}
	// TODO: make it to use a list of Ingresses
	gatewayFilter := getEnvoyFilterForGateway(servicesAgainstVersions, versionsAgainstExtHost, sharedServiceBackupVersions)
	envoyFilters = append(envoyFilters, *gatewayFilter)

	return types.ClusterResources{
		Services: serviceList,

		Deployments: lo.FilterMap(clusterTopology.Services, func(service *resolved.Service, _ int) (appsv1.Deployment, bool) {
			// Deployment spec is nil for external services, don't need to add anything to cluster
			if service.DeploymentSpec == nil {
				return appsv1.Deployment{}, false
			}
			return *getDeployment(service, namespace), true
		}),

		Gateway: *getGateway(clusterTopology.Ingresses, namespace),

		VirtualServices: virtualServices,

		DestinationRules: destinationRules,

		EnvoyFilters: envoyFilters,

		AuthorizationPolicies: []securityv1beta1.AuthorizationPolicy{},
	}
}

func getTCPRoute(service *resolved.Service, servicePort *v1.ServicePort) *v1alpha3.TCPRoute {
	return &v1alpha3.TCPRoute{
		Match: []*v1alpha3.L4MatchAttributes{{
			Port: uint32(servicePort.Port),
			SourceLabels: map[string]string{
				"version": service.Version,
			},
		}},
		Route: []*v1alpha3.RouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   service.ServiceID,
					Subset: service.Version,
					Port: &v1alpha3.PortSelector{
						Number: uint32(servicePort.Port),
					},
				},
				Weight: 100,
			},
		},
	}
}

func getHTTPRoute(service *resolved.Service, host *string) *v1alpha3.HTTPRoute {
	matches := []*v1alpha3.HTTPMatchRequest{
		{
			Headers: map[string]*v1alpha3.StringMatch{
				"x-kardinal-destination": {
					MatchType: &v1alpha3.StringMatch_Exact{
						Exact: service.ServiceID + "-" + service.Version,
					},
				},
			},
		},
	}

	if host != nil {
		matches = append(matches, &v1alpha3.HTTPMatchRequest{
			Headers: map[string]*v1alpha3.StringMatch{
				"x-kardinal-destination": {
					MatchType: &v1alpha3.StringMatch_Exact{
						Exact: *host + "-" + service.Version,
					},
				},
			},
		})
	}

	return &v1alpha3.HTTPRoute{
		Match: matches,
		Route: []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   service.ServiceID,
					Subset: service.Version,
				},
			},
		},
	}
}

func getVirtualService(serviceID string, services []*resolved.Service, namespace string, gateway *string, extHost *string) (*istioclient.VirtualService, *istioclient.DestinationRule) {
	extHosts := []string{}

	httpRoutes := []*v1alpha3.HTTPRoute{}
	tcpRoutes := []*v1alpha3.TCPRoute{}
	destinationRule := getDestinationRule(serviceID, services, namespace)

	for _, service := range services {
		// TODO: Support for multiple ports
		servicePort := &service.ServiceSpec.Ports[0]
		var flowHost *string
		if extHost != nil {
			flowHostTemp := resolved.ReplaceOrAddSubdomain(*extHost, service.Version)
			flowHost = &flowHostTemp
			extHosts = append(extHosts, *flowHost)
		}

		if servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP" {
			httpRoutes = append(httpRoutes, getHTTPRoute(service, flowHost))
		} else {
			tcpRoutes = append(tcpRoutes, getTCPRoute(service, servicePort))
		}
	}

	virtualServiceSpec := v1alpha3.VirtualService{}
	virtualServiceSpec.Http = httpRoutes
	virtualServiceSpec.Tcp = tcpRoutes

	if gateway != nil {
		virtualServiceSpec.Gateways = []string{*gateway}
	}

	if extHost != nil {
		virtualServiceSpec.Hosts = extHosts
	} else {
		virtualServiceSpec.Hosts = []string{serviceID}
	}

	return &istioclient.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceID,
			Namespace: namespace,
		},
		Spec: virtualServiceSpec,
	}, destinationRule
}

func getDestinationRule(serviceID string, services []*resolved.Service, namespace string) *istioclient.DestinationRule {
	// TODO(shared-annotation) - we could store "shared" versions somewhere so that the pointers are the same
	// if we do that then the render work around isn't necessary
	subsets := lo.UniqBy(
		lo.Map(services, func(service *resolved.Service, _ int) *v1alpha3.Subset {

			newSubset := &v1alpha3.Subset{
				Name: service.Version,
				Labels: map[string]string{
					"version": service.Version,
				},
			}

			// TODO Narrow down this configuration to only subsets created for telepresence intercepts or find a way to enable TLS for telepresence intercepts https://github.com/kurtosis-tech/kardinal-kontrol/issues/14
			// This config is necessary for Kardinal/Telepresence (https://www.telepresence.io/) integration
			if service.Version != prodVersion {
				newTrafficPolicy := &v1alpha3.TrafficPolicy{
					Tls: &v1alpha3.ClientTLSSettings{
						Mode: v1alpha3.ClientTLSSettings_DISABLE,
					},
				}
				newSubset.TrafficPolicy = newTrafficPolicy
			}

			return newSubset
		}),
		func(subset *v1alpha3.Subset) string {
			return subset.Name
		},
	)

	return &istioclient.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceID,
			Namespace: namespace,
		},
		Spec: v1alpha3.DestinationRule{
			Host:    serviceID,
			Subsets: subsets,
		},
	}
}

func getService(service *resolved.Service, namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ServiceID,
			Namespace: namespace,
			Labels: map[string]string{
				"app": service.ServiceID,
			},
		},
		Spec: *service.ServiceSpec,
	}
}

func getDeployment(service *resolved.Service, namespace string) *appsv1.Deployment {
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.ServiceID, service.Version),
			Namespace: namespace,
			Labels: map[string]string{
				"app":     service.ServiceID,
				"version": service.Version,
			},
		},
		Spec: *service.DeploymentSpec,
	}

	numReplicas := int32(1)
	deployment.Spec.Replicas = int32Ptr(numReplicas)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}
	vol25pct := intstr.FromString("25%")
	deployment.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       &vol25pct,
			MaxUnavailable: &vol25pct,
		},
	}
	deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Annotations: map[string]string{
			"sidecar.istio.io/inject": "true",
		},
		Labels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}

	return &deployment
}

func getGateway(ingresses []*resolved.Ingress, namespace string) *istioclient.Gateway {
	extHosts := []string{}
	for _, ingress := range ingresses {
		ingressHost := ingress.GetHost()
		if ingressHost != nil {
			allFlowHosts := lo.Map(ingress.ActiveFlowIDs, func(flowId string, _ int) string { return resolved.ReplaceOrAddSubdomain(*ingressHost, flowId) })
			extHosts = append(extHosts, allFlowHosts...)
		}
	}
	extHosts = lo.Uniq(extHosts)

	// We need to return a gateway as part of the cluster resources so we return a dummy one
	// if there are no ingresses defined.  This can happen when the tenant does not have a base
	// cluster topology: no initial deploy or the topologies have been deleted.
	ingressId := "dummy"
	if len(ingresses) > 0 {
		ingressId = ingresses[0].IngressID
	} else {
		extHosts = []string{"dummy.kardinal.dev"}
	}

	return &istioclient.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressId,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     ingressId,
				"version": "v1",
			},
		},
		Spec: v1alpha3.Gateway{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: extHosts,
				},
			},
		},
	}
}

func getEnvoyFilters(service *resolved.Service, namespace string) []istioclient.EnvoyFilter {
	if !service.IsHTTP() {
		return []istioclient.EnvoyFilter{}
	}
	inboundFilter := &istioclient.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-inbound-trace-id-check", service.ServiceID),
			Namespace: namespace,
		},
		Spec: v1alpha3.EnvoyFilter{
			WorkloadSelector: &v1alpha3.WorkloadSelector{
				Labels: map[string]string{
					"app": service.ServiceID,
				},
			},
			ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
					Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
						ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: &v1alpha3.EnvoyFilter_Patch{
						Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
						Value: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"name": {Kind: &structpb.Value_StringValue{StringValue: "envoy.lua"}},
								"typed_config": {
									Kind: &structpb.Value_StructValue{
										StructValue: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"@type":      {Kind: &structpb.Value_StringValue{StringValue: luaFilterType}},
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: inboundRequestTraceIDFilter}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	outboundFilter := &istioclient.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-outbound-trace-router", service.ServiceID),
			Namespace: namespace,
		},
		Spec: v1alpha3.EnvoyFilter{
			WorkloadSelector: &v1alpha3.WorkloadSelector{
				Labels: map[string]string{
					"app": service.ServiceID,
				},
			},
			ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
					Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3.EnvoyFilter_SIDECAR_OUTBOUND,
						ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: &v1alpha3.EnvoyFilter_Patch{
						Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
						Value: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"name": {Kind: &structpb.Value_StringValue{StringValue: "envoy.lua"}},
								"typed_config": {
									Kind: &structpb.Value_StructValue{
										StructValue: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"@type":      {Kind: &structpb.Value_StringValue{StringValue: luaFilterType}},
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: getOutgoingRequestTraceIDFilter()}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return []istioclient.EnvoyFilter{*inboundFilter, *outboundFilter}
}

// getAuthorizationPolicy returns an authorization policy that denies requests with the missing header
// this is not really needed as we have an inbound rule
func getAuthorizationPolicy(service *resolved.Service, namespace string) *securityv1beta1.AuthorizationPolicy {
	if !service.IsHTTP() {
		return nil
	}
	return &securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-require-trace-id", service.ServiceID),
			Namespace: namespace,
		},
		Spec: securityapi.AuthorizationPolicy{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app": service.ServiceID,
				},
			},
			Action: securityapi.AuthorizationPolicy_DENY,
			Rules: []*securityapi.Rule{
				{
					When: []*securityapi.Condition{
						{
							Key:       "request.headers[x-kardinal-trace-id]",
							NotValues: []string{"*"},
						},
					},
				},
			},
		},
	}
}

func getEnvoyFilterForGateway(servicesAgainstVersions map[string][]string, serviceAndVersionAgainstExtHost map[string]string, serviceAgainstBackupVersions map[string][]string) *istioclient.EnvoyFilter {
	luaScript := generateDynamicLuaScript(servicesAgainstVersions, serviceAndVersionAgainstExtHost, serviceAgainstBackupVersions)

	return &istioclient.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kardinal-gateway-tracing",
			Namespace: "istio-system",
		},
		Spec: v1alpha3.EnvoyFilter{
			WorkloadSelector: &v1alpha3.WorkloadSelector{
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
					Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3.EnvoyFilter_GATEWAY,
						ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
									},
								},
							},
						},
					},
					Patch: &v1alpha3.EnvoyFilter_Patch{
						Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
						Value: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"name": {Kind: &structpb.Value_StringValue{StringValue: "envoy.lua"}},
								"typed_config": {
									Kind: &structpb.Value_StructValue{
										StructValue: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"@type":      {Kind: &structpb.Value_StringValue{StringValue: luaFilterType}},
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: luaScript}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateDynamicLuaScript(servicesAgainstVersions map[string][]string, versionAgainstExtHost map[string]string, serviceAgainstBackupVersions map[string][]string) string {
	var setRouteCalls strings.Builder

	// Helper function to add a setRoute call
	addSetRouteCall := func(service, destination string) {
		setRouteCalls.WriteString(fmt.Sprintf(`
    request_handle:httpCall(
      "outbound|8080||trace-router.default.svc.cluster.local",
      {
        [":method"] = "POST",
        [":path"] = "/set-route?trace_id=" .. trace_id .. "&hostname=%s&destination=%s",
        [":authority"] = "trace-router.default.svc.cluster.local",
        ["Content-Type"] = "application/json"
      },
      "{}",
      5000
    )
`, service, destination))
	}

	// For prod host all goes to prod
	// For non prod host we add a non prod mapping for the non prod service and everything else falls back
	for service, versions := range servicesAgainstVersions {
		for _, version := range versions {
			if version == constants.SharedVersionVersionString {
				continue
			}
			setRouteCalls.WriteString(fmt.Sprintf(`
    if hostname == "%s" then`, versionAgainstExtHost[version]))
			destination := fmt.Sprintf("%s-%s", service, version)
			addSetRouteCall(service, destination)
			setRouteCalls.WriteString(`
    end`)
		}
	}

	// we handle shared versions separately by finding host markings for original version
	for service, originalVersions := range serviceAgainstBackupVersions {
		for _, originalVersion := range originalVersions {
			setRouteCalls.WriteString(fmt.Sprintf(`
    if hostname == "%s" then`, versionAgainstExtHost[originalVersion]))
			destination := fmt.Sprintf("%s-%s", service, constants.SharedVersionVersionString)
			addSetRouteCall(service, destination)
			setRouteCalls.WriteString(`
    end`)
		}
	}

	// this gets consumed by the virtual source route for the gateway
	for version, extHost := range versionAgainstExtHost {
		destination := fmt.Sprintf("%s-%s", extHost, version)
		addSetRouteCall(extHost, destination)
	}

	return fmt.Sprintf(`
%s

function get_trace_id(headers)
  for _, header_name in ipairs(trace_header_priorities) do
    local trace_id = headers:get(header_name)
    if trace_id then
      return trace_id, header_name
    end
  end

  return nil, nil
end

function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id = headers:get("x-kardinal-trace-id")
  local hostname = headers:get(":authority")

  request_handle:logInfo("Processing request - Initial trace ID: " .. (trace_id or "none") .. ", Hostname: " .. (hostname or "none"))

  if not trace_id then
    local found_trace_id, source_header = get_trace_id(headers)
    if found_trace_id then
      trace_id = found_trace_id
      request_handle:logInfo("Using existing trace ID from " .. source_header .. ": " .. trace_id)
    else
      local generate_headers, generate_body = request_handle:httpCall(
        "outbound|8080||trace-router.default.svc.cluster.local",
        {
         [":method"] = "GET",
         [":path"] = "/generate-trace-id",
         [":authority"] = "trace-router.default.svc.cluster.local"
        },
        "",
        5000
      )

      if generate_headers and generate_headers[":status"] == "200" then
        trace_id = generate_body
        request_handle:logInfo("Received trace ID from trace-router: " .. trace_id)
      else
        trace_id = string.format("%%032x", math.random(2^128 - 1))
        request_handle:logWarn("Failed to get trace ID from trace-router, using locally generated: " .. trace_id)
      end
    end

    request_handle:headers():add("x-kardinal-trace-id", trace_id)

    %s
  end

  local determine_headers, determine_body = request_handle:httpCall(
    "outbound|8080||trace-router.default.svc.cluster.local",
    {
      [":method"] = "GET",
      [":path"] = "/route?trace_id=" .. trace_id .. "&hostname=" .. hostname,
      [":authority"] = "trace-router.default.svc.cluster.local"
    },
    "",
    5000
  )

  local destination
  if determine_headers and determine_headers[":status"] == "200" then
    destination = determine_body
    request_handle:logInfo("Determined destination: " .. destination)
  else
    destination = hostname .. "-prod"
    request_handle:logWarn("Failed to determine destination, using fallback: " .. destination)
  end

  request_handle:headers():add("x-kardinal-destination", destination)
  request_handle:logInfo("Final headers - Trace ID: " .. trace_id .. ", Destination: " .. destination)
end
`, generateLuaTraceHeaderPriorities(), setRouteCalls.String())
}
