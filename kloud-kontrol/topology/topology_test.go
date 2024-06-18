package topology

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	redisServiceName    = "redis-prod"
	redisServiceVersion = "6.0.8"
	redisServiceID      = "node-1"

	votingAppServiceName    = "voting-app-ui"
	votingAppServiceVersion = "latest"
	votingAppServiceID      = "node-2"
)

func TestComposeToTopology(t *testing.T) {
	testCompose := []types.ServiceConfig{}
	portStr := "6379"
	allowEmpty := "yes"
	redisContainerName := "redis-prod"
	testCompose = append(testCompose, types.ServiceConfig{
		Name:          "azure-vote-back",
		Image:         "bitnami/redis:6.0.8",
		ContainerName: redisContainerName,
		Environment: map[string]*string{
			"ALLOW_EMPTY_PASSWORD": &allowEmpty,
			"REDIS_PORT_NUMBER":    &portStr,
		},
		Ports: []types.ServicePortConfig{
			{
				Target:    6379,
				Published: "6379",
				Protocol:  "TCP",
				Mode:      "ingress",
			},
		},
	})

	testCompose = append(testCompose, types.ServiceConfig{
		Name:          "azure-vote-front",
		Image:         "voting-app-ui",
		ContainerName: "voting-app-ui",
		Environment: map[string]*string{
			"REDIS": &redisContainerName,
		},
		Ports: []types.ServicePortConfig{
			{
				Target:    80,
				Published: "80",
				Protocol:  "TCP",
				Mode:      "ingress",
			},
		},
	})

	topo := ComposeToTopology(&testCompose)
	require.NotNil(t, topo)
}
