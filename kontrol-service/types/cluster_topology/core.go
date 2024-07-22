package cluster_topology

import (
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

type ClusterTopology struct {
	Ingress            Ingress
	Services           []Service
	ServiceDependecies []ServiceDependency
}

type Service struct {
	ServiceID      string
	ServiceSpec    *corev1.ServiceSpec
	IsExternal     bool
	IsStateful     bool
	StatefulPlugin string
}

type ServiceDependency struct {
	ServiceID          string
	DependsOnServiceID string
	DependencyPort     string
}

type Ingress struct {
	IngressUUID  string
	IngressRules []*net.IngressRule
}
