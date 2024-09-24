package resolved

import (
	"fmt"
	"regexp"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/mohae/deepcopy"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	gateway "sigs.k8s.io/gateway-api/apis/v1"
)

type ClusterTopology struct {
	FlowID              string              `json:"flowID"`
	GatewayAndRoutes    *GatewayAndRoutes   `json:"gatewayAndRoutes"`
	Ingress             *Ingress            `json:"ingress"`
	Services            []*Service          `json:"services"`
	ServiceDependencies []ServiceDependency `json:"serviceDependencies"`
	Namespace           string              `json:"namespace"`
}

type Service struct {
	ServiceID               string                 `json:"serviceID"`
	Version                 string                 `json:"version"`
	ServiceSpec             *corev1.ServiceSpec    `json:"serviceSpec"`
	DeploymentSpec          *appsv1.DeploymentSpec `json:"deploymentSpec"`
	IsExternal              bool                   `json:"isExternal"`
	IsStateful              bool                   `json:"isStateful"`
	StatefulPlugins         []*StatefulPlugin      `json:"statefulPlugins"`
	IsShared                bool                   `json:"isShared"`
	OriginalVersionIfShared string                 `json:"originalVersionIfShared"`
}

type ServiceDependency struct {
	Service          *Service            `json:"service"`
	DependsOnService *Service            `json:"dependsOnService"`
	DependencyPort   *corev1.ServicePort `json:"dependencyPort"`
}

type Ingress struct {
	ActiveFlowIDs []string      `json:"activeFlowIDs"`
	Ingresses     []net.Ingress `json:"ingresses"`
}

type GatewayAndRoutes struct {
	ActiveFlowIDs []string                 `json:"activeFlowIDs"`
	Gateways      []*gateway.Gateway       `json:"gateway"`
	GatewayRoutes []*gateway.HTTPRouteSpec `json:"gatewayRoutes"`
}

type IngressAccessEntry struct {
	FlowID    string `json:"flowID"`
	Hostname  string `json:"hostname"`
	Service   string `json:"service"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
}

func (clusterTopology *ClusterTopology) GetServiceAndPort(serviceName string, servicePortName string) (*Service, *corev1.ServicePort, error) {
	for _, service := range clusterTopology.Services {
		if service.ServiceID == serviceName {
			for _, port := range service.ServiceSpec.Ports {
				if port.Name == servicePortName {
					return service, &port, nil
				}
			}
		}
	}

	return nil, nil, stacktrace.NewError("Service %s and Port %s not found in the list of services", serviceName, servicePortName)
}

func (clusterTopology *ClusterTopology) GetService(serviceName string) (*Service, error) {
	for _, service := range clusterTopology.Services {
		if service.ServiceID == serviceName {
			return service, nil
		}
	}

	return nil, stacktrace.NewError("Service %s not found in the list of services", serviceName)
}

func (clusterTopology *ClusterTopology) UpdateWithService(modifiedService *Service) error {
	for idx, service := range clusterTopology.Services {
		if service.ServiceID == modifiedService.ServiceID {
			clusterTopology.Services[idx] = modifiedService
			clusterTopology.UpdateDependencies(service, modifiedService)
			return nil
		}
	}

	return stacktrace.NewError("Service %s not found in the list of services", modifiedService.ServiceID)
}

func (clusterTopology *ClusterTopology) FindImmediateParents(service *Service) []*Service {
	parents := make([]*Service, 0)
	for _, dependency := range clusterTopology.ServiceDependencies {
		if dependency.DependsOnService.ServiceID == service.ServiceID {
			parents = append(parents, dependency.Service)
		}
	}
	return parents
}

func (clusterTopology *ClusterTopology) UpdateDependencies(targetService *Service, modifiedService *Service) {
	for ix, dependency := range clusterTopology.ServiceDependencies {
		if dependency.Service == targetService {
			dependency.Service = modifiedService
		}
		if dependency.DependsOnService == targetService {
			dependency.DependsOnService = modifiedService
		}
		clusterTopology.ServiceDependencies[ix] = dependency
	}
}

func (clusterTopology *ClusterTopology) MoveServiceToVersion(service *Service, version string) error {
	// Don't duplicate if its already duplicated
	duplicatedService := deepcopy.Copy(service).(*Service)
	duplicatedService.Version = version
	return clusterTopology.UpdateWithService(duplicatedService)
}

func ReplaceOrAddSubdomain(url string, newSubdomain string) string {
	re := regexp.MustCompile(`^(https?://)?(([^./]+\.)?([^./]+\.[^./]+))(.*)$`)
	return re.ReplaceAllString(url, fmt.Sprintf("${1}%s.${4}${5}", newSubdomain))
}

func (service *Service) IsHTTP() bool {
	if service == nil || service.ServiceSpec == nil || len(service.ServiceSpec.Ports) == 0 {
		return false
	}
	servicePort := service.ServiceSpec.Ports[0]
	return servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP"
}

func getIngressFlowHostMap(ingress *Ingress) map[string][]IngressAccessEntry {
	flowHostMapping := map[string][]IngressAccessEntry{}

	if ingress == nil {
		return flowHostMapping
	}

	for _, flowID := range ingress.ActiveFlowIDs {
		_, found := flowHostMapping[flowID]
		if !found {
			flowHostMapping[flowID] = []IngressAccessEntry{}
		}
		for _, ing := range ingress.Ingresses {
			for _, rule := range ing.Spec.Rules {
				host := ReplaceOrAddSubdomain(rule.Host, flowID)
				ns := "default"
				if ing.Namespace != "" {
					ns = ing.Namespace
				}
				entry := IngressAccessEntry{
					FlowID:    flowID,
					Hostname:  host,
					Service:   ing.Name,
					Namespace: ns,
					Type:      "ingress",
				}
				flowHostMapping[flowID] = append(flowHostMapping[flowID], entry)
			}
		}
	}

	return flowHostMapping
}

func getGatewayFlowHostMap(gw *GatewayAndRoutes) map[string][]IngressAccessEntry {
	flowHostMapping := map[string][]IngressAccessEntry{}

	if gw == nil {
		return flowHostMapping
	}

	for _, flowID := range gw.ActiveFlowIDs {
		_, found := flowHostMapping[flowID]
		if !found {
			flowHostMapping[flowID] = []IngressAccessEntry{}
		}
		for _, route := range gw.GatewayRoutes {
			for _, ref := range route.ParentRefs {
				for _, originalHost := range route.Hostnames {
					host := ReplaceOrAddSubdomain(string(originalHost), flowID)
					ns := "default"
					if ref.Namespace != nil {
						ns = string(*ref.Namespace)
					}
					entry := IngressAccessEntry{
						FlowID:    flowID,
						Hostname:  host,
						Service:   string(ref.Name),
						Namespace: ns,
						Type:      "gateway",
					}
					flowHostMapping[flowID] = append(flowHostMapping[flowID], entry)
				}
			}
		}
	}

	return flowHostMapping
}

func (clusterTopology *ClusterTopology) GetFlowHostMapping() map[string][]IngressAccessEntry {
	flowHostMapping := map[string][]IngressAccessEntry{}
	gatewayFlowHostMap := getGatewayFlowHostMap(clusterTopology.GatewayAndRoutes)
	ingressFlowHostMap := getIngressFlowHostMap(clusterTopology.Ingress)

	for flowID, entries := range gatewayFlowHostMap {
		imap, found := ingressFlowHostMap[flowID]
		if found {
			entries = append(entries, imap...)
		}
		flowHostMapping[flowID] = entries
	}
	for flowID, entries := range ingressFlowHostMap {
		_, found := flowHostMapping[flowID]
		if !found {
			flowHostMapping[flowID] = entries
		}
	}

	return flowHostMapping
}
