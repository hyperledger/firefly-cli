package besu

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestNewBesuProvider(t *testing.T) {
	var ctx context.Context
	testCases := []struct {
		Name  string
		Ctx   context.Context
		Stack *types.Stack
	}{
		{
			Name: "testcase1",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                   "TestBesuProviderWithEthconnect",
				Members:                []*types.Organization{{OrgName: "Org1"}},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "EthConnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "besu"),
			},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Members:                []*types.Organization{{OrgName: "Org2"}, {OrgName: "org4"}},
				Name:                   "TestBesuProviderWithEvmconnect",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "EvmConnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "besu"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			besuProvider := NewBesuProvider(tc.Ctx, tc.Stack)
			assert.NotNil(t, besuProvider)
		})
	}
}

func TestParseAccount(t *testing.T) {
	input := map[string]interface{}{
		"address":    "0xAddress1",
		"privateKey": "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
	}

	besuProvider := &BesuProvider{}
	result := besuProvider.ParseAccount(input)

	// Assert that the result is of type *ethereum.Account
	if _, ok := result.(*ethereum.Account); !ok {
		t.Errorf("Expected result to be of type *ethereum.Account, but got %T", result)
	}
	expectedAccount := &ethereum.Account{
		Address:    "0xAddress1",
		PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
	}
	assert.Equal(t, expectedAccount, result, "Generated result unmatched")

}

func TestGetContracts(t *testing.T) {
	FilePath := t.TempDir()
	testContractFile := filepath.Join(FilePath, "/test_contracts.json")
	// Sample contract JSON content for testing
	const testContractJSON = `{
			"contracts": {
				"Contract1": {
					"name": "sample_1",
					"abi": "sample_abi_1",
					"bin": "sample_bin_1"
				},
				"Contract2": {
					"name": "sample_2",
					"abi": "sample_abi_2",
					"bin": "sample_bin_2"
				}
			}
		}`
	p := &BesuProvider{}

	err := os.WriteFile(testContractFile, []byte(testContractJSON), 0755)
	if err != nil {
		t.Log("unable to write file:", err)
	}
	contracts, err := p.GetContracts(testContractFile, nil)
	if err != nil {
		t.Log("unable to get contract", err)
	}
	// Assert that the returned contracts match the expected contract names
	expectedContracts := []string{"Contract1", "Contract2"}
	assert.ElementsMatch(t, contracts, expectedContracts)
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
				NodeName: "besu",
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
				NodeName: "besu",
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
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &BesuProvider{}

			Orgconfig := p.GetOrgConfig(tc.Stack, tc.Org)
			assert.NotNil(t, Orgconfig)
			assert.Equal(t, tc.OrgConfig, Orgconfig)
		})
	}
}
