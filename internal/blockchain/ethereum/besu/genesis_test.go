package besu

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateGenesis(t *testing.T) {
	//test different parameter cases for the CreateGenesis()
	testCases := []struct {
		Name        string
		addresses   []string
		blockPeriod int
		chainID     int64
	}{
		{
			Name:        "testcase1",
			addresses:   []string{"0xAddress1", "0xAddress2"},
			blockPeriod: 10,
			chainID:     int64(123),
		},
		{
			Name:        "testcase2",
			addresses:   []string{"0xAddress3", "0xAddress4"},
			blockPeriod: 7,
			chainID:     int64(456),
		},
		{
			Name:        "testcase3",
			addresses:   []string{"0xAddress14", "0xAddress29"},
			blockPeriod: 22,
			chainID:     345,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			genesis := CreateGenesis(tc.addresses, tc.blockPeriod, tc.chainID)
			extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
			alloc := make(map[string]*Alloc)
			for _, address := range tc.addresses {
				alloc[address] = &Alloc{
					Balance: "0x200000000000000000000000000000000000000000000000000000000000000",
				}
				extraData += address
			}
			extraData = strings.ReplaceAll(fmt.Sprintf("%-236s", extraData), " ", "0")
			expectedGenesis := &Genesis{
				Config: &GenesisConfig{
					ChainID:                tc.chainID,
					ConstantinopleFixBlock: 0,
					Clique: &CliqueConfig{
						BlockPeriodSeconds: tc.blockPeriod,
						EpochLength:        30000,
					},
				},
				Coinbase:   "0x0000000000000000000000000000000000000000",
				Difficulty: "0x1",
				ExtraData:  extraData,
				GasLimit:   "0xffffffff",
				MixHash:    "0x0000000000000000000000000000000000000000000000000000000000000000",
				Nonce:      "0x0",
				Timestamp:  "0x5c51a607",
				Alloc:      alloc,
				Number:     "0x0",
				GasUsed:    "0x0",
				ParentHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
			}
			// Assert that the generated Genesis is equal to the expected Genesis
			assert.Equal(t, expectedGenesis, genesis, "Generated Genesis does not match the expected Genesis")
		})
	}

}

func TestWriteGenesisJSON(t *testing.T) {
	filepath := t.TempDir()

	testCases := []struct {
		Name          string
		SampleGenesis Genesis
		filename      string
	}{
		{
			Name: "TestCase1",
			SampleGenesis: Genesis{
				Config: &GenesisConfig{
					ChainID:                int64(456),
					ConstantinopleFixBlock: 0,
					Clique: &CliqueConfig{
						BlockPeriodSeconds: 20,
						EpochLength:        2000,
					},
				},
			},
			filename: filepath + "/genesis1_output.json",
		},
		{
			Name: "TestCase2",
			SampleGenesis: Genesis{
				Config: &GenesisConfig{
					ChainID:                int64(338),
					ConstantinopleFixBlock: 0,
					Clique: &CliqueConfig{
						BlockPeriodSeconds: 40,
						EpochLength:        4000,
					},
				},
			},
			filename: filepath + "/genesis2_output.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.SampleGenesis.WriteGenesisJSON(tc.filename)
			if err != nil {
				t.Log("unable to write Genesis JSON", err)
			}
			// Assert that there is no error
			assert.NoError(t, err)

			writtenJSONBytes, err := os.ReadFile(tc.filename)
			if err != nil {
				t.Log("Unable to write JSON Bytes", err)
			}
			assert.NoError(t, err)
			var writtenGenesis Genesis

			err = json.Unmarshal(writtenJSONBytes, &writtenGenesis)
			if err != nil {
				t.Log("unable to unmarshal JSON", err)
			}
			assert.NoError(t, err)

			// Assert that the written Genesis matches the original Genesis
			assert.Equal(t, tc.SampleGenesis, writtenGenesis)
		})

	}

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
				NodeName: "besu",
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
				NodeName: "besu",
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
		p := &BesuProvider{}
		result := p.GetConnectorExternalURL(tc.Org)
		assert.Equal(t, tc.ExpectedPort, result)
	}
}
