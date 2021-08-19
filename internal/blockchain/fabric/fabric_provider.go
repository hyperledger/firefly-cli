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
	"path"

	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/fabric/fabconnect"
	"github.com/hyperledger-labs/firefly-cli/internal/constants"
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
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	cryptogenDirectory := path.Join(blockchainDirectory, "cryptogen")
	if err := WriteCryptogenConfig(len(p.Stack.Members), cryptogenDirectory); err != nil {
		return err
	}
	if err := WriteNetworkConfig(blockchainDirectory, path.Join(blockchainDirectory, "fabric.yaml")); err != nil {
		return err
	}
	return fabconnect.WriteFabconnectConfig(path.Join(blockchainDirectory, "fabconnect.yaml"))
}

func (p *FabricProvider) RunFirstTimeSetup() error {
	/*
		TODO:
		1) Start fab-ca
		2) Generate MSP for orderer, peer
		3) Run orderer and peer
		4) Create signers for client POST /identities <-- this may be done by firefly and may not need to be in this file
	*/

	// docker cp cc66fbd7106f:/etc/hyperledger/fabric-ca-server/tls-cert.pem ~/Desktop/fabric/fabric-ca-tls-cert.pem
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
	serviceDefinitions := GenerateDockerServiceDefinitions(p.Stack)
	serviceDefinitions = append(serviceDefinitions, p.getFabconnectServiceDefinitions(p.Stack.Members)...)
	return serviceDefinitions
}

func (p *FabricProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
	return &core.BlockchainConfig{
		Type: "fabric",
		Fabric: &core.FabricConfig{
			Fabconnect: &core.FabconnectConfig{
				URL:      p.getFabconnectUrl(m),
				Instance: "/contracts/firefly",
				Topic:    m.ID,
			},
		},
	}
}

func (p *FabricProvider) getFabconnectServiceDefinitions(members []*types.Member) []*docker.ServiceDefinition {
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	serviceDefinitions := make([]*docker.ServiceDefinition, len(members))
	for i, member := range members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "fabconnect_" + member.ID,
			Service: &docker.Service{
				Image:   "ghcr.io/hyperledger-labs/firefly-fabconnect:latest",
				Command: "-f /fabconnect/fabconnect.yaml",
				DependsOn: map[string]map[string]string{
					"fabric_ca":      {"condition": "service_started"},
					"fabric_peer":    {"condition": "service_started"},
					"fabric_orderer": {"condition": "service_started"},
				},
				Ports: []string{fmt.Sprintf("%d:3000", member.ExposedEthconnectPort)},
				Volumes: []string{
					fmt.Sprintf("fabconnect_receipts%s:/fabconnect/receipts", member.ID),
					fmt.Sprintf("fabconnect_events%s:/fabconnect/events", member.ID),
					fmt.Sprintf("%s:/fabconnect/fabconnect.yaml", path.Join(blockchainDirectory, "fabconnect.yaml")),
					fmt.Sprintf("%s:/fabconnect/fabric.yaml", path.Join(blockchainDirectory, "fabric.yaml")),
				},
				Logging: docker.StandardLogOptions,
			},
			VolumeNames: []string{
				"fabconnect_receipts_" + member.ID,
				"fabconnect_events_" + member.ID,
			},
		}
	}
	return serviceDefinitions
}

func (p *FabricProvider) getFabconnectUrl(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://fabconnect_%s:3000", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	}
}
