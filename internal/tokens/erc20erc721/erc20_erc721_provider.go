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

package erc20erc721

import (
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"gopkg.in/yaml.v3"
)

const tokenProviderName = "erc20_erc721"

type ERC20ERC721Provider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

type HexAddress string

// Explicitly quote hex addresses so that they are interpreted as string (not int)
func (h HexAddress) MarshalYAML() (interface{}, error) {
	return yaml.Node{
		Value: string(h),
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}, nil
}

func (p *ERC20ERC721Provider) DeploySmartContracts(tokenIndex int) (*types.ContractDeploymentResult, error) {
	return DeployContracts(p.Stack, p.Log, p.Verbose, tokenIndex)
}

func (p *ERC20ERC721Provider) FirstTimeSetup(tokenIdx int) error {
	for _, member := range p.Stack.Members {
		p.Log.Info(fmt.Sprintf("initializing tokens on member %s", member.ID))
		tokenInitUrl := fmt.Sprintf("http://localhost:%d/api/v1/init", member.ExposedTokensPorts[tokenIdx])
		if err := core.RequestWithRetry("POST", tokenInitUrl, nil, nil, p.Verbose); err != nil {
			return err
		}
	}
	return nil
}

func (p *ERC20ERC721Provider) GetDockerServiceDefinitions(tokenIdx int) []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 0, len(p.Stack.Members))
	for i, member := range p.Stack.Members {
		connectorName := fmt.Sprintf("tokens_%v_%v", member.ID, tokenIdx)

		var factoryAddress HexAddress
		for _, contract := range p.Stack.State.DeployedContracts {
			if contract.Name == contractName(tokenIdx) {
				switch loc := contract.Location.(type) {
				case map[string]string:
					factoryAddress = HexAddress(loc["address"])
				}
			}
		}

		env := map[string]interface{}{
			"ETHCONNECT_URL":           p.getEthconnectURL(member),
			"ETHCONNECT_TOPIC":         connectorName,
			"FACTORY_CONTRACT_ADDRESS": factoryAddress,
			"AUTO_INIT":                "false",
		}
		if p.Stack.FFTMEnabled {
			env["FFTM_URL"] = p.getFFTMURL(member)
		}

		serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
			ServiceName: connectorName,
			Service: &docker.Service{
				Image:         p.Stack.VersionManifest.TokensERC20ERC721.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_tokens_%v_%v", p.Stack.Name, i, tokenIdx),
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

func (p *ERC20ERC721Provider) GetFireflyConfig(m *types.Organization, tokenIdx int) *types.TokensConfig {
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

func (p *ERC20ERC721Provider) getEthconnectURL(member *types.Organization) string {
	return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
}

func (p *ERC20ERC721Provider) getFFTMURL(member *types.Organization) string {
	return fmt.Sprintf("http://fftm_%s:5008", member.ID)
}

func (p *ERC20ERC721Provider) getTokensURL(member *types.Organization, tokenIdx int) string {
	if !member.External {
		return fmt.Sprintf("http://tokens_%s_%d:3000", member.ID, tokenIdx)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedTokensPorts[tokenIdx])
	}
}

func (p *ERC20ERC721Provider) GetName() string {
	return tokenProviderName
}
