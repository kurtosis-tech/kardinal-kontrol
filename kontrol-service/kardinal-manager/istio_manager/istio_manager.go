package istio_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	"istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// IstIO ontology:
// - virtual services
// 	 - host
//   - routing rules
//   - destination rules
//
// - destination rules
// 		- hosts
//		- traffic policy
//		- subsets
//
// TODO: implement this ontology later
// - gateways
// - service entries

// use cases IstIo manager needs to support:
// - ability to configure traffic routing rules for services in a cluster
//   - change the distribution of traffic to a service
//   - redirect which service traffic is going to
//   - duplicate traffic to services
//
// - ability to add new versions of a service
//   - updating destination rules

type IstIoManager struct {
	istioClient *versioned.Clientset

	virtualServicesClient v1alpha3.VirtualServiceInterface

	destinationRulesClient v1alpha3.DestinationRuleInterface
}

func CreateIstIoManager(k8sConfig *rest.Config) (*IstIoManager, error) {
	ic, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating IstIo client from k8s config: %v", k8sConfig)
	}
	namespace := "default"
	vsClient := ic.NetworkingV1alpha3().VirtualServices(namespace)
	drClient := ic.NetworkingV1alpha3().DestinationRules(namespace)
	return &IstIoManager{
		istioClient:            ic,
		virtualServicesClient:  vsClient,
		destinationRulesClient: drClient,
	}, nil
}

func (iom *IstIoManager) GetVirtualServices(ctx context.Context) ([]*istionetworking.VirtualService, error) {
	virtualServiceList, err := iom.virtualServicesClient.List(ctx, metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving virtual services from IstIo client.")
	}
	return virtualServiceList.Items, nil
}

func (iom *IstIoManager) GetVirtualService(ctx context.Context, name string) (*istionetworking.VirtualService, error) {
	virtualService, err := iom.virtualServicesClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving virtual service '%s' from IstIo client", name)
	}
	return virtualService, nil
}

func (iom *IstIoManager) GetDestinationRules(ctx context.Context) ([]*istionetworking.DestinationRule, error) {
	destinationRules, err := iom.destinationRulesClient.List(ctx, metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
		SendInitialEvents:    nil,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retreiving destination rules.")
	}
	return destinationRules.Items, nil
}

func (iom *IstIoManager) GetDestinationRule(ctx context.Context, rule string) (*istionetworking.DestinationRule, error) {
	destinationRule, err := iom.destinationRulesClient.Get(ctx, rule, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving destination rule '%s' from IstIo client", rule)
	}
	return destinationRule, nil
}

func (iom *IstIoManager) CreateVirtualService() error {
	return nil
}

func (iom *IstIoManager) CreateDestinationRule() error {
	return nil
}

func (iom *IstIoManager) AddRoutingRule() error {
	return nil
}
