package flow

import (
	"fmt"
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

	allActiveFlows := lo.FlatMap(clusterTopology.Ingress, func(item *resolved.Ingress, _ int) []string { return item.ActiveFlowIDs })
	var versionsAgainstExtHost map[string]string

	groupedServices := lo.GroupBy(clusterTopology.Services, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		servicesAgainstVersions[serviceID] = lo.Map(services, func(item *resolved.Service, _ int) string { return item.Version })
		if len(services) > 0 {
			// TODO: this assumes service specs didn't change. May we need a new version to ClusterTopology data structure
			serviceList = append(serviceList, *getService(services[0], namespace))

			var gateway *string
			var extHost *string
			if ingressService, found := clusterTopology.GetIngressForService(services[0]); found {
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

	// TODO: make it to use a list of Ingresses
	gatewayFilter := getEnvoyFilterForGateway(servicesAgainstVersions, versionsAgainstExtHost)
	envoyFilters = append(envoyFilters, *gatewayFilter)

	return types.ClusterResources{
		Services: serviceList,

		Deployments: lo.Map(clusterTopology.Services, func(service *resolved.Service, _ int) appsv1.Deployment {
			return *getDeployment(service, namespace)
		}),

		Gateway: *getGateway(clusterTopology.Ingress, namespace),

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
		}},
		// TODO(edgar) - do we need the version here?
		Route: []*v1alpha3.RouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: service.ServiceID,
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
	subsets := lo.Map(services, func(service *resolved.Service, _ int) *v1alpha3.Subset {
		return &v1alpha3.Subset{
			Name: service.Version,
			Labels: map[string]string{
				"version": service.Version,
			},
		}
	})

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
			APIVersion: "v1",
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

	ingressId := ingresses[0].IngressID

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
	if !isHttp(service) {
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
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: outgoingRequestTraceIDFilter}},
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
	if !isHttp(service) {
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

func getEnvoyFilterForGateway(servicesAgainstVersions map[string][]string, serviceAndVersionAgainstExtHost map[string]string) *istioclient.EnvoyFilter {
	luaScript := generateDynamicLuaScript(servicesAgainstVersions, serviceAndVersionAgainstExtHost)

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

func generateDynamicLuaScript(servicesAgainstVersions map[string][]string, versionAgainstExtHost map[string]string) string {
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

	// This is the one we are interested in
	// For prod host all goes to prod
	// For non prod host we add a non prod mapping for the non prod service and everything else falls back
	for service, versions := range servicesAgainstVersions {
		for _, version := range versions {
			setRouteCalls.WriteString(fmt.Sprintf(`
    if hostname == "%s" then`, versionAgainstExtHost[version]))
			destination := fmt.Sprintf("%s-%s", service, version)
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
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id = headers:get("x-kardinal-trace-id")
  local hostname = headers:get(":authority")

  request_handle:logInfo("Processing request - Initial trace ID: " .. (trace_id or "none") .. ", Hostname: " .. (hostname or "none"))

  if not trace_id then
    trace_id = string.format("%%032x", math.random(2^128 - 1))
    request_handle:logInfo("Generated new trace ID: " .. trace_id)

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
      request_handle:logWarn("Failed to get trace ID from trace-router, using generated: " .. trace_id)
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
`, setRouteCalls.String())
}

// isHttp - TODO support multiple ports
func isHttp(service *resolved.Service) bool {
	servicePort := &service.ServiceSpec.Ports[0]
	return servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP"
}