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
	// Create a temporary contract JSON file for testing (replace with your actual test file)
	FilePath := "testdata"
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
	if !CompareStringSlices(contracts, expectedContracts) {
		t.Errorf("Expected contracts: %v, Got: %v", expectedContracts, contracts)
	}
}

func CompareStringSlices(a, b []string) bool {
	//compare string lengths
	if len(a) != len(b) {
		return false
	}
	//compare values
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
