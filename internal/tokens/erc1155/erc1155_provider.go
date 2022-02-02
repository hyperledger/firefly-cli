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
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type ERC1155Provider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *ERC1155Provider) DeploySmartContracts(tokenIndex int) error {
	return DeployContracts(p.Stack, p.Log, p.Verbose, tokenIndex)
}

func (p *ERC1155Provider) FirstTimeSetup(tokenIdx int) error {
	for _, member := range p.Stack.Members {
		p.Log.Info(fmt.Sprintf("initializing tokens on member %s", member.ID))
		tokenInitUrl := fmt.Sprintf("http://localhost:%d/api/v1/init", member.ExposedTokensPorts[tokenIdx])
		if err := core.RequestWithRetry("POST", tokenInitUrl, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (p *ERC1155Provider) GetDockerServiceDefinitions(tokenIdx int) []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 0, len(p.Stack.Members))
	for i, member := range p.Stack.Members {
		serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
			ServiceName: fmt.Sprintf("tokens_%v_%v", member.ID, tokenIdx),
			Service: &docker.Service{
				Image:         p.Stack.VersionManifest.TokensERC1155.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_tokens_%v_%v", p.Stack.Name, i, tokenIdx),
				Ports:         []string{fmt.Sprintf("%d:3000", member.ExposedTokensPorts[tokenIdx])},
				Environment: map[string]string{
					"ETHCONNECT_URL":      p.getEthconnectURL(member, member.ExposedTokensPorts[tokenIdx]),
					"ETHCONNECT_INSTANCE": "/contracts/erc1155",
					"AUTO_INIT":           "false",
				},
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

func (p *ERC1155Provider) GetFireflyConfig(m *types.Member, tokenIdx int) *core.TokenConnector {
	return &core.TokenConnector{
		Plugin: "fftokens",
		Name:   "erc1155",
		URL:    p.getTokensURL(m, tokenIdx),
	}
}

func (p *ERC1155Provider) getEthconnectURL(member *types.Member, tokenIdx int) string {
	return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
}

func (p *ERC1155Provider) getTokensURL(member *types.Member, tokenIdx int) string {
	if !member.External {
		return fmt.Sprintf("http://tokens_%s_%d:3000", member.ID, tokenIdx)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedTokensPorts[tokenIdx])
	}
}
