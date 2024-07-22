package resolved

import (
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
	IngressUUID  string
	IngressRules []*net.IngressRule
}
