package quorum

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/docker/mocks"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewQuorumProvider(t *testing.T) {
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
				Name: "TestQuorumWithEvmConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org1",
						NodeName: "quorum",
						Account: &ethereum.Account{
							Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
						},
					},
					{
						OrgName:  "Org2",
						NodeName: "quorum",
						Account: &ethereum.Account{
							Address:    "0x1234567890abcdef012345670000000000000000",
							PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "evmconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
			},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name: "TestQuorumWithEthConnect",
				Members: []*types.Organization{
					{
						OrgName:  "Org55",
						NodeName: "quorum",
						Account: &ethereum.Account{
							Address:    "0x1f2a000000000000000000000000000000000000",
							PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
						},
					},
				},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "ethconnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			quorumProvider := NewQuorumProvider(tc.Ctx, tc.Stack)
			assert.NotNil(t, quorumProvider)
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
				"address":      "0x1234567890abcdef0123456789abcdef6789abcd",
				"privateKey":   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				"ptmPublicKey": "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
			},
			ExpectedAccount: &ethereum.Account{
				Address:      "0x1234567890abcdef0123456789abcdef6789abcd",
				PrivateKey:   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
			},
		},
		{
			Name: "Account 2",
			Address: map[string]interface{}{
				"address":      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
				"privateKey":   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
				"ptmPublicKey": "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
			},
			ExpectedAccount: &ethereum.Account{
				Address:      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
				PrivateKey:   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
				PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
			},
		},
		{
			Name: "Account 3",
			Address: map[string]interface{}{
				"address":    "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
				"privateKey": "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
			},
			ExpectedAccount: &ethereum.Account{
				Address:    "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
				PrivateKey: "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			quorumProvider := &QuorumProvider{}
			result := quorumProvider.ParseAccount(tc.Address)

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
				NodeName: "quorum",
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
				NodeName: "quorum",
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
				NodeName: "quorum",
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
			p := &QuorumProvider{}

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
	p := &QuorumProvider{}

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
				NodeName: "quorum",
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
				NodeName: "quorum",
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
		p := &QuorumProvider{}
		result := p.GetConnectorExternalURL(tc.Org)
		assert.Equal(t, tc.ExpectedPort, result)
	}
}

func TestCreateAccount(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name                      string
		Ctx                       context.Context
		Stack                     *types.Stack
		PrivateTransactionManager fftypes.FFEnum
		Args                      []string
	}{
		{
			Name: "testcase1",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "ethconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
			},
			Args: []string{"Org-1_quorum", "Org-1_quorum", "0"},
		},
		{
			Name: "testcase2",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-2_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "ethconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
			},
			Args: []string{"Org-2_quorum", "Org-2_quorum", "1"},
		},
		{
			Name: "testcase3",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-3_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "Ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "EvmConnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerTessera,
			},
			Args: []string{"Org-3_quorum", "Org-3_quorum", "1"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := NewQuorumProvider(tc.Ctx, tc.Stack)
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
			// Check if the tessera public key is a non-empty string
			if tc.Stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
				assert.NotEmpty(t, account.PtmPublicKey)
			} else {
				assert.Empty(t, account.PtmPublicKey)
			}
		})
	}
}

func TestPostStart(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name                      string
		Ctx                       context.Context
		Stack                     *types.Stack
		PrivateTransactionManager fftypes.FFEnum
		Args                      []string
	}{
		{
			Name: "testcase1_with_tessera_enabled",
			Ctx:  ctx,
			Stack: &types.Stack{
				State: &types.StackState{
					DeployedContracts: make([]*types.DeployedContract, 0),
				},
				ExposedBlockchainPort:     8545,
				PrivateTransactionManager: types.PrivateTransactionManagerTessera,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
						Account: &ethereum.Account{
							Address:      "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey:   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
					{
						Index: &[]int{1}[0],
						Account: &ethereum.Account{
							Address:      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
							PrivateKey:   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
				},
			},
		},
		{
			Name: "testcase2_with_tessera_disabled",
			Ctx:  ctx,
			Stack: &types.Stack{
				State: &types.StackState{
					DeployedContracts: make([]*types.DeployedContract, 0),
				},
				ExposedBlockchainPort:     8545,
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
						Account: &ethereum.Account{
							Address:      "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey:   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
					{
						Index: &[]int{1}[0],
						Account: &ethereum.Account{
							Address:      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
							PrivateKey:   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := make([]interface{}, len(tc.Stack.Members))
			for memberIndex, member := range tc.Stack.Members {
				accounts[memberIndex] = member.Account

			}
			tc.Stack.State.Accounts = accounts
			p := NewQuorumProvider(tc.Ctx, tc.Stack)
			utils.StartMockServer(t)
			// mock quorum rpc response during the unlocking of accounts
			for _, member := range tc.Stack.Members {
				rpcUrl := fmt.Sprintf("http://127.0.0.1:%v", p.stack.ExposedBlockchainPort+(*member.Index*ExposedBlockchainPortMultiplier))
				httpmock.RegisterResponder(
					"POST",
					rpcUrl,
					httpmock.NewStringResponder(200, "{\"JSONRPC\": \"2.0\"}"))
			}
			httpmock.Activate()
			hasRunBefore, _ := p.stack.HasRunBefore()
			err := p.PostStart(hasRunBefore)
			assert.Nil(t, err, "post start should not have an error")
			utils.StopMockServer(t)
		})
	}
}

func TestFirstTimeSetup(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name                      string
		Ctx                       context.Context
		Stack                     *types.Stack
		PrivateTransactionManager fftypes.FFEnum
		Args                      []string
	}{
		{
			Name: "testcase1_no_members",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
			},
		},
		{
			Name: "testcase2_with_members",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
					},
					{
						Index: &[]int{1}[0],
					},
				},
			},
		},
		{
			Name: "testcase3_with_members_and_tessera_enabled",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerTessera,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := NewQuorumProvider(tc.Ctx, tc.Stack)
			p.dockerMgr = mocks.NewDockerManager() // docker related functionality should be tested in docker package
			err := p.FirstTimeSetup()
			assert.Nil(t, err, "first time setup should not throw an error")
		})
	}
}

func TestWriteConfig(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name                      string
		Ctx                       context.Context
		Stack                     *types.Stack
		PrivateTransactionManager fftypes.FFEnum
		Options                   *types.InitOptions
	}{
		{
			Name: "testcase1_no_members",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
			},
			Options: &types.InitOptions{
				BlockPeriod: 5,
			},
		},
		{
			Name: "testcase2_with_members",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerNone,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
						Account: &ethereum.Account{
							Address:      "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey:   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
					{
						Index: &[]int{1}[0],
						Account: &ethereum.Account{
							Address:      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
							PrivateKey:   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
				},
			},
			Options: &types.InitOptions{
				BlockPeriod:              5,
				ExtraConnectorConfigPath: "",
			},
		},
		{
			Name: "testcase3_with_members_and_tessera_enabled",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                      "Org-1_quorum",
				BlockchainProvider:        fftypes.FFEnumValue("BlockchainProvider", "ethereum"),
				BlockchainConnector:       fftypes.FFEnumValue("BlockChainConnector", "evmconnect"),
				BlockchainNodeProvider:    fftypes.FFEnumValue("BlockchainNodeProvider", "quorum"),
				InitDir:                   t.TempDir(),
				RuntimeDir:                t.TempDir(),
				PrivateTransactionManager: types.PrivateTransactionManagerTessera,
				Members: []*types.Organization{
					{
						Index: &[]int{0}[0],
						Account: &ethereum.Account{
							Address:      "0x1234567890abcdef0123456789abcdef6789abcd",
							PrivateKey:   "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
					{
						Index: &[]int{1}[0],
						Account: &ethereum.Account{
							Address:      "0x618E98197aF52F44D1B05Af0952a59b9f702dea4",
							PrivateKey:   "1b2b1ac0127957bb57e914993c47bfd69c5b0acc86425ee8ab2108f684a68a15",
							PtmPublicKey: "SBEV8qc12zSe7XfhqSChloYryb5aDK0XdBF3IwxZADE=",
						},
					},
				},
			},
			Options: &types.InitOptions{
				BlockPeriod:              5,
				ExtraConnectorConfigPath: "",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := NewQuorumProvider(tc.Ctx, tc.Stack)
			err := p.WriteConfig(tc.Options)
			assert.Nil(t, err, "writing config should not throw an error")
		})
	}
}
