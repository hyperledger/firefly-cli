// Copyright © 2024 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package besu

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Storage struct {
	Field1 string `json:"0x0000000000000000000000000000000000000000000000000000000000000000"`
	Field2 string `json:"0x0000000000000000000000000000000000000000000000000000000000000001"`
	Field3 string `json:"0x0000000000000000000000000000000000000000000000000000000000000004"`
}

type Genesis struct {
	Config     *GenesisConfig    `json:"config"`
	Nonce      string            `json:"nonce"`
	Timestamp  string            `json:"timestamp"`
	ExtraData  string            `json:"extraData"`
	GasLimit   string            `json:"gasLimit"`
	Difficulty string            `json:"difficulty"`
	MixHash    string            `json:"mixHash"`
	Coinbase   string            `json:"coinbase"`
	Alloc      map[string]*Alloc `json:"alloc"`
	Number     string            `json:"number"`
	GasUsed    string            `json:"gasUsed"`
	ParentHash string            `json:"parentHash"`
}

type GenesisConfig struct {
	ChainID                int64         `json:"chainId"`
	ConstantinopleFixBlock int           `json:"constantinoplefixblock"`
	Clique                 *CliqueConfig `json:"clique"`
}

type CliqueConfig struct {
	EpochLength        int `json:"epochlength"`
	BlockPeriodSeconds int `json:"blockperiodseconds"`
}

type Alloc struct {
	Balance string   `json:"balance"`
	Code    string   `json:"code,omitempty"`
	Storage *Storage `json:"storage,omitempty"`
}

func (g *Genesis) WriteGenesisJSON(filename string) error {
	genesisJSONBytes, _ := json.MarshalIndent(g, "", " ")
	if err := os.WriteFile(filename, genesisJSONBytes, 0755); err != nil {
		return err
	}
	return nil
}

func CreateGenesis(addresses []string, blockPeriod int, chainID int64) *Genesis {
	if blockPeriod == -1 {
		blockPeriod = 5
	}
	extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
	alloc := make(map[string]*Alloc)
	for _, address := range addresses {
		alloc[address] = &Alloc{
			Balance: "0x200000000000000000000000000000000000000000000000000000000000000",
		}
		extraData += address
	}
	extraData = strings.ReplaceAll(fmt.Sprintf("%-236s", extraData), " ", "0")
	return &Genesis{
		Config: &GenesisConfig{
			ChainID:                chainID,
			ConstantinopleFixBlock: 0,
			Clique: &CliqueConfig{
				BlockPeriodSeconds: blockPeriod,
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
}
