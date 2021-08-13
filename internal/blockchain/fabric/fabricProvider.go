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

package fabric

import (
	"fmt"

	"github.com/hyperledger-labs/firefly-cli/internal/core"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/internal/log"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type FabricProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

func (p *FabricProvider) WriteConfig() error {
	return nil
}

func (p *FabricProvider) RunFirstTimeSetup() error {
	/*
		TODO:
		1) Start fab-ca
		2) Generate MSP for orderer, peer
		3) Run orderer and peer
		4) Create signers for client POST /identities <-- this may be done by firefly and may not need to be in this file
	*/

	return nil
}

func (p *FabricProvider) DeploySmartContracts() error {
	// TODO: figure out how to deploy fabric chaincode
	return nil
}

func (p *FabricProvider) PreStart() error {
	return nil
}

func (p *FabricProvider) PostStart() error {
	return nil
}

func (p *FabricProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 2)

	// TODO: figure out the fabric command line
	fabricPeerCommand := ""
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "fabric-peer",
		Service: &docker.Service{
			Image:   "hyperledger/fabric-peer:latest",
			Command: fabricPeerCommand,
			Volumes: []string{"fabric-peer:/data"},
			Logging: docker.StandardLogOptions,
			// TODO: Which ports do we need here?
			Ports: []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"fabric-peer"},
	}

	// TODO: figure out the fabric command line
	fabricOrdererCommand := ""
	serviceDefinitions[1] = &docker.ServiceDefinition{
		ServiceName: "fabric-orderer",
		Service: &docker.Service{
			Image:   "hyperledger/fabric-orderer:latest",
			Command: fabricOrdererCommand,
			Volumes: []string{"fabric-orderer:/data"},
			Logging: docker.StandardLogOptions,
			// TODO: Which ports do we need here?
			Ports: []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"fabric-orderer"},
	}

	serviceDefinitions = append(serviceDefinitions, p.getFabconnectServiceDefinitions(p.Stack.Members)...)

	return serviceDefinitions
}

func (p *FabricProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
	return &core.BlockchainConfig{
		Type: "fabric",
		Fabric: &core.FabricConfig{
			Fabconnect: &core.EthconnectConfig{
				URL:      p.getFabconnectUrl(m),
				Instance: "/contracts/firefly",
				Topic:    m.ID,
			},
		},
	}
}

func (p *FabricProvider) getFabconnectServiceDefinitions(members []*types.Member) []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, len(members))
	for i, member := range members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "fabconnect_" + member.ID,
			Service: &docker.Service{
				Image: "ghcr.io/hyperledger-labs/firefly-fabconnect:latest",
				// TODO: Figure out the correct command line parameters for fabconnect
				Command: "rest -U http://127.0.0.1:8080 -I ./abis -r http://fabric-peer:8545 -E ./events -d 3",
				DependsOn: map[string]map[string]string{
					"fabric-peer":    {"condition": "service_started"},
					"fabric-orderer": {"condition": "service_started"},
				},
				Ports: []string{fmt.Sprintf("%d:8080", member.ExposedEthconnectPort)},
				Volumes: []string{
					fmt.Sprintf("ethconnect_events_%s:/ethconnect/events", member.ID),
				},
				Logging: docker.StandardLogOptions,
			},
			VolumeNames: []string{"ethconnect_events_" + member.ID},
		}
	}
	return serviceDefinitions
}

func (p *FabricProvider) getFabconnectUrl(member *types.Member) string {
	if !member.External {
		// TODO: verify this is the correct port
		return fmt.Sprintf("http://fabconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	}
}
