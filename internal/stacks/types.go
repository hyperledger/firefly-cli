// Copyright © 2021 Kaleido, Inc.
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

package stacks

import (
	"fmt"
	"strings"

	"github.com/hyperledger/firefly-cli/pkg/types"
)

type DatabaseSelection int

const (
	PostgreSQL DatabaseSelection = iota
	SQLite3
)

var DBSelectionStrings = []string{"postgres", "sqlite3"}

func (db DatabaseSelection) String() string {
	return DBSelectionStrings[db]
}

func DatabaseSelectionFromString(s string) (DatabaseSelection, error) {
	for i, dbSelection := range DBSelectionStrings {
		if strings.ToLower(s) == dbSelection {
			return DatabaseSelection(i), nil
		}
	}
	return SQLite3, fmt.Errorf("\"%s\" is not a valid database selection. valid options are: %v", s, DBSelectionStrings)
}

type BlockchainProvider int

const (
	GoEthereum BlockchainProvider = iota
	HyperledgerBesu
	HyperledgerFabric
	Corda
)

var BlockchainProviderStrings = []string{"geth", "besu", "fabric", "corda"}

func (blockchainProvider BlockchainProvider) String() string {
	return BlockchainProviderStrings[blockchainProvider]
}

func BlockchainProviderFromString(s string) (BlockchainProvider, error) {
	for i, blockchainProviderSelection := range BlockchainProviderStrings {
		if strings.ToLower(s) == blockchainProviderSelection {
			return BlockchainProvider(i), nil
		}
	}
	return GoEthereum, fmt.Errorf("\"%s\" is not a valid blockchain provider selection. valid options are: %v", s, BlockchainProviderStrings)
}

const (
	NilTokens    types.TokenProvider = "none"
	ERC1155      types.TokenProvider = "erc1155"
	ERC20_ERC721 types.TokenProvider = "erc20_erc721"
)

var ValidTokenProviders = []types.TokenProvider{NilTokens, ERC1155, ERC20_ERC721}

func TokenProvidersFromStrings(strTokens []string) (tps types.TokenProviders, err error) {
	tps = make([]types.TokenProvider, 0, len(strTokens))
	for _, s := range strTokens {
		found := false
		for _, tokensProviderSelection := range ValidTokenProviders {
			if strings.ToLower(s) == string(tokensProviderSelection) {
				found = true
				if tokensProviderSelection != NilTokens {
					tps = append(tps, tokensProviderSelection)
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("\"%s\" is not a valid tokens provider selection. valid options are: %v", s, ValidTokenProviders)
		}
	}
	return tps, nil
}
