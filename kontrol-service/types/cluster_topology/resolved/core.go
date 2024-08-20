package resolved

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mohae/deepcopy"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	"slices"

	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	"regexp"
)

type ClusterTopology struct {
	FlowID              string              `json:"flowID"`
	Ingresses           []*Ingress          `json:"ingress"`
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
	ActiveFlowIDs []string           `json:"activeFlowIDs"`
	IngressID     string             `json:"ingressID"`
	IngressRules  []*net.IngressRule `json:"ingressRules"`
	// IngressSpec and ServiceSpec are mutually exclusive
	// IngressSpec is set if a k8s Ingress type is being used for this Ingress
	// ServiceSpec is set if a k8s Service type is acting as an Ingress (eg. LoadBalancer, custom gateway)
	IngressSpec *net.IngressSpec    `json:"ingressSpec"`
	ServiceSpec *corev1.ServiceSpec `json:"serviceSpec"`
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

func (clusterTopology *ClusterTopology) IsIngressDestination(service *Service) bool {
	return lo.SomeBy(clusterTopology.Ingresses, func(item *Ingress) bool {
		return slices.Contains(item.GetTargetServices(), service.ServiceID)
	})
}

func (clusterTopology *ClusterTopology) GetIngressForService(service *Service) (*Ingress, bool) {
	// TODO: How to force that a service can't have more than one ingress?
	return lo.Find(clusterTopology.Ingresses, func(item *Ingress) bool {
		return slices.Contains(item.GetTargetServices(), service.ServiceID)
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

	if clusterTopology != nil && clusterTopology.Ingresses != nil {
		for _, ing := range clusterTopology.Ingresses {
			for key, value := range ing.GetFlowHostMapping() {
				result[key] = append(result[key], value)
			}
		}
	}
	return result
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

func (ingress *Ingress) GetTargetServices() []string {
	targetServices := []string{}
	if ingress.IngressSpec != nil {
		for _, rule := range ingress.IngressSpec.Rules {
			for _, httpPath := range rule.HTTP.Paths {
				targetServices = append(targetServices, httpPath.Backend.Service.Name)
			}
		}
	}
	if ingress.ServiceSpec != nil {
		appName, ok := ingress.ServiceSpec.Selector["app"]
		if ok {
			targetServices = append(targetServices, appName)
		}
	}
	return targetServices
}

func (service *Service) IsHTTP() bool {
	if len(service.ServiceSpec.Ports) == 0 {
		return false
	}
	servicePort := service.ServiceSpec.Ports[0]
	return servicePort.AppProtocol != nil && *servicePort.AppProtocol == "HTTP"
}
