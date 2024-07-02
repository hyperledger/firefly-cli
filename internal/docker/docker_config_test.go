package docker

import (
	"testing"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

type MockManfest struct {
	types.ManifestEntry
}

func TestCreateDockerComposeEnvironmentVars(t *testing.T) {
	getManifest := &MockManfest{}
	testStacks := []struct {
		Id              int
		EnvironmentVars map[string]interface{}
		Members         []*types.Organization
		VersionManifest *types.VersionManifest
	}{
		{
			Id:              1,
			Members:         []*types.Organization{{ID: "firefly_1"}},
			VersionManifest: &types.VersionManifest{FireFly: &getManifest.ManifestEntry, DataExchange: &getManifest.ManifestEntry},
			EnvironmentVars: map[string]interface{}{
				"http_proxy":  "",
				"https_proxy": "",
				"HTTP_PROXY":  "127.0.0.1:8080",
				"HTTPS_PROXY": "127.0.0.1:8080",
				"no_proxy":    "",
			},
		},
		{
			Id:              2,
			Members:         []*types.Organization{{ID: "firefly_2"}},
			VersionManifest: &types.VersionManifest{FireFly: &getManifest.ManifestEntry, DataExchange: &getManifest.ManifestEntry},
			EnvironmentVars: nil,
		},
	}
	for _, test := range testStacks {
		cfg := CreateDockerCompose(&types.Stack{
			Members:         test.Members,
			VersionManifest: test.VersionManifest,
			EnvironmentVars: test.EnvironmentVars,
		})
		for _, service := range cfg.Services {
			assert.Equal(t, len(test.EnvironmentVars), len(service.Environment), "service [%v] test ID [%v]", service.ContainerName, test.Id)
			for envVar := range service.Environment {
				assert.Equal(t, test.EnvironmentVars[envVar], service.Environment[envVar])
			}
		}
	}
}
