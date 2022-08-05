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

package erc1155

import (
	"context"
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/blockchain"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

const tokenProviderName = "erc1155"

type ERC1155Provider struct {
	ctx                context.Context
	stack              *types.Stack
	blockchainProvider blockchain.IBlockchainProvider
}

func NewERC1155Provider(ctx context.Context, stack *types.Stack, blockchainProvider blockchain.IBlockchainProvider) *ERC1155Provider {
	return &ERC1155Provider{
		ctx:                ctx,
		stack:              stack,
		blockchainProvider: blockchainProvider,
	}
}

func (p *ERC1155Provider) DeploySmartContracts(tokenIndex int) (*types.ContractDeploymentResult, error) {
	return DeployContracts(p.ctx, p.stack, tokenIndex)
}

func (p *ERC1155Provider) FirstTimeSetup(tokenIdx int) error {
	l := log.LoggerFromContext(p.ctx)
	for _, member := range p.stack.Members {
		l.Info(fmt.Sprintf("initializing tokens on member %s", member.ID))
		tokenInitUrl := fmt.Sprintf("http://localhost:%d/api/v1/init", member.ExposedTokensPorts[tokenIdx])
		if err := core.RequestWithRetry(p.ctx, "POST", tokenInitUrl, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (p *ERC1155Provider) GetDockerServiceDefinitions(tokenIdx int) []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 0, len(p.stack.Members))
	for i, member := range p.stack.Members {
		connectorName := fmt.Sprintf("tokens_%v_%v", member.ID, tokenIdx)
		env := map[string]interface{}{
			"ETHCONNECT_URL":   p.getEthconnectURL(member),
			"ETHCONNECT_TOPIC": connectorName,
			"AUTO_INIT":        "false",
		}
		if p.stack.FFTMEnabled {
			env["FFTM_URL"] = p.getFFTMURL(member)
		}
		serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
			ServiceName: connectorName,
			Service: &docker.Service{
				Image:         p.stack.VersionManifest.TokensERC1155.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_tokens_%v_%v", p.stack.Name, i, tokenIdx),
				Ports:         []string{fmt.Sprintf("%d:3000", member.ExposedTokensPorts[tokenIdx])},
				Environment:   env,
				DependsOn: map[string]map[string]string{
					"ethconnect_" + member.ID: {"condition": "service_started"},
				},
				HealthCheck: &docker.HealthCheck{
					Test: []string{"CMD", "curl", "http://localhost:3000/api"},
				},
				Logging: docker.StandardLogOptions,
			},
		})
	}
	return serviceDefinitions
}

func (p *ERC1155Provider) GetFireflyConfig(m *types.Organization, tokenIdx int) *types.TokensConfig {
	name := tokenProviderName
	if tokenIdx > 0 {
		name = fmt.Sprintf("%s_%d", tokenProviderName, tokenIdx)
	}
	return &types.TokensConfig{
		Type: "fftokens",
		Name: name,
		FFTokens: &types.FFTokensConfig{
			URL: p.getTokensURL(m, tokenIdx),
		},
	}
}

func (p *ERC1155Provider) getEthconnectURL(member *types.Organization) string {
	return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
}

func (p *ERC1155Provider) getFFTMURL(member *types.Organization) string {
	return fmt.Sprintf("http://fftm_%s:5008", member.ID)
}

func (p *ERC1155Provider) getTokensURL(member *types.Organization, tokenIdx int) string {
	if !member.External {
		return fmt.Sprintf("http://tokens_%s_%d:3000", member.ID, tokenIdx)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedTokensPorts[tokenIdx])
	}
}

func (p *ERC1155Provider) GetName() string {
	return tokenProviderName
}
