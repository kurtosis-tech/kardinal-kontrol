package kubernetes_client

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type KubernetesClient struct {
	config          *rest.Config
	clientSet       *kubernetes.Clientset
	dynamicClient   *dynamic.DynamicClient
	discoveryMapper *restmapper.DeferredDiscoveryRESTMapper
}

func newKubernetesClient(config *rest.Config, clientSet *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient, discoveryMapper *restmapper.DeferredDiscoveryRESTMapper) *KubernetesClient {
	return &KubernetesClient{config: config, clientSet: clientSet, dynamicClient: dynamicClient, discoveryMapper: discoveryMapper}
}

func (client *KubernetesClient) GetConfig() *rest.Config {
	return client.config
}

func (client *KubernetesClient) GetClientSet() *kubernetes.Clientset {
	return client.clientSet
}

func (client *KubernetesClient) GetDynamicClient() *dynamic.DynamicClient {
	return client.dynamicClient
}

func (client *KubernetesClient) GetDiscoveryMapper() *restmapper.DeferredDiscoveryRESTMapper {
	return client.discoveryMapper
}
