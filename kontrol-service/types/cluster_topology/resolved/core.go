package resolved

import (
	"fmt"
	"regexp"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

type ClusterTopology struct {
	FlowID             string              `json:"flowID"`
	Ingress            []*Ingress          `json:"ingress"`
	Services           []*Service          `json:"services"`
	ServiceDependecies []ServiceDependency `json:"serviceDependencies"`
}

type Service struct {
	ServiceID       string                 `json:"serviceID"`
	Version         string                 `json:"version"`
	ServiceSpec     *corev1.ServiceSpec    `json:"serviceSpec"`
	DeploymentSpec  *appsv1.DeploymentSpec `json:"deploymentSpec"`
	IsExternal      bool                   `json:"isExternal"`
	IsStateful      bool                   `json:"isStateful"`
	StatefulPlugins []*StatefulPlugin      `json:"statefulPlugins"`
}

type ServiceDependency struct {
	Service          *Service            `json:"service"`
	DependsOnService *Service            `json:"dependsOnService"`
	DependencyPort   *corev1.ServicePort `json:"dependencyPort"`
}

type Ingress struct {
	ActiveFlowIDs []string            `json:"activeFlowIDs"`
	IngressID     string              `json:"ingressID"`
	IngressRules  []*net.IngressRule  `json:"ingressRules"`
	ServiceSpec   *corev1.ServiceSpec `json:"serviceSpec"`
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

func (clusterTopology *ClusterTopology) IsIngressDestination(service *Service) bool {
	return lo.SomeBy(clusterTopology.Ingress, func(item *Ingress) bool {
		appName, ok := item.ServiceSpec.Selector["app"]
		if ok {
			return appName == service.ServiceID
		}
		return false
	})
}

func (clusterTopology *ClusterTopology) GetIngressForService(service *Service) (*Ingress, bool) {
	// TODO: How to force that a service can't have more than one ingress?
	return lo.Find(clusterTopology.Ingress, func(item *Ingress) bool {
		appName, ok := item.ServiceSpec.Selector["app"]
		if ok {
			return appName == service.ServiceID
		}
		return false
	})
}

func (ingress *Ingress) GetHost() *string {
	if len(ingress.IngressRules) > 0 {
		return &ingress.IngressRules[0].Host
	}

	return nil
}

func (clusterTopology *ClusterTopology) GetFlowHostMapping() map[string][]string {
	result := make(map[string][]string)

	if clusterTopology != nil && clusterTopology.Ingress != nil {
		for _, ing := range clusterTopology.Ingress {
			for key, value := range ing.GetFlowHostMapping() {
				result[key] = append(result[key], value)
			}
		}
	}
	return result
}

func (ingress *Ingress) GetFlowHostMapping() map[string]string {
	mapping := make(map[string]string)
	if len(ingress.IngressRules) > 0 {
		baseHost := &ingress.IngressRules[0].Host
		for _, flowID := range ingress.ActiveFlowIDs {
			mapping[flowID] = ReplaceOrAddSubdomain(*baseHost, flowID)
		}
	}

	return mapping
}

func ReplaceOrAddSubdomain(url string, newSubdomain string) string {
	re := regexp.MustCompile(`^(https?://)?(([^./]+\.)?([^./]+\.[^./]+))(.*)$`)
	return re.ReplaceAllString(url, fmt.Sprintf("${1}%s.${4}${5}", newSubdomain))
}

func (ingress *Ingress) GetSelectorAppName() *string {
	appName, ok := ingress.ServiceSpec.Selector["app"]
	if ok {
		return &appName
	}

	return nil
}

func (service *Service) IsHTTP() bool {
	if len(service.ServiceSpec.Ports) == 0 {
		return false
	}
	servicePort := service.ServiceSpec.Ports[0]
	return servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP"
}
