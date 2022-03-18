// Copyright © 2022 Kaleido, Inc.
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
// See the License for the specific lan

package types

import (
	"fmt"
	"strings"
)

type PullOptions struct {
	Retries int
}

type StartOptions struct {
	NoRollback bool
}

type InitOptions struct {
	FireFlyBasePort           int
	ServicesBasePort          int
	DatabaseSelection         DatabaseSelection
	Verbose                   bool
	ExternalProcesses         int
	OrgNames                  []string
	NodeNames                 []string
	BlockchainProvider        BlockchainProvider
	TokenProviders            TokenProviders
	FireFlyVersion            string
	ManifestPath              string
	PrometheusEnabled         bool
	PrometheusPort            int
	ExtraCoreConfigPath       string
	ExtraEthconnectConfigPath string
	BlockPeriod               int
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

type TokenProvider string

type TokenProviders []TokenProvider

func (tps TokenProviders) Strings() []string {
	ret := make([]string, len(tps))
	for i, t := range tps {
		ret[i] = string(t)
	}
	return ret
}

const (
	NilTokens    TokenProvider = "none"
	ERC1155      TokenProvider = "erc1155"
	ERC20_ERC721 TokenProvider = "erc20_erc721"
)

var ValidTokenProviders = []TokenProvider{NilTokens, ERC1155, ERC20_ERC721}

func TokenProvidersFromStrings(strTokens []string) (tps TokenProviders, err error) {
	tps = make([]TokenProvider, 0, len(strTokens))
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
