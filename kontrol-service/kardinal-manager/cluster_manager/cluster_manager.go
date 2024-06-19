package cluster_manager

import "kardinal.kontrol/kardinal-manager/kubernetes_client"

type ClusterManager struct {
	kubernetesClient *kubernetes_client.KubernetesClient
}
