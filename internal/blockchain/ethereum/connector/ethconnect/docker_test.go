package ethconnect

import (
	"testing"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

type MockManfest struct {
	types.ManifestEntry
	GetDockerImageStringMck func() string
}

func TestGetServiceDefinition(t *testing.T) {
	getManifest := &MockManfest{
		GetDockerImageStringMck: func() string {
			return "ethconnect_alpine:latest"
		},
	}
	testServices := []struct {
		Name              string
		Members           *types.Stack
		DependentServices map[string]string
		ServiceName       string
	}{
		{
			Name: "test_service_1",
			Members: &types.Stack{
				Members:         []*types.Organization{{ID: "firefly_1", ExposedConnectorPort: 3000}},
				VersionManifest: &types.VersionManifest{Ethconnect: &getManifest.ManifestEntry},
			},
			DependentServices: map[string]string{
				"service1": "running",
				"service2": "stopped",
			},
			ServiceName: "ethconnect_firefly_1",
		},
		{
			Name: "test_service_2",
			Members: &types.Stack{
				Members:         []*types.Organization{{ID: "firefly_2", ExposedConnectorPort: 8002}},
				VersionManifest: &types.VersionManifest{Ethconnect: &getManifest.ManifestEntry},
			},
			DependentServices: map[string]string{
				"service1": "stopped",
				"service2": "running",
			},
			ServiceName: "ethconnect_firefly_2",
		},
		{
			Name: "test_service_3",
			Members: &types.Stack{
				Members:         []*types.Organization{{ID: "firefly_3", ExposedConnectorPort: 8000}},
				VersionManifest: &types.VersionManifest{Ethconnect: &getManifest.ManifestEntry},
			},
			DependentServices: map[string]string{
				"service1": "stopped",
				"service2": "stopped",
				"service3": "running",
			},
			ServiceName: "ethconnect_firefly_3",
		},
		{
			Name: "test_service_4",
			Members: &types.Stack{
				Members:         []*types.Organization{{ID: "firefly_4", ExposedConnectorPort: 7892}},
				VersionManifest: &types.VersionManifest{Ethconnect: &getManifest.ManifestEntry},
			},
			DependentServices: map[string]string{
				"service1": "stopped",
				"service2": "stopped",
				"service3": "stopped",
				"service4": "running",
			},
			ServiceName: "ethconnect_firefly_4",
		},
		{
			Name: "test_env_vars_5",
			Members: &types.Stack{
				Members:         []*types.Organization{{ID: "firefly_5", ExposedConnectorPort: 7892}},
				VersionManifest: &types.VersionManifest{Ethconnect: &getManifest.ManifestEntry},
				EnvironmentVars: map[string]interface{}{"HTTP_PROXY": ""},
			},
			DependentServices: map[string]string{
				"service1": "running",
				"service2": "stopped",
			},
			ServiceName: "ethconnect_firefly_5",
		},
	}
	for _, tc := range testServices {
		t.Run(tc.Name, func(t *testing.T) {
			e := Ethconnect{}

			serviceDefinitions := e.GetServiceDefinitions(tc.Members, tc.DependentServices)
			assert.NotNil(t, serviceDefinitions)

			expectedCommand := "server -f ./config/config.yaml -d 2"
			if serviceDefinitions[0].Service.Command != expectedCommand {
				t.Errorf("Expected Command %q, got %q", expectedCommand, serviceDefinitions[0].Service.Command)
			}
			if serviceDefinitions[0].ServiceName != tc.ServiceName {
				t.Errorf("Expected ServiceName %q, got %q", tc.ServiceName, serviceDefinitions[0].ServiceName)
			}

		})
	}

}
