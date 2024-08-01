package resolved

import (
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

func (ingress *Ingress) GetSelectorAppName() *string {
	appName, ok := ingress.ServiceSpec.Selector["app"]
	if ok {
		return &appName
	}

	return nil
}
