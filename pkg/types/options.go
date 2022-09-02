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
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
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
	DatabaseProvider         string
	ExternalProcesses        int
	OrgNames                 []string
	NodeNames                []string
	BlockchainConnector      string
	BlockchainProvider       string
	BlockchainNodeProvider   string
	TokenProviders           []string
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
	ReleaseChannel           string
	MultipartyEnabled        bool
	IPFSMode                 string
}

const IPFSMode = "ipfs_mode"

var (
	IPFSModePrivate = fftypes.FFEnumValue(IPFSMode, "private")
	IPFSModePublic  = fftypes.FFEnumValue(IPFSMode, "public")
)

const BlockchainProvider = "blockchain_provider"

var (
	BlockchainProviderEthereum = fftypes.FFEnumValue(BlockchainProvider, "ethereum")
	BlockchainProviderFabric   = fftypes.FFEnumValue(BlockchainProvider, "fabric")
	BlockchainProviderCorda    = fftypes.FFEnumValue(BlockchainProvider, "corda")
)

const BlockchainConnector = "blockchain_connector"

var (
	BlockchainConnectorEthconnect = fftypes.FFEnumValue(BlockchainConnector, "ethconnect")
	BlockchainConnectorEvmconnect = fftypes.FFEnumValue(BlockchainConnector, "evmconnect")
	BlockchainConnectorFabconnect = fftypes.FFEnumValue(BlockchainConnector, "fabric")
)

const BlockchainNodeProvider = "blockchain_node_provider"

var (
	BlockchainNodeProviderGeth      = fftypes.FFEnumValue(BlockchainNodeProvider, "geth")
	BlockchainNodeProviderBesu      = fftypes.FFEnumValue(BlockchainNodeProvider, "besu")
	BlockchainNodeProviderRemoteRPC = fftypes.FFEnumValue(BlockchainNodeProvider, "remote-rpc")
)

const DatabaseSelection = "database_selection"

var (
	DatabaseSelectionSQLite   = fftypes.FFEnumValue(DatabaseSelection, "sqlite")
	DatabaseSelectionPostgres = fftypes.FFEnumValue(DatabaseSelection, "postgres")
)

const TokenProvider = "token_provider"

var (
	TokenProviderNone         = fftypes.FFEnumValue(TokenProvider, "none")
	TokenProviderERC1155      = fftypes.FFEnumValue(TokenProvider, "erc1155")
	TokenProviderERC20_ERC721 = fftypes.FFEnumValue(TokenProvider, "erc20_erc721")
)

const ReleaseChannelSelection = "release_channel"

var (
	ReleaseChannelStable = fftypes.FFEnumValue(ReleaseChannelSelection, "stable")
	ReleaseChannelAlpha  = fftypes.FFEnumValue(ReleaseChannelSelection, "alpha")
	ReleaseChannelBeta   = fftypes.FFEnumValue(ReleaseChannelSelection, "beta")
	ReleaseChannelRC     = fftypes.FFEnumValue(ReleaseChannelSelection, "rc")
)

func FFEnumArray(ctx context.Context, a []string) ([]fftypes.FFEnum, error) {
	enums := make([]fftypes.FFEnum, len(a))
	for i, v := range a {
		enums[i] = fftypes.FFEnum(v)
	}
	return enums, nil
}

func FFEnumArrayToStrings(input []fftypes.FFEnum) []string {
	s := make([]string, len(input))
	for i, e := range input {
		s[i] = e.String()
	}
	return s
}
