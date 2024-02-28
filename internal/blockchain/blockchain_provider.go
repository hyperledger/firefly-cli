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

package blockchain

import (
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type IBlockchainProvider interface {
	WriteConfig(options *types.InitOptions) error
	FirstTimeSetup() error
	DeployFireFlyContract() (*types.ContractDeploymentResult, error)
	PreStart() error
	PostStart(firstTimeSetup bool) error
	GetDockerServiceDefinitions() []*docker.ServiceDefinition
	GetBlockchainPluginConfig(stack *types.Stack, org *types.Organization) (blockchainConfig *types.BlockchainConfig)
	GetOrgConfig(stack *types.Stack, org *types.Organization) (coreConfig *types.OrgConfig)
	Reset() error
	GetContracts(filename string, extraArgs []string) ([]string, error)
	DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error)
	CreateAccount(args []string) (interface{}, error)
	ParseAccount(interface{}) interface{}
	GetConnectorName() string
	GetConnectorURL(org *types.Organization) string
	GetConnectorExternalURL(org *types.Organization) string
}
