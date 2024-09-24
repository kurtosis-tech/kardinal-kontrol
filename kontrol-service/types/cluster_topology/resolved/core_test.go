package resolved

import (
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

const dummyPluginName = "https://github.com/h4ck3rk3y/identity-plugin.git"

var httpProtocol = "HTTP"
var dummySpec = &appsv1.DeploymentSpec{}

func TestHashFunc(t *testing.T) {
	feSer1 := createService()
	feSer2 := createService()

	assert.Equal(t, feSer1.Hash(), feSer2.Hash())
}

func createService() *Service {

	return &Service{
		ServiceID: "frontend",
		ServiceSpec: &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:        "http",
					Port:        80,
					AppProtocol: &httpProtocol,
				},
			},
			Selector: map[string]string{
				"app": "frontend",
			},
		},
		DeploymentSpec: dummySpec,
		IsExternal:     false,
		IsStateful:     false,
		StatefulPlugins: []*StatefulPlugin{
			{
				Name:        dummyPluginName,
				ServiceName: "free-currency-api",
				Type:        "external",
				Args:        map[string]string{},
			},
		},
	}
}