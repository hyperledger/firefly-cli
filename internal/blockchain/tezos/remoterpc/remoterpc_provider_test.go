package remoterpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/tezos"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestParseAccount(t *testing.T) {
	tests := []struct {
		Name            string
		Account         map[string]interface{}
		ExpectedAccount *tezos.Account
	}{
		{
			Name: "Account 1",
			Account: map[string]interface{}{
				"address":    "0x1234567890abcdef0123456789abcdef6789abcd",
				"privateKey": "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			},
			ExpectedAccount: &tezos.Account{
				Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
				PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			},
		},
		{
			Name: "Account 2",
			Account: map[string]interface{}{
				"address":    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
				"privateKey": "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			},
			ExpectedAccount: &tezos.Account{
				Address:    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
				PrivateKey: "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			p := &RemoteRPCProvider{}

			account := p.ParseAccount(tc.Account)
			_, ok := account.(*tezos.Account)
			if !ok {
				t.Errorf("Expected result to be of type *ethereum.Account, but got %T", account)
			}
			assert.Equal(t, tc.ExpectedAccount, account, "Generated account unmatched")
		})
	}
}

func TestGetOrgConfig(t *testing.T) {
	testCases := []struct {
		Name           string
		Stack          *types.Stack
		Member         *types.Organization
		ExpectedConfig *types.OrgConfig
	}{
		{
			Name:  "TestOrg-1",
			Stack: &types.Stack{Name: "Hyperledger_1"},
			Member: &types.Organization{
				OrgName: "Hyperledger_1",
				Account: &tezos.Account{
					Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
					PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				},
			},
			ExpectedConfig: &types.OrgConfig{
				Name: "Hyperledger_1",
				Key:  "0x1234567890abcdef0123456789abcdef6789abcd",
			},
		},
		{
			Name:  "TestOrg-2",
			Stack: &types.Stack{Name: "Hyperledger_2"},
			Member: &types.Organization{
				OrgName: "Hyperledger_2",
				Account: &tezos.Account{
					Address:    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
					PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
				},
			},
			ExpectedConfig: &types.OrgConfig{
				Name: "Hyperledger_2",
				Key:  "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
			},
		},
		{
			Name:  "TetsOrg-3",
			Stack: &types.Stack{Name: "Hyperledger_3"},
			Member: &types.Organization{
				OrgName: "Hyperledger_3",
				Account: &tezos.Account{
					Address:    "0xabcdeffedcba9876543210abcdeffedc00000000",
					PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
				},
			},
			ExpectedConfig: &types.OrgConfig{
				Name: "Hyperledger_3",
				Key:  "0xabcdeffedcba9876543210abcdeffedc00000000",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := RemoteRPCProvider{}
			Orgconfig := p.GetOrgConfig(tc.Stack, tc.Member)
			assert.NotNil(t, Orgconfig)
			assert.NotNil(t, Orgconfig.Name)
			assert.NotNil(t, Orgconfig.Key)
			assert.Equal(t, tc.ExpectedConfig, Orgconfig)
		})
	}
}

func TestGetBlockChainConfig(t *testing.T) {
	tests := []struct {
		Name  string
		Stack *types.Stack
		M     *types.Organization
	}{
		{
			Name:  "TestCase_1",
			Stack: &types.Stack{Name: "User_1"},
			M:     &types.Organization{ID: "256", External: true},
		},
		{
			Name:  "TestCase_2",
			Stack: &types.Stack{Name: "User_2"},
			M:     &types.Organization{ID: "353", External: true},
		},
		{
			Name:  "TestCase_3",
			Stack: &types.Stack{Name: "User_3"},
			M:     &types.Organization{ID: "278", External: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			p := &RemoteRPCProvider{
				stack: tc.Stack,
			}
			blockchainConfig := p.GetBlockchainPluginConfig(tc.Stack, tc.M)
			assert.NotNil(t, blockchainConfig)
		})
	}
}

func TestNewRemoteRPCProvider(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		Name  string
		Ctx   context.Context
		Stack *types.Stack
	}{
		{
			Name: "testcase-1",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestRPCwithEvmConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org17",
						NodeName: "rpc",
						Account: &tezos.Account{
							Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Tezos"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "tezosconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "rpc"),
			},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestRPCWithTezosConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org34",
						NodeName: "rpc",
						Account: &tezos.Account{
							Address:    "0x1f2a000000000000000000000000000000000000",
							PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Tezos"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "tezosconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "rpc"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rpcProvider := NewRemoteRPCProvider(tc.Ctx, tc.Stack)
			assert.NotNil(t, rpcProvider)
		})
	}
}

func TestGetConnectorExternalURL(t *testing.T) {

	testURLS := []struct {
		Name        string
		OrgPort     *types.Organization
		ExpectedURL string
	}{
		{
			Name: "TestCase-1", OrgPort: &types.Organization{ExposedConnectorPort: 8000}, ExpectedURL: fmt.Sprintf("http://127.0.0.1:%v", 8000),
		},
		{
			Name: "TestCase-2", OrgPort: &types.Organization{ExposedConnectorPort: 8001}, ExpectedURL: fmt.Sprintf("http://127.0.0.1:%v", 8001),
		},
		{
			Name: "TestCase-3", OrgPort: &types.Organization{ExposedConnectorPort: 8003}, ExpectedURL: fmt.Sprintf("http://127.0.0.1:%v", 8003),
		},
		{
			Name: "TestCase-4", OrgPort: &types.Organization{ExposedConnectorPort: 7008}, ExpectedURL: fmt.Sprintf("http://127.0.0.1:%v", 7008),
		},
		{
			Name: "TestCase-3", OrgPort: &types.Organization{ExposedConnectorPort: 8080}, ExpectedURL: fmt.Sprintf("http://127.0.0.1:%v", 8080),
		},
	}
	for _, tc := range testURLS {
		t.Run(tc.Name, func(t *testing.T) {
			p := &RemoteRPCProvider{}
			ExternalURL := p.GetConnectorExternalURL(tc.OrgPort)
			assert.NotNil(t, ExternalURL)
			assert.Equal(t, tc.ExpectedURL, ExternalURL)
		})
	}
}
