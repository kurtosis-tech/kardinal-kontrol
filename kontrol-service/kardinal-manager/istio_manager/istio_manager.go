package istio_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	istio "istio.io/api/networking/v1alpha3"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
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
// 		- host
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

type IstioManager struct {
	istioClient *versioned.Clientset

	virtualServicesClient v1alpha3.VirtualServiceInterface

	destinationRulesClient v1alpha3.DestinationRuleInterface
}

func CreateIstIoManager(k8sConfig *rest.Config) (*IstioManager, error) {
	ic, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating IstIo client from k8s config: %v", k8sConfig)
	}
	namespace := "default"
	vsClient := ic.NetworkingV1alpha3().VirtualServices(namespace)
	drClient := ic.NetworkingV1alpha3().DestinationRules(namespace)
	return &IstioManager{
		istioClient:            ic,
		virtualServicesClient:  vsClient,
		destinationRulesClient: drClient,
	}, nil
}

func (iom *IstioManager) GetVirtualServices(ctx context.Context) ([]*istio_networking.VirtualService, error) {
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

func (iom *IstioManager) GetVirtualService(ctx context.Context, name string) (*istio_networking.VirtualService, error) {
	virtualService, err := iom.virtualServicesClient.Get(ctx, name, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving virtual service '%s' from IstIo client", name)
	}
	return virtualService, nil
}

func (iom *IstioManager) GetDestinationRules(ctx context.Context) ([]*istio_networking.DestinationRule, error) {
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

func (iom *IstioManager) GetDestinationRule(ctx context.Context, rule string) (*istio_networking.DestinationRule, error) {
	destinationRule, err := iom.destinationRulesClient.Get(ctx, rule, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving destination rule '%s' from IstIo client", rule)
	}
	return destinationRule, nil
}

func (iom *IstioManager) CreateVirtualService(ctx context.Context, vs *istio_networking.VirtualService) error {
	_, err := iom.virtualServicesClient.Create(ctx, vs, metav1.CreateOptions{
		TypeMeta:        metav1.TypeMeta{},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating virtual service: %s", vs.Name)
	}
	return nil
}

func (iom *IstioManager) CreateDestinationRule(ctx context.Context, dr *istio_networking.DestinationRule) error {
	_, err := iom.destinationRulesClient.Create(ctx, dr, metav1.CreateOptions{
		TypeMeta:        metav1.TypeMeta{},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating virtual service: %s", dr.Name)
	}
	return nil
}

// how to expose API to configure ordering of routing rule? https://istio.io/latest/docs/concepts/traffic-management/#routing-rule-precedence
func (iom *IstioManager) AddRoutingRule(ctx context.Context, vsName string, routingRule *istio.HTTPRoute) error {
	vs, err := iom.virtualServicesClient.Get(ctx, vsName, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving virtual service '%s'", vsName)
	}
	// always prepend routing rules due to routing rule precedence
	vs.Spec.Http = append([]*istio.HTTPRoute{routingRule}, vs.Spec.Http...)
	_, err = iom.virtualServicesClient.Update(ctx, vs, metav1.UpdateOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating virtual service '%s' with routing rule: %v", vsName, routingRule)
	}
	return nil
}

func (iom *IstioManager) AddSubset(ctx context.Context, drName string, subset *istio.Subset) error {
	dr, err := iom.destinationRulesClient.Get(ctx, drName, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving destination rule '%s'", drName)
	}
	// if there already exists a subset for the same , just update it
	shouldAddNewSubset := true
	for _, s := range dr.Spec.Subsets {
		if s.Name == subset.Name {
			s = subset
			shouldAddNewSubset = false
		}
	}
	if shouldAddNewSubset {
		dr.Spec.Subsets = append(dr.Spec.Subsets, subset)
	}
	_, err = iom.destinationRulesClient.Update(ctx, dr, metav1.UpdateOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating destination rule '%s' with subset: %v", drName, subset)
	}
	return nil
}
