// Copyright © 2025 Kaleido, Inc.
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
	StackName                 string
	MemberCount               int
	FireFlyBasePort           int
	ServicesBasePort          int
	PtmBasePort               int
	DatabaseProvider          string
	ExternalProcesses         int
	OrgNames                  []string
	NodeNames                 []string
	BlockchainConnector       string
	BlockchainProvider        string
	BlockchainNodeProvider    string
	PrivateTransactionManager string
	Consensus                 string
	TokenProviders            []string
	FireFlyVersion            string
	ManifestPath              string
	PrometheusEnabled         bool
	PrometheusPort            int
	SandboxEnabled            bool
	ExtraCoreConfigPath       string
	ExtraConnectorConfigPath  string
	BlockPeriod               int
	ContractAddress           string
	RemoteNodeURL             string
	ChainID                   int64
	Network                   string
	Socket                    string
	BlockfrostKey             string
	BlockfrostBaseURL         string
	DisableTokenFactories     bool
	RequestTimeout            int
	ReleaseChannel            string
	MultipartyEnabled         bool
	IPFSMode                  string
	CCPYAMLPaths              []string
	MSPPaths                  []string
	ChannelName               string
	ChaincodeName             string
	CustomPinSupport          bool
	RemoteNodeDeploy          bool
	EnvironmentVars           map[string]string
}

const IPFSMode = "ipfs_mode"

var (
	IPFSModePrivate = fftypes.FFEnumValue(IPFSMode, "private")
	IPFSModePublic  = fftypes.FFEnumValue(IPFSMode, "public")
)

const BlockchainProvider = "blockchain_provider"

var (
	BlockchainProviderCardano  = fftypes.FFEnumValue(BlockchainProvider, "cardano")
	BlockchainProviderEthereum = fftypes.FFEnumValue(BlockchainProvider, "ethereum")
	BlockchainProviderTezos    = fftypes.FFEnumValue(BlockchainProvider, "tezos")
	BlockchainProviderFabric   = fftypes.FFEnumValue(BlockchainProvider, "fabric")
	BlockchainProviderCorda    = fftypes.FFEnumValue(BlockchainProvider, "corda")
)

const BlockchainConnector = "blockchain_connector"

var (
	BlockchainConnectorCardanoConnect = fftypes.FFEnumValue(BlockchainConnector, "cardanoconnect")
	BlockchainConnectorEthconnect     = fftypes.FFEnumValue(BlockchainConnector, "ethconnect")
	BlockchainConnectorEvmconnect     = fftypes.FFEnumValue(BlockchainConnector, "evmconnect")
	BlockchainConnectorTezosconnect   = fftypes.FFEnumValue(BlockchainConnector, "tezosconnect")
	BlockchainConnectorFabconnect     = fftypes.FFEnumValue(BlockchainConnector, "fabric")
)

const BlockchainNodeProvider = "blockchain_node_provider"

var (
	BlockchainNodeProviderGeth      = fftypes.FFEnumValue(BlockchainNodeProvider, "geth")
	BlockchainNodeProviderQuorum    = fftypes.FFEnumValue(BlockchainNodeProvider, "quorum")
	BlockchainNodeProviderBesu      = fftypes.FFEnumValue(BlockchainNodeProvider, "besu")
	BlockchainNodeProviderRemoteRPC = fftypes.FFEnumValue(BlockchainNodeProvider, "remote-rpc")
)

const Consensus = "consensus"

var (
	ConsensusClique = fftypes.FFEnumValue(Consensus, "clique")
	ConsensusRaft   = fftypes.FFEnumValue(Consensus, "raft")
	ConsensusIbft   = fftypes.FFEnumValue(Consensus, "ibft")
	ConsensusQbft   = fftypes.FFEnumValue(Consensus, "qbft")
)

const PrivateTransactionManager = "private_transaction_manager"

var (
	PrivateTransactionManagerNone    = fftypes.FFEnumValue(PrivateTransactionManager, "none")
	PrivateTransactionManagerTessera = fftypes.FFEnumValue(PrivateTransactionManager, "tessera")
)

const DatabaseSelection = "database_selection"

var (
	DatabaseSelectionSQLite   = fftypes.FFEnumValue(DatabaseSelection, "sqlite3")
	DatabaseSelectionPostgres = fftypes.FFEnumValue(DatabaseSelection, "postgres")
)

const TokenProvider = "token_provider"

var (
	TokenProviderNone        = fftypes.FFEnumValue(TokenProvider, "none")
	TokenProviderERC1155     = fftypes.FFEnumValue(TokenProvider, "erc1155")
	TokenProviderERC20ERC721 = fftypes.FFEnumValue(TokenProvider, "erc20_erc721")
)

const ReleaseChannelSelection = "release_channel"

var (
	ReleaseChannelStable = fftypes.FFEnumValue(ReleaseChannelSelection, "stable")
	ReleaseChannelHead   = fftypes.FFEnumValue(ReleaseChannelSelection, "head")
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
	s := make([]string, 0)
	for _, e := range input {
		if e != "none" {
			s = append(s, e.String())
		}
	}
	return s
}
