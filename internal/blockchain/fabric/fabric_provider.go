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
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

//go:embed configtx.yaml
var configtxYaml string

//go:embed config.yaml
var configYaml string

func (p *FabricProvider) WriteConfig() error {
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	cryptogenYamlPath := path.Join(blockchainDirectory, "cryptogen.yaml")

	if err := WriteCryptogenConfig(len(p.Stack.Members), cryptogenYamlPath); err != nil {
		return err
	}
	if err := WriteNetworkConfig(path.Join(blockchainDirectory, "ccp.yaml")); err != nil {
		return err
	}
	if err := fabconnect.WriteFabconnectConfig(path.Join(blockchainDirectory, "fabconnect.yaml")); err != nil {
		return err
	}
	if err := p.writeConfigtxYaml(); err != nil {
		return err
	}

	return nil
}

func (p *FabricProvider) FirstTimeSetup() error {
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	organizationsDirectory := path.Join(blockchainDirectory, "organizations")
	cryptogenYamlPath := path.Join(blockchainDirectory, "cryptogen.yaml")

	if err := os.MkdirAll(organizationsDirectory, 0755); err != nil {
		return err
	}

	// Run cryptogen to generate MSP
	if err := docker.RunDockerCommand(blockchainDirectory, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/etc/template.yml", cryptogenYamlPath), "-v", fmt.Sprintf("%s:/output", organizationsDirectory), "hyperledger/fabric-tools:2.3", "cryptogen", "generate", "--config", "/etc/template.yml", "--output", "/output"); err != nil {
		return err
	}

	// Generate genesis block
	if err := docker.RunDockerCommand(blockchainDirectory, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/firefly", blockchainDirectory), "-v", fmt.Sprintf("%s:/etc/hyperledger/fabric/configtx.yaml", path.Join(blockchainDirectory, "configtx.yaml")), "hyperledger/fabric-tools:2.3", "configtxgen", "-outputBlock", "/firefly/firefly.block", "-profile", "SingleOrgApplicationGenesis", "-channelID", "firefly"); err != nil {
		return err
	}

	if err := p.writeConfigYaml(); err != nil {
		return err
	}

	return nil
}

func (p *FabricProvider) DeploySmartContracts() error {
	if err := p.extractChaincode(); err != nil {
		return err
	}

	if err := p.createChannel(); err != nil {
		return err
	}

	if err := p.joinChannel(); err != nil {
		return err
	}

	if err := p.installChaincode(); err != nil {
		return err
	}

	res, err := p.queryInstalled()
	if err != nil {
		return err
	}
	if len(res.InstalledChaincodes) == 0 {
		return fmt.Errorf("failed to find installed chaincode")
	}

	if err := p.approveChaincode(res.InstalledChaincodes[0].PackageID); err != nil {
		return err
	}

	if err := p.commitChaincode(); err != nil {
		return err
	}

	if err := p.registerIdentities(); err != nil {
		return err
	}

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
				URL:       p.getFabconnectUrl(m),
				Chaincode: "firefly",
				Channel:   "firefly",
				Signer:    m.Address,
				Topic:     m.ID,
			},
		},
	}
}

func (p *FabricProvider) Reset() error {
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	organizationsDirectory := path.Join(blockchainDirectory, "organizations")
	if err := os.RemoveAll(organizationsDirectory); err != nil {
		return err
	}
	return nil
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
					fmt.Sprintf("fabconnect_receipts_%s:/fabconnect/receipts", member.ID),
					fmt.Sprintf("fabconnect_events_%s:/fabconnect/events", member.ID),
					fmt.Sprintf("%s:/fabconnect/fabconnect.yaml", path.Join(blockchainDirectory, "fabconnect.yaml")),
					fmt.Sprintf("%s:/fabconnect/ccp.yaml", path.Join(blockchainDirectory, "ccp.yaml")),
					fmt.Sprintf("%s:/fabconnect/organizations", path.Join(blockchainDirectory, "organizations")),
					fmt.Sprintf("%s:/fabconnect/ca-cert.pem", path.Join(blockchainDirectory, "organizations", "peerOrganizations", "org1.example.com", "ca", "fabric_ca.org1.example.com-cert.pem")),
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

func (p *FabricProvider) writeConfigtxYaml() error {
	filePath := path.Join(constants.StacksDir, p.Stack.Name, "blockchain", "configtx.yaml")
	return ioutil.WriteFile(filePath, []byte(configtxYaml), 0755)
}

func (p *FabricProvider) writeConfigYaml() error {
	filePath := path.Join(constants.StacksDir, p.Stack.Name, "blockchain", "organizations", "peerOrganizations", "org1.example.com", "peers", "fabric_peer.org1.example.com", "msp", "config.yaml")
	return ioutil.WriteFile(filePath, []byte(configYaml), 0755)
}

func (p *FabricProvider) createChannel() error {
	p.Log.Info("creating channel")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "osnadmin", "channel", "join", "--channelID", "firefly", "--config-block", "/firefly/firefly.block", "-o", "fabric_orderer:7053", "--ca-file", "/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/ca.crt", "--client-cert", "/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.crt", "--client-key", "/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.key")
}

func (p *FabricProvider) joinChannel() error {
	p.Log.Info("joining channel")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "peer", "channel", "join", "-b", "/firefly/firefly.block")
}

func (p *FabricProvider) extractChaincode() error {
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	contractsDir := path.Join(stackDir, "contracts")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return err
	}

	var containerName string
	for _, member := range p.Stack.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_firefly_core_%s_1", p.Stack.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	p.Log.Info("extracting smart contracts")
	if err := docker.CopyFromContainer(containerName, "/firefly/contracts/firefly_fabric.tar.gz", path.Join(contractsDir, "firefly_fabric.tar.gz"), p.Verbose); err != nil {
		return err
	}
	return nil
}

func (p *FabricProvider) installChaincode() error {
	p.Log.Info("installing chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/contracts", path.Join(stackDir, "contracts")), "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "install", "/contracts/firefly_fabric.tar.gz")
}

func (p *FabricProvider) queryInstalled() (*QueryInstalledResponse, error) {
	p.Log.Info("querying installed chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	str, err := docker.RunDockerCommandBuffered(stackDir, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "queryinstalled", "--output", "json")
	if err != nil {
		return nil, err
	}
	var res *QueryInstalledResponse
	err = json.Unmarshal([]byte(str), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p *FabricProvider) approveChaincode(packageId string) error {
	p.Log.Info("approving chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "approveformyorg", "-o", "fabric_orderer:7050", "--ordererTLSHostnameOverride", "fabric_orderer", "--channelID", "firefly", "--name", "firefly", "--version", "1.0", "--package-id", packageId, "--sequence", "1", "--tls", "--cafile", "/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem")
}

func (p *FabricProvider) commitChaincode() error {
	p.Log.Info("committing chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/firefly", path.Join(stackDir, "blockchain")), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "commit", "-o", "fabric_orderer:7050", "--ordererTLSHostnameOverride", "fabric_orderer", "--channelID", "firefly", "--name", "firefly", "--version", "1.0", "--sequence", "1", "--tls", "--cafile", "/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem")
}

func (p *FabricProvider) registerIdentities() error {
	p.Log.Info("registering identities")
	for _, m := range p.Stack.Members {
		res, err := fabconnect.CreateIdentity(fmt.Sprintf("http://127.0.0.1:%v", m.ExposedEthconnectPort), m.Address)
		if err != nil {
			return err
		}
		_, err = fabconnect.EnrollIdentity(fmt.Sprintf("http://127.0.0.1:%v", m.ExposedEthconnectPort), m.Address, res.Secret)
		if err != nil {
			return err
		}
	}
	return nil
}
