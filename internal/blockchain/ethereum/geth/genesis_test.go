package geth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGenesis(t *testing.T) {
	testCases := []struct {
		Name        string
		addresses   []string
		blockPeriod int
		chainID     int64
	}{
		{
			Name:        "testcase1",
			addresses:   []string{"0xAddress20", "0xAddress27"},
			blockPeriod: 28,
			chainID:     int64(21),
		},
		{
			Name:        "testcase2",
			addresses:   []string{"0xAddress36", "0xAddress45"},
			blockPeriod: 26,
			chainID:     int64(98),
		},
		{
			Name:        "testcase3",
			addresses:   []string{"0xAddress19", "0xAddress38", "0xAddress64", "0xAddress74"},
			blockPeriod: 40,
			chainID:     int64(93),
		},
		{
			Name:        "testcase4",
			addresses:   []string{"0xAddress96", "0xAddress25", "0xAddress49", "0xAddress24", "0xAddress37", "0xAddress12"},
			blockPeriod: 12,
			chainID:     int64(5000),
		},
		{
			Name:        "testcase5",
			addresses:   []string{"0xAddress62536", "0xAddress3261", "0xAddress82721"},
			blockPeriod: 14,
			chainID:     int64(900000),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			genesis := CreateGenesis(tc.addresses, tc.blockPeriod, tc.chainID)
			extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
			alloc := make(map[string]*Alloc)
			for _, address := range tc.addresses {
				alloc[address] = &Alloc{
					"0x200000000000000000000000000000000000000000000000000000000000000",
				}
				extraData += address
			}
			extraData = strings.ReplaceAll(fmt.Sprintf("%-236s", extraData), " ", "0")
			expectedGenesis := &Genesis{
				Config: &GenesisConfig{
					ChainID:             tc.chainID,
					HomesteadBlock:      0,
					Eip150Hash:          "0x0000000000000000000000000000000000000000000000000000000000000000",
					Eip155Block:         0,
					Eip158Block:         0,
					ByzantiumBlock:      0,
					ConstantinopleBlock: 0,
					Clique: &CliqueConfig{
						Period: tc.blockPeriod,
						Epoch:  30000,
					},
				},
				Nonce:      "0x0",
				Timestamp:  "0x60edb1c7",
				ExtraData:  extraData,
				GasLimit:   "0xffffff",
				Difficulty: "0x1",
				MixHash:    "0x0000000000000000000000000000000000000000000000000000000000000000",
				Coinbase:   "0x0000000000000000000000000000000000000000",
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
					ChainID:             int64(456),
					Eip155Block:         0,
					Eip158Block:         0,
					ByzantiumBlock:      0,
					ConstantinopleBlock: 0,
					IstanbulBlock:       0,
					Clique: &CliqueConfig{
						Period: 20,
						Epoch:  2000,
					},
				},
			},
			filename: filepath + "/genesis1_output.json",
		},
		{
			Name: "TestCase2",
			SampleGenesis: Genesis{
				Config: &GenesisConfig{
					ChainID:             int64(338),
					ConstantinopleBlock: 0,
					Eip155Block:         0,
					Eip158Block:         0,
					ByzantiumBlock:      0,
					IstanbulBlock:       0,
					Clique: &CliqueConfig{
						Period: 40,
						Epoch:  4000,
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
