package geth

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"

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
				NodeName: "geth",
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
				NodeName: "geth",
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
				NodeName: "geth",
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
			p := &GethProvider{}

			Orgconfig := p.GetOrgConfig(tc.Stack, tc.Org)
			assert.NotNil(t, Orgconfig)
			assert.Equal(t, tc.OrgConfig, Orgconfig)
		})
	}
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
	p := &GethProvider{}

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
func TestGetConnectorExternal(t *testing.T) {
	testcase := []struct {
		Name         string
		Org          *types.Organization
		ExpectedPort string
	}{
		{
			Name: "testcase1",
			Org: &types.Organization{
				OrgName:  "Org-1",
				NodeName: "geth",
				Account: &ethereum.Account{
					Address:    "0x1f2a000000000000000000000000000000000000",
					PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
				},
				ExposedConnectorPort: 8584,
			},
			ExpectedPort: "http://127.0.0.1:8584",
		},
		{
			Name: "testcase2",
			Org: &types.Organization{
				OrgName:  "Org-2",
				NodeName: "geth",
				Account: &ethereum.Account{
					Address:    "0xabcdeffedcba9876543210abcdeffedc00000000",
					PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
				},
				ExposedConnectorPort: 8000,
			},
			ExpectedPort: "http://127.0.0.1:8000",
		},
	}
	for _, tc := range testcase {
		p := &GethProvider{}
		result := p.GetConnectorExternalURL(tc.Org)
		assert.Equal(t, tc.ExpectedPort, result)
	}
}

func TestCreateAccount(t *testing.T) {
	testcases := []struct {
		Name  string
		Stack *types.Stack
		Args  []string
	}{
		{
			Name: "testcase1",
			Stack: &types.Stack{
				Name:                   "Org-1_geth",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockChainConnector", "Ethconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "geth"),
				InitDir:                t.TempDir(),
				RuntimeDir:             t.TempDir(),
			},
			Args: []string{},
		},
		{
			Name: "testcase1",
			Stack: &types.Stack{
				Name:                   "Org-2_geth",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockChainConnector", "Ethconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "geth"),
				InitDir:                t.TempDir(),
				RuntimeDir:             t.TempDir(),
			},
			Args: []string{},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &GethProvider{
				stack: tc.Stack,
			}
			Account, err := p.CreateAccount(tc.Args)
			if err != nil {
				t.Log("unable to create account", err)
			}
			//validate properties of account
			assert.NotNil(t, Account)
			account, ok := Account.(*ethereum.Account)
			assert.True(t, ok, "unexpected Type for account")

			//check if Ethereum Addresss is valid
			assert.NotEmpty(t, account.Address)
			// Check if the private key is a non-empty hex string
			assert.NotEmpty(t, account.PrivateKey)
			_, err = hex.DecodeString(account.PrivateKey)
			assert.NoError(t, err, "invalid private key format")
		})
	}
}
