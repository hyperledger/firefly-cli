package geth

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestNewGethProvider(t *testing.T) {
	var ctx context.Context

	testcases := []struct {
		Name  string
		Ctx   context.Context
		Stack *types.Stack
	}{
		{
			Name: "testcase1",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestGethWithEvmConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org1",
						NodeName: "geth",
						Account: &ethereum.Account{
							Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
						},
					},
					{
						OrgName:  "Org2",
						NodeName: "geth",
						Account: &ethereum.Account{
							Address:    "0x1234567890abcdef012345670000000000000000",
							PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "Evmconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "geth"),
			},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestGethWithEthConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org55",
						NodeName: "geth",
						Account: &ethereum.Account{
							Address:    "0x1f2a000000000000000000000000000000000000",
							PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "Ethconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "geth"),
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			gethProvider := NewGethProvider(tc.Ctx, tc.Stack)
			assert.NotNil(t, gethProvider)
		})
	}
}

func TestParseAccount(t *testing.T) {
	testcases := []struct {
		Name            string
		Address         map[string]interface{}
		ExpectedAccount *ethereum.Account
	}{
		{
			Name: "Account 1",
			Address: map[string]interface{}{
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
			Address: map[string]interface{}{
				"address":    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
				"privateKey": "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			},
			ExpectedAccount: &ethereum.Account{
				Address:    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
				PrivateKey: "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			gethProvider := &GethProvider{}
			result := gethProvider.ParseAccount(tc.Address)

			_, ok := result.(*ethereum.Account)
			if !ok {
				t.Errorf("Expected result to be of type *ethereum.Account, but got %T", result)
			}
			assert.Equal(t, tc.ExpectedAccount, result, "Generated result unmatched")
		})
	}
}
