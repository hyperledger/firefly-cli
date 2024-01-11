package fabric

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestNewFabricProvider(t *testing.T) {
	Stack := &types.Stack{
		Name:                   "FabricUser",
		Members:                []*types.Organization{{OrgName: "Hyperledger-fabric"}},
		BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "fabric"),
		BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "fabonnect"),
		BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "fabric"),
	}
	ctx := log.WithLogger(context.Background(), &log.StdoutLogger{})
	fabricProvider := NewFabricProvider(ctx, Stack)
	assert.NotNil(t, fabricProvider)
	assert.NotNil(t, fabricProvider.ctx)
	assert.NotNil(t, fabricProvider.stack)
	assert.NotNil(t, fabricProvider.log)
}

func TestGetOrgConfig(t *testing.T) {
	testcases := []struct {
		Name      string
		Stack     *types.Stack
		OrgConfig *types.OrgConfig
		Member    *types.Organization
	}{
		{
			Name:  "TestFabric-1",
			Stack: &types.Stack{Name: "fabric_user-1"},
			Member: &types.Organization{
				OrgName:  "Org-1",
				NodeName: "fabric",
				Account: &ethereum.Account{
					Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
					PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-1",
				Key:  "Org-1",
			},
		},
		{
			Name:  "TestFabric-2",
			Stack: &types.Stack{Name: "fabri_user-2"},
			Member: &types.Organization{
				OrgName:  "Org-2",
				NodeName: "besu",
				Account: &ethereum.Account{
					Address:    "0x1f2a000000000000000000000000000000000000",
					PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-2",
				Key:  "Org-2",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &FabricProvider{}
			orgConfig := p.GetOrgConfig(tc.Stack, tc.Member)
			assert.NotNil(t, orgConfig)
			assert.Equal(t, tc.OrgConfig, orgConfig)
		})

	}
}

func TestGetFabconnetServiceDefinitions(t *testing.T) {
	stack := &types.Stack{
		Name:       "fabric_user-1",
		RuntimeDir: "mock_runtime_dir",
		VersionManifest: &types.VersionManifest{
			Fabconnect: &types.ManifestEntry{
				Image: "fabric-apline",
				Local: true,
			},
		},
		RemoteFabricNetwork: true,
	}
	p := &FabricProvider{
		stack: stack,
	}
	// Create mock organizations
	members := []*types.Organization{
		{
			ID:                   "org1",
			ExposedConnectorPort: 4000,
		},
		{
			ID:                   "org2",
			ExposedConnectorPort: 4001,
		},
		{
			ID:                   "org3",
			ExposedConnectorPort: 4002,
		},
		{
			ID:                   "org4",
			ExposedConnectorPort: 4003,
		},
		{
			ID:                   "org5",
			ExposedConnectorPort: 4004,
		},
	}
	serviceDefinitions := p.getFabconnectServiceDefinitions(members)
	assert.NotNil(t, serviceDefinitions)
}

func TestParseAccount(t *testing.T) {
	input := map[string]interface{}{
		"name":    "user-1",
		"orgName": "hyperledger",
	}

	besuProvider := &FabricProvider{}
	result := besuProvider.ParseAccount(input)

	if _, ok := result.(*Account); !ok {
		t.Errorf("Expected result to be of type *ethereum.Account, but got %T", result)
	}
	expectedAccount := &Account{
		Name:    "user-1",
		OrgName: "hyperledger",
	}
	assert.Equal(t, expectedAccount, result, "Generated result unmatched")

}

func TestGetConnectorName(t *testing.T) {
	testString := "fabconnect"
	p := &FabricProvider{}
	connector := p.GetConnectorName()
	assert.NotNil(t, connector)
	assert.Equal(t, testString, connector)
}

func TestGetConnectorURL(t *testing.T) {
	testCases := []struct {
		Name        string
		Org         *types.Organization
		ExpectedURL string
	}{
		{
			Name: "testcase-1",
			Org: &types.Organization{
				ID:       "user-1",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Nicko",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-1:3000",
		},
		{
			Name: "testcase-2",
			Org: &types.Organization{
				ID:       "user-2",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Richardson",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-2:3000",
		},
		{
			Name: "testcase-3",
			Org: &types.Organization{
				ID:       "user-3",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Philip",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-3:3000",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &FabricProvider{}
			URL := p.GetConnectorURL(tc.Org)
			assert.Equal(t, tc.ExpectedURL, URL)
		})
	}

}

func TestGetConnectorExternalURL(t*testing.T){
	testCases := []struct {
		Name        string
		Org         *types.Organization
		ExpectedURL string
	}{
		{
			
		},
	}
}

func TestGetDockerPlatform(t *testing.T) {
	expectedString := "linux/amd64"
	String := getDockerPlatform()
	assert.Equal(t, expectedString, String)
}

// func TestCreateChannel(t *testing.T) {
// 	ctx := log.WithLogger(context.Background(), &log.StdoutLogger{
// 		LogLevel: log.Info,
// 	})

// 	Stack := &types.Stack{
// 		Name:     "fabric_user-1",
// 		StackDir: "mock_stack_dir",
// 	}

// 	p := &FabricProvider{
// 		stack: Stack,
// 		ctx:   ctx,
// 		log:   Log,
// 	}
// 	err := p.createChannel()
// 	if err != nil {
// 		t.Log("unable to create channel:", err)
// 	}
// }
