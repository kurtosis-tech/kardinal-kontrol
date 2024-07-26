package resolved

import (
	"github.com/kurtosis-tech/stacktrace"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

type ClusterTopology struct {
	FlowID             string
	Ingress            Ingress
	Services           []Service
	ServiceDependecies []ServiceDependency
}

type Service struct {
	ServiceID       string
	Version         string
	ServiceSpec     *corev1.ServiceSpec
	DeploymentSpec  *appsv1.DeploymentSpec
	IsExternal      bool
	IsStateful      bool
	StatefulPlugins []*StatefulPlugin
}

type ServiceDependency struct {
	Service          *Service
	DependsOnService *Service
	DependencyPort   *corev1.ServicePort
}

type Ingress struct {
	IngressID    string
	IngressRules []*net.IngressRule
	ServiceSpec  *corev1.ServiceSpec
}

func (clusterTopology *ClusterTopology) GetServiceAndPort(serviceName string, servicePortName string) (*Service, *corev1.ServicePort, error) {
	for _, service := range clusterTopology.Services {
		if service.ServiceID == serviceName {
			for _, port := range service.ServiceSpec.Ports {
				if port.Name == servicePortName {
					return &service, &port, nil
				}
			}
		}
	}

	return nil, nil, stacktrace.NewError("Service %s and Port %s not found in the list of services", serviceName, servicePortName)
}

func (clusterTopology *ClusterTopology) GetService(serviceName string) (*Service, error) {
	for _, service := range clusterTopology.Services {
		if service.ServiceID == serviceName {
			return &service, nil
		}
	}

	return nil, stacktrace.NewError("Service %s not found in the list of services", serviceName)
}

func (clusterTopology *ClusterTopology) IsIngressDestination(service *Service) bool {
	appName, ok := clusterTopology.Ingress.ServiceSpec.Selector["app"]
	if ok {
		if appName == service.ServiceID {
			return true
		}
	}

	return false
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
