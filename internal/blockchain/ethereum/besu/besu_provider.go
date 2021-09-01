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

package besu

import (
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger-labs/firefly-cli/internal/core"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/internal/log"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type BesuProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

func (p *BesuProvider) WriteConfig() error {
	return nil
}

func (p *BesuProvider) FirstTimeSetup() error {
	return nil
}

func (p *BesuProvider) DeploySmartContracts() error {
	return ethereum.DeployContracts(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart() error {
	return nil
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "besu",
		Service:     &docker.Service{},
	}
	return serviceDefinitions
}

func (p *BesuProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
	return &core.BlockchainConfig{}
}

func (p *BesuProvider) Reset() error {
	return nil
}
