package flow

import (
	"fmt"
	"slices"
	"strings"

	"kardinal.kontrol-service/constants"
	"kardinal.kontrol-service/types"
	"kardinal.kontrol-service/types/cluster_topology/resolved"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/intstr"

	securityapi "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"
)

// RenderClusterResources returns a cluster resource for a given topology
// Perhaps we can make this throw an error if the # of extHosts != # of versions
// This assumes that there is a dev version of the ext host as well
func RenderClusterResources(clusterTopology *resolved.ClusterTopology, namespace string) types.ClusterResources {
	virtualServices := []istioclient.VirtualService{}
	destinationRules := []istioclient.DestinationRule{}
	envoyFilters := []istioclient.EnvoyFilter{}

	authorizationPolicies := []securityv1beta1.AuthorizationPolicy{}

	serviceList := []v1.Service{}

	targetHttpRouteServices := lo.Uniq(
		lo.FlatMap(clusterTopology.GatewayAndRoutes.GatewayRoutes, func(item *gateway.HTTPRouteSpec, _ int) []string {
			return lo.FlatMap(item.Rules, func(rule gateway.HTTPRouteRule, _ int) []string {
				return lo.Map(rule.BackendRefs, func(ref gateway.HTTPBackendRef, _ int) string {
					// TODO: we are ignoring the namespace from the ref, should we?
					targetNS := namespace
					if ref.Namespace != nil {
						targetNS = string(*ref.Namespace)
					}
					return fmt.Sprintf("%s/%s", targetNS, string(ref.Name))
				})
			})
		}))

	groupedServices := lo.GroupBy(clusterTopology.Services, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		logrus.Infof("Rendering service with id: '%v'.", serviceID)
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
			}

			virtualService, destinationRule := getVirtualService(serviceID, services, namespace, gateway, extHost)
			virtualServices = append(virtualServices, *virtualService)
			if destinationRule != nil {
				destinationRules = append(destinationRules, *destinationRule)
			}
			logrus.Infof("adding filters and authorization policies for service '%s'", serviceID)

			authorizationPolicy := getAuthorizationPolicy(services[0], namespace)
			if authorizationPolicy != nil {
				authorizationPolicies = append(authorizationPolicies, *authorizationPolicy)
			}
		}
	}

	envoyFiltersForService := getEnvoyFilters(clusterTopology.Services, namespace, targetHttpRouteServices)
	envoyFilters = append(envoyFilters, envoyFiltersForService...)
	logrus.Infof("have total of %d envoy filters", len(envoyFilters))

	routes, frontServices := getHTTPRoutes(clusterTopology.GatewayAndRoutes, clusterTopology.Services, namespace)

	return types.ClusterResources{
		Services: append(serviceList, frontServices...),

		Deployments: lo.FilterMap(clusterTopology.Services, func(service *resolved.Service, _ int) (appsv1.Deployment, bool) {
			// Deployment spec is nil for external services, don't need to add anything to cluster
			if service.DeploymentSpec == nil {
				return appsv1.Deployment{}, false
			}
			return *getDeployment(service, namespace), true
		}),

		Gateways: getGateways(clusterTopology.GatewayAndRoutes),

		HTTPRoutes: routes,

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
			// TODO: make this a flag to help debugging
			// One can view the logs with: kubeclt logs -f -l app=<serviceID> -n <namespace> -c istio-proxy
			"sidecar.istio.io/componentLogLevel": "lua:info",
		},
		Labels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}

	return &deployment
}

func getGateways(gatewayAndRoutes *resolved.GatewayAndRoutes) []gateway.Gateway {
	return lo.Map(gatewayAndRoutes.Gateways, func(gateway *gateway.Gateway, gwId int) gateway.Gateway {
		if gateway.Namespace == "" {
			gateway.Namespace = "default"
		}
		return *gateway
	})
}

func findBackendRefService(backendRef gateway.HTTPBackendRef, serviceVersion string, services []*resolved.Service) (*resolved.Service, bool) {
	return lo.Find(services, func(service *resolved.Service) bool {
		return service.ServiceID == string(backendRef.Name) && service.Version == serviceVersion
	})
}

func getVersionedService(service *resolved.Service, flowVersion string, namespace string) v1.Service {
	serviceSpecCopy := service.ServiceSpec.DeepCopy()

	serviceSpecCopy.Selector = map[string]string{
		"app":     service.ServiceID,
		"version": service.Version,
	}

	return v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.ServiceID, flowVersion),
			Namespace: namespace,
			Labels: map[string]string{
				"app": service.ServiceID,
			},
		},
		Spec: *serviceSpecCopy,
	}
}

func getHTTPRoutes(
	gatewayAndRoutes *resolved.GatewayAndRoutes,
	services []*resolved.Service,
	namespace string,
) ([]gateway.HTTPRoute, []v1.Service) {
	routes := []gateway.HTTPRoute{}
	frontServices := map[string]v1.Service{}

	for _, activeFlowID := range gatewayAndRoutes.ActiveFlowIDs {
		logrus.Infof("Setting gateway route for active flow ID: %v", activeFlowID)
		for routeId, routeSpecOriginal := range gatewayAndRoutes.GatewayRoutes {
			routeSpec := routeSpecOriginal.DeepCopy()

			for _, rule := range routeSpec.Rules {
				for refIx, ref := range rule.BackendRefs {
					target, found := findBackendRefService(ref, activeFlowID, services)
					// fallback to prod if backend not found at the active flow
					if !found {
						target, found = findBackendRefService(ref, "prod", services)
					}
					if found {
						idVersion := fmt.Sprintf("%s-%s", target.ServiceID, activeFlowID)
						_, serviceAlreadyAdded := frontServices[idVersion]
						if !serviceAlreadyAdded {
							frontServices[idVersion] = getVersionedService(target, activeFlowID, namespace)
							ref.Name = gateway.ObjectName(idVersion)
							rule.BackendRefs[refIx] = ref
						}
					} else {
						logrus.Errorf(">> service not found %v", ref.Name)
					}
				}
			}

			for parentRefIx, parentRef := range routeSpec.ParentRefs {
				if parentRef.Namespace == nil || string(*parentRef.Namespace) == "" {
					defaultNS := gateway.Namespace("default")
					parentRef.Namespace = &defaultNS
				}
				routeSpec.ParentRefs[parentRefIx] = parentRef
			}
			routeSpec.Hostnames = lo.Map(routeSpec.Hostnames, func(hostname gateway.Hostname, _ int) gateway.Hostname {
				return gateway.Hostname(resolved.ReplaceOrAddSubdomain(string(hostname), activeFlowID))
			})

			route := gateway.HTTPRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "gateway.networking.k8s.io/v1",
					Kind:       "HTTPRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("http-route-%d-%s", routeId, activeFlowID),
					Namespace: namespace,
				},
				Spec: *routeSpec,
			}
			routes = append(routes, route)
		}
	}

	return routes, lo.Values(frontServices)
}

func getEnvoyFilters(
	allServices []*resolved.Service,
	namespace string,
	targetHttpRouteServices []string,
) []istioclient.EnvoyFilter {
	luaFilterFrontend := generateDynamicLuaScript(allServices)

	filters := []istioclient.EnvoyFilter{}

	for _, service := range allServices {
		if service == nil || !service.IsHTTP() {
			continue
		}

		serviceID := service.ServiceID
		isTargertService := slices.Contains(targetHttpRouteServices, fmt.Sprintf("%s/%s", namespace, serviceID))

		luaFilter := inboundRequestTraceIDFilter
		if isTargertService {
			luaFilter = luaFilterFrontend
		}

		inboundFilter := istioclient.EnvoyFilter{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "EnvoyFilter",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-inbound-trace-id-check", serviceID),
				Namespace: namespace,
			},
			Spec: v1alpha3.EnvoyFilter{
				WorkloadSelector: &v1alpha3.WorkloadSelector{
					Labels: map[string]string{
						"app": serviceID,
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
													"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: luaFilter}},
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

		outboundFilter := istioclient.EnvoyFilter{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "EnvoyFilter",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-outbound-trace-router", serviceID),
				Namespace: namespace,
			},
			Spec: v1alpha3.EnvoyFilter{
				WorkloadSelector: &v1alpha3.WorkloadSelector{
					Labels: map[string]string{
						"app": serviceID,
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
		filters = append(filters, inboundFilter, outboundFilter)
	}

	return filters
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

func generateDynamicLuaScript(services []*resolved.Service) string {
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

	for _, service := range services {
		if service == nil {
			continue
		}

		if service.IsShared {
			setRouteCalls.WriteString(fmt.Sprintf(`\n    if hostname == "%s" then`, service.OriginalVersionIfShared))
			destination := fmt.Sprintf("%s-%s", service.ServiceID, constants.SharedVersionVersionString)
			addSetRouteCall(service.ServiceID, destination)
			setRouteCalls.WriteString(`\n    end`)
		} else {
			destination := fmt.Sprintf("%s-%s", service.ServiceID, service.Version)
			addSetRouteCall(service.ServiceID, destination)
		}
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

end
`, generateLuaTraceHeaderPriorities(), setRouteCalls.String())
}
