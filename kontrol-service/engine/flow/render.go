package flow

import (
	"fmt"
	"regexp"
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
	net "k8s.io/api/networking/v1"
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

	targeIngressServices := lo.Uniq(
		lo.FlatMap(clusterTopology.Ingress.Ingresses, func(ing net.Ingress, _ int) []string {
			return lo.FlatMap(ing.Spec.Rules, func(rule net.IngressRule, _ int) []string {
				return lo.FilterMap(rule.HTTP.Paths, func(path net.HTTPIngressPath, _ int) (string, bool) {
					// TODO: we are ignoring the namespace from the path, should we?
					targetNS := namespace
					if ing.Namespace != "" {
						targetNS = string(ing.Namespace)
					}
					if path.Backend.Service == nil {
						logrus.Errorf("Ingress %v has a nil backend service", ing.Name)
						return "", false
					}
					return fmt.Sprintf("%s/%s", targetNS, string(path.Backend.Service.Name)), true
				})
			})
		}))

	targetServices := lo.Uniq(append(targetHttpRouteServices, targeIngressServices...))

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

			virtualService, destinationRule := getVirtualService(serviceID, services, namespace)
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

	envoyFiltersForService := getEnvoyFilters(clusterTopology.Services, namespace, targetServices)
	envoyFilters = append(envoyFilters, envoyFiltersForService...)

	routes, frontServices, inboundFrontFilters := getHTTPRoutes(clusterTopology.GatewayAndRoutes, clusterTopology.Services, namespace)
	serviceList = append(serviceList, frontServices...)
	envoyFilters = append(envoyFilters, inboundFrontFilters...)

	ingresses, frontServices, inboundFrontFilters := getIngresses(clusterTopology.Ingress, clusterTopology.Services, namespace)
	serviceList = append(serviceList, frontServices...)
	envoyFilters = append(envoyFilters, inboundFrontFilters...)

	logrus.Infof("have total of %d envoy filters", len(envoyFilters))

	return types.ClusterResources{
		Services: serviceList,

		Deployments: lo.FilterMap(clusterTopology.Services, func(service *resolved.Service, _ int) (appsv1.Deployment, bool) {
			// Deployment spec is nil for external services, don't need to add anything to cluster
			deployment := getDeployment(service, namespace)
			if deployment == nil {
				return appsv1.Deployment{}, false
			}
			return *deployment, true
		}),

		StatefulSets: lo.FilterMap(clusterTopology.Services, func(service *resolved.Service, _ int) (appsv1.StatefulSet, bool) {
			// StatefulSet spec is nil for external services, don't need to add anything to cluster
			statefulSet := getStatefulSet(service, namespace)
			if statefulSet == nil {
				return appsv1.StatefulSet{}, false
			}
			return *statefulSet, true
		}),

		Gateways: getGateways(clusterTopology.GatewayAndRoutes),

		HTTPRoutes: routes,

		Ingresses: ingresses,

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

func getVirtualService(serviceID string, services []*resolved.Service, namespace string) (*istioclient.VirtualService, *istioclient.DestinationRule) {
	httpRoutes := []*v1alpha3.HTTPRoute{}
	tcpRoutes := []*v1alpha3.TCPRoute{}
	destinationRule := getDestinationRule(serviceID, services, namespace)

	for _, service := range services {
		// TODO: Support for multiple ports
		servicePort := &service.ServiceSpec.Ports[0]
		var flowHost *string

		if servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP" {
			httpRoutes = append(httpRoutes, getHTTPRoute(service, flowHost))
		} else {
			tcpRoutes = append(tcpRoutes, getTCPRoute(service, servicePort))
		}
	}

	virtualServiceSpec := v1alpha3.VirtualService{}
	virtualServiceSpec.Http = httpRoutes
	virtualServiceSpec.Tcp = tcpRoutes
	virtualServiceSpec.Hosts = []string{serviceID}

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
	// the baseline topology (or prod topology) flow ID and flow version are equal to the namespace these three should use same value
	baselineFlowVersion := namespace
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
			if service.Version != baselineFlowVersion {
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

func getStatefulSet(service *resolved.Service, namespace string) *appsv1.StatefulSet {
	if !service.WorkloadSpec.IsStatefulSet() {
		return nil
	}

	statefulSet := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "statefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.ServiceID, service.Version),
			Namespace: namespace,
			Labels: map[string]string{
				"app":     service.ServiceID,
				"version": service.Version,
			},
		},
		Spec: *service.WorkloadSpec.GetStatefulSetSpec(),
	}

	numReplicas := int32(1)
	statefulSet.Spec.Replicas = int32Ptr(numReplicas)
	statefulSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app":     service.ServiceID,
			"version": service.Version,
		},
	}
	statefulSet.Spec.Template.ObjectMeta = metav1.ObjectMeta{
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

	return &statefulSet
}

func getDeployment(service *resolved.Service, namespace string) *appsv1.Deployment {
	if !service.WorkloadSpec.IsDeployment() {
		return nil
	}

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
		Spec: *service.WorkloadSpec.GetDeploymentSpec(),
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

func findBackendRefService(serviceName string, serviceVersion string, services []*resolved.Service) (*resolved.Service, bool) {
	return lo.Find(services, func(service *resolved.Service) bool {
		return service.ServiceID == serviceName && service.Version == serviceVersion
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
				"app":     service.ServiceID,
				"version": flowVersion,
			},
		},
		Spec: *serviceSpecCopy,
	}
}

func getIngresses(
	ingress *resolved.Ingress,
	allServices []*resolved.Service,
	namespace string,
) ([]net.Ingress, []v1.Service, []istioclient.EnvoyFilter) {
	ingressList := []net.Ingress{}
	frontServices := map[string]v1.Service{}
	filters := []istioclient.EnvoyFilter{}

	for _, ingressSpecOriginal := range ingress.Ingresses {
		ingressDefinition := ingressSpecOriginal.DeepCopy()
		newRules := []net.IngressRule{}

		for _, ruleOriginal := range ingressDefinition.Spec.Rules {
			for _, activeFlowID := range ingress.ActiveFlowIDs {
				logrus.Infof("Setting gateway route for active flow ID: %v", activeFlowID)

				newPaths := []net.HTTPIngressPath{}

				rule := ruleOriginal.DeepCopy()

				flowHostname := resolved.ReplaceOrAddSubdomain(rule.Host, activeFlowID)
				rule.Host = flowHostname
				hostnames := []string{
					flowHostname,
				}

				for _, pathOriginal := range ruleOriginal.HTTP.Paths {
					target, found := findBackendRefService(pathOriginal.Backend.Service.Name, activeFlowID, allServices)
					// fallback to baseline if backend not found at the active flow
					// the baseline topology (or prod topology) flow ID and flow version are equal to the namespace these three should use same value
					baselineFlowVersion := namespace
					if !found {
						target, found = findBackendRefService(pathOriginal.Backend.Service.Name, baselineFlowVersion, allServices)
					}
					if found {
						path := *pathOriginal.DeepCopy()
						idVersion := fmt.Sprintf("%s-%s", target.ServiceID, activeFlowID)
						_, serviceAlreadyAdded := frontServices[idVersion]
						if !serviceAlreadyAdded {
							frontServices[idVersion] = getVersionedService(target, activeFlowID, namespace)
							path.Backend.Service.Name = idVersion
							newPaths = append(newPaths, path)

							// Set Envoy FIlter for the service
							filter := &externalInboudFilter{
								filter: generateDynamicLuaScript(allServices, activeFlowID, namespace, hostnames),
								name:   strings.Join(hostnames, "-"),
							}
							inboundFilter := getInboundFilter(target.ServiceID, namespace, -1, &target.Version, filter)
							logrus.Debugf("Adding inbound filter to setup routing table for flow '%s' on service '%s', version '%s'", activeFlowID, target.ServiceID, target.Version)
							filters = append(filters, inboundFilter)
						}
					} else {
						logrus.Errorf("Backend service %v for Ingress %v not found", pathOriginal.Backend.Service.Name, ingressDefinition.Name)
					}
				}
				rule.HTTP.Paths = newPaths
				newRules = append(newRules, *rule)
			}
		}

		ingressDefinition.Spec.Rules = newRules

		if ingressDefinition.Namespace == "" {
			ingressDefinition.Namespace = namespace
		}

		ingressList = append(ingressList, *ingressDefinition)
	}

	return ingressList, lo.Values(frontServices), filters
}

func getHTTPRoutes(
	gatewayAndRoutes *resolved.GatewayAndRoutes,
	allServices []*resolved.Service,
	namespace string,
) ([]gateway.HTTPRoute, []v1.Service, []istioclient.EnvoyFilter) {
	routes := []gateway.HTTPRoute{}
	frontServices := map[string]v1.Service{}
	filters := []istioclient.EnvoyFilter{}

	for _, activeFlowID := range gatewayAndRoutes.ActiveFlowIDs {
		logrus.Infof("Setting gateway route for active flow ID: %v", activeFlowID)
		for routeId, routeSpecOriginal := range gatewayAndRoutes.GatewayRoutes {
			routeSpec := routeSpecOriginal.DeepCopy()

			routeSpec.Hostnames = lo.Map(routeSpec.Hostnames, func(hostname gateway.Hostname, _ int) gateway.Hostname {
				return gateway.Hostname(resolved.ReplaceOrAddSubdomain(string(hostname), activeFlowID))
			})

			for _, rule := range routeSpec.Rules {
				for refIx, ref := range rule.BackendRefs {
					target, found := findBackendRefService(string(ref.Name), activeFlowID, allServices)
					// fallback to baseline if backend not found at the active flow
					// the baseline topology (or prod topology) flow ID and flow version are equal to the namespace these three should use same value
					baselineFlowVersion := namespace
					if !found {
						target, found = findBackendRefService(string(ref.Name), baselineFlowVersion, allServices)
					}
					if found {
						idVersion := fmt.Sprintf("%s-%s", target.ServiceID, activeFlowID)
						_, serviceAlreadyAdded := frontServices[idVersion]
						if !serviceAlreadyAdded {
							frontServices[idVersion] = getVersionedService(target, activeFlowID, namespace)
							ref.Name = gateway.ObjectName(idVersion)
							rule.BackendRefs[refIx] = ref

							hostnames := lo.Map(routeSpec.Hostnames, func(item gateway.Hostname, _ int) string { return string(item) })
							// Set Envoy FIlter for the service
							filter := &externalInboudFilter{
								filter: generateDynamicLuaScript(allServices, activeFlowID, namespace, hostnames),
								name:   strings.Join(hostnames, "-"),
							}
							inboundFilter := getInboundFilter(target.ServiceID, namespace, -1, &target.Version, filter)
							logrus.Debugf("Adding inbound filter to setup routing table for flow '%s' on service '%s', version '%s'", activeFlowID, target.ServiceID, target.Version)
							filters = append(filters, inboundFilter)
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

	return routes, lo.Values(frontServices), filters
}

func getEnvoyFilters(
	allServices []*resolved.Service,
	namespace string,
	targetServices []string,
) []istioclient.EnvoyFilter {
	filters := []istioclient.EnvoyFilter{}

	// HttpRoute (workload) are applied at the serviceID level, not the serviceID-version level
	groupedServices := lo.GroupBy(allServices, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		if len(services) == 0 {
			continue
		}

		anyNonHttp := lo.SomeBy(services, func(service *resolved.Service) bool {
			return !service.IsHTTP()
		})
		if anyNonHttp {
			logrus.Infof("Service '%s' is not an HTTP service, skipping filters", serviceID)
			continue
		}

		isTargertService := slices.Contains(targetServices, fmt.Sprintf("%s/%s", namespace, serviceID))

		// more inbound EnvoyFilters for routing routing traffic on frontend services are added by the getHTTPRoutes function
		if isTargertService {
			logrus.Debugf("Adding inbound filter to enforce trace IDs for service '%s'", serviceID)
			inboundFilter := getInboundFilter(serviceID, namespace, 0, nil, &traceIdEnforcer{})
			filters = append(filters, inboundFilter)
		} else {
			logrus.Debugf("Adding inbound filter for inner service '%s'", serviceID)
			inboundFilter := getInboundFilter(serviceID, namespace, 0, nil, &innerInboundFilter{})
			filters = append(filters, inboundFilter)
		}

		outboundFilter := getOutboundFilter(serviceID, namespace)
		filters = append(filters, outboundFilter)
	}

	return filters
}

type luaFilter interface {
	getName() string
	getFilter() string
}

type innerInboundFilter struct{}

type externalInboudFilter struct {
	filter string
	name   string
}

type traceIdEnforcer struct{}

func (f *innerInboundFilter) getName() string {
	return "inbound-router"
}

func (f *innerInboundFilter) getFilter() string {
	return inboundRequestTraceIDFilter
}

func (f *externalInboudFilter) getName() string {
	id := regexp.MustCompile(`[^a-z0-9.-]`).ReplaceAllString(f.name, "")
	return "inbound-router-" + id + "-external"
}

func (f *externalInboudFilter) getFilter() string {
	return f.filter
}

func (f *traceIdEnforcer) getName() string {
	return "trace-id-enforcer"
}

func (f *traceIdEnforcer) getFilter() string {
	return generateTraceIDEnforcerLuaScript()
}

func getInboundFilter(serviceID, namespace string, priority int32, versionSelector *string, luaFilter luaFilter) istioclient.EnvoyFilter {
	labelSelector := map[string]string{
		"app": serviceID,
	}
	if versionSelector != nil {
		labelSelector["version"] = *versionSelector
	}

	if luaFilter == nil {
		luaFilter = &innerInboundFilter{}
	}
	filterName := luaFilter.getName()
	ids := []*string{&serviceID, versionSelector, &filterName}
	name := strings.Join(lo.FilterMap(ids, func(id *string, _ int) (string, bool) {
		if id != nil {
			return *id, true
		} else {
			return "", false
		}
	}), "-")

	return istioclient.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha3.EnvoyFilter{
			Priority: priority,
			WorkloadSelector: &v1alpha3.WorkloadSelector{
				Labels: labelSelector,
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
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: luaFilter.getFilter()}},
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

func getOutboundFilter(serviceID, namespace string) istioclient.EnvoyFilter {
	// the baseline topology (or prod topology) flow ID and flow version and host are equal to the namespace these four should use same value
	baselineHostName := namespace
	return istioclient.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-outbound-router", serviceID),
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
												"inlineCode": {Kind: &structpb.Value_StringValue{StringValue: getOutgoingRequestTraceIDFilter(baselineHostName)}},
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

func generateTraceIDEnforcerLuaScript() string {
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

  request_handle:logInfo("Enforcing trace ID header - Initial trace ID: " .. (trace_id or "none") .. ", Hostname: " .. (hostname or "none"))

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
  end

end
`, generateLuaTraceHeaderPriorities())
}

func generateDynamicLuaScript(allServices []*resolved.Service, flowId string, namespace string, hostnames []string) string {
	// fallback to baseline if backend not found at the active flow
	// the baseline topology (or prod topology) flow ID and flow version are equal to the namespace these three should use same value
	baselineFlowVersion := namespace
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

	groupedServices := lo.GroupBy(allServices, func(item *resolved.Service) string { return item.ServiceID })
	for serviceID, services := range groupedServices {
		if len(services) == 0 {
			continue
		}

		var service, fallbackService *resolved.Service
		for _, s := range services {
			if s.Version == flowId {
				service = s
			}
			if s.Version == baselineFlowVersion {
				fallbackService = s
			}
		}

		if service == nil {
			service = fallbackService
		}
		if service == nil {
			logrus.Errorf("No service found for '%s' for version '%s' or baseline '%s'. No routing can configured.", serviceID, flowId, baselineFlowVersion)
			continue
		}

		if service.IsShared {
			setRouteCalls.WriteString(fmt.Sprintf(`
    if hostname == "%s" then`, service.OriginalVersionIfShared))
			destination := fmt.Sprintf("%s-%s", service.ServiceID, constants.SharedVersionVersionString)
			addSetRouteCall(service.ServiceID, destination)
			setRouteCalls.WriteString(`
    end`)
		} else {
			for _, hostname := range hostnames {
				setRouteCalls.WriteString(fmt.Sprintf(`
    if hostname == "%s" then`, string(hostname)))
				destination := fmt.Sprintf("%s-%s", service.ServiceID, service.Version)
				addSetRouteCall(service.ServiceID, destination)
				setRouteCalls.WriteString(`
    end`)
			}
		}
	}

	return fmt.Sprintf(`
%s

function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id = headers:get("x-kardinal-trace-id")
  local hostname = headers:get(":authority")

  request_handle:logInfo("Setting routing table for flowId %s, trace ID: " .. (trace_id or "none") .. ", Hostname: " .. (hostname or "none"))

  if not trace_id then
    request_handle:logWarn("Missing trace ID from " .. source_header .. ", make sure traceId enforcer filter was apply before.")
  else
    %s
  end

end
`, generateLuaTraceHeaderPriorities(), flowId, setRouteCalls.String())
}
