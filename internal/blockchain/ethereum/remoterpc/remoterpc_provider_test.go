package remoterpc

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestNewRemoteRPCProvider(t *testing.T) {
	var ctx context.Context
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
						OrgName:  "Org1",
						NodeName: "rpc",
						Account: &ethereum.Account{
							Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "Evmconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "rpc"),
			},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestRPCWithEthConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org7",
						NodeName: "rpc",
						Account: &ethereum.Account{
							Address:    "0x1f2a000000000000000000000000000000000000",
							PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "Ethconnect"),
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

func TestParseAccount(t *testing.T) {
	tests := []struct {
		Name            string
		ExpectedAccount *ethereum.Account
		Account         map[string]interface{}
	}{
		{
			Name: "Account 1",
			Account: map[string]interface{}{
				"address":    "0x1234567890abcdef0123456789abcdef6789abcd",
				"privateKey": "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			},
			ExpectedAccount: &ethereum.Account{
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
			ExpectedAccount: &ethereum.Account{
				Address:    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
				PrivateKey: "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			p := &RemoteRPCProvider{}

			account := p.ParseAccount(tc.Account)
			_, ok := account.(*ethereum.Account)
			if !ok {
				t.Errorf("Expected result to be of type *ethereum.Account, but got %T", account)
			}
			assert.Equal(t, tc.ExpectedAccount, account, "Generated account unmatched")
		})
	}
}

func TestGetOrgConfig(t *testing.T) {
	testCases := []struct {
		Name      string
		Org       *types.Organization
		Stack     *types.Stack
		OrgConfig *types.OrgConfig
	}{
		{
			Name: "Testcase1",
			Stack: &types.Stack{
				Name: "Org-1",
			},
			Org: &types.Organization{
				OrgName:  "Org-1",
				NodeName: "rpc",
				Account: &ethereum.Account{
					Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
					PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-1",
				Key:  "0x1234567890abcdef0123456789abcdef6789abcd",
			},
		},
		{
			Name: "Testcase2",
			Stack: &types.Stack{
				Name: "Org-2",
			},
			Org: &types.Organization{
				OrgName:  "Org-2",
				NodeName: "rpc",
				Account: &ethereum.Account{
					Address:    "0x1f2a000000000000000000000000000000000000",
					PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-2",
				Key:  "0x1f2a000000000000000000000000000000000000",
			},
		},
		{
			Name: "Testcase3",
			Stack: &types.Stack{
				Name: "Org-3",
			},
			Org: &types.Organization{
				OrgName:  "Org-3",
				NodeName: "rpc",
				Account: &ethereum.Account{
					Address:    "0xabcdeffedcba9876543210abcdeffedc00000000",
					PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-3",
				Key:  "0xabcdeffedcba9876543210abcdeffedc00000000",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &RemoteRPCProvider{}

			Orgconfig := p.GetOrgConfig(tc.Stack, tc.Org)
			assert.NotNil(t, Orgconfig)
			assert.Equal(t, tc.OrgConfig, Orgconfig)
		})
	}
}
