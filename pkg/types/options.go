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
	FireFlyBasePort          int
	ServicesBasePort         int
	DatabaseSelection        DatabaseSelection
	Verbose                  bool
	ExternalProcesses        int
	OrgNames                 []string
	NodeNames                []string
	BlockchainConnector      BlockchainConnector
	BlockchainProvider       BlockchainProvider
	BlockchainNodeProvider   BlockchainNodeProvider
	TokenProviders           TokenProviders
	FireFlyVersion           string
	ManifestPath             string
	PrometheusEnabled        bool
	PrometheusPort           int
	SandboxEnabled           bool
	ExtraCoreConfigPath      string
	ExtraConnectorConfigPath string
	BlockPeriod              int
	ContractAddress          string
	RemoteNodeURL            string
	ChainID                  int64
	DisableTokenFactories    bool
	RequestTimeout           int
	ReleaseChannel           ReleaseChannelSelection
	MultipartyEnabled        bool
}

type BlockchainProvider int

const (
	Ethereum BlockchainProvider = iota
	HyperledgerFabric
	Corda
)

type BlockchainConnector int

const (
	Ethconnect BlockchainConnector = iota
	Evmconnect
	Fabconnect
)

var BlockchainConnectorStrings = []string{"ethconnect", "evmconnect", "fabconnect"}

func (blockchainConnector BlockchainConnector) String() string {
	return BlockchainConnectorStrings[blockchainConnector]
}

func BlockchainConnectorFromStrings(s string) (BlockchainConnector, error) {
	for i, blockchainConnectorSelection := range BlockchainConnectorStrings {
		if strings.ToLower(s) == blockchainConnectorSelection {
			return BlockchainConnector(i), nil
		}
	}
	return Ethconnect, fmt.Errorf("\"%s\" is not a valid blockchain connector selection. valid options are: %v", s, BlockchainConnectorStrings)
}

var BlockchainProviderStrings = []string{"ethereum", "fabric", "corda"}

func (blockchainProvider BlockchainProvider) String() string {
	return BlockchainProviderStrings[blockchainProvider]
}

type BlockchainNodeProvider int

const (
	GoEthereum BlockchainNodeProvider = iota
	HyperledgerBesu
	RemoteRPC
)

var BlockchainNodeProviderStrings = []string{"geth", "besu", "remote-rpc"}

func (blockchainNodeProvider BlockchainNodeProvider) String() string {
	if blockchainNodeProvider < 0 {
		return ""
	}
	return BlockchainNodeProviderStrings[blockchainNodeProvider]
}

func BlockchainFromStrings(blockchainString, nodeString string) (blockchain BlockchainProvider, node BlockchainNodeProvider, err error) {
	blockchain = -1
	for i, blockchainProviderSelection := range BlockchainProviderStrings {
		if strings.ToLower(blockchainString) == blockchainProviderSelection {
			blockchain = BlockchainProvider(i)
			break
		}
	}
	if blockchain < 0 {
		// Migration cases for how we did things previously:
		// - For when "-b geth" or "-b besu" used to be the thing to do
		// - For when we persisted integers into the stack.json
		switch blockchainString {
		case "0", "geth":
			blockchain = Ethereum
			nodeString = "geth"
		case "1", "besu":
			blockchain = Ethereum
			nodeString = "besu"
		case "2":
			blockchain = HyperledgerFabric
		case "3":
			blockchain = HyperledgerFabric
		default:
			return -1, -1, fmt.Errorf("\"%s\" is not a valid blockchain provider selection. valid options are: %v", blockchainString, BlockchainProviderStrings)
		}
	}
	if blockchain == Ethereum {
		if nodeString == "" {
			node = GoEthereum
		} else {
			for i, blockchainNodeProviderSelection := range BlockchainNodeProviderStrings {
				if strings.ToLower(nodeString) == blockchainNodeProviderSelection {
					node = BlockchainNodeProvider(i)
					break
				}
			}
		}
		if node == -1 {
			return -1, -1, fmt.Errorf("\"%s\" is not a valid blockchain node selection. valid options are: %v", nodeString, BlockchainNodeProviderStrings)
		}
	} else {
		node = -1 // not currently applicable
	}
	return blockchain, node, nil
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

type ReleaseChannelSelection int

const (
	Stable ReleaseChannelSelection = iota
	Alpha
	Beta
	RC
)

var ReleaseChannelSelectionStrings = []string{"stable", "rc", "beta", "alpha", "head"}

func (rc ReleaseChannelSelection) String() string {
	return ReleaseChannelSelectionStrings[rc]
}

func ReleaseChannelSelectionFromString(s string) (ReleaseChannelSelection, error) {
	for i, releaseChannelSelection := range ReleaseChannelSelectionStrings {
		if strings.ToLower(s) == releaseChannelSelection {
			return ReleaseChannelSelection(i), nil
		}
	}
	return Stable, fmt.Errorf("\"%s\" is not a valid release channel selection. valid options are: %v", s, ReleaseChannelSelectionStrings)
}
