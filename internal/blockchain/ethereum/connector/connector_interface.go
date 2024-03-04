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

package connector

import (
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethtypes"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type Connector interface {
	FirstTimeSetup(stack *types.Stack) error
	GetServiceDefinitions(s *types.Stack, dependentServices map[string]string) []*docker.ServiceDefinition
	DeployContract(contract *ethtypes.CompiledContract, contractName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error)
	GenerateConfig(stack *types.Stack, member *types.Organization, blockchainServiceName string) Config
	Name() string
	Port() int
}

type Config interface {
	WriteConfig(filename string, extraConnectorConfigPath string) error
}
