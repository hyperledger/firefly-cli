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

	"github.com/hyperledger/firefly-cli/internal/blockchain/fabric/fabconnect"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type FabricProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

//go:embed configtx.yaml
var configtxYaml string

func (p *FabricProvider) WriteConfig(options *types.InitOptions) error {
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
	cryptogenYamlPath := path.Join(blockchainDirectory, "cryptogen.yaml")
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)

	if err := docker.CreateVolume(volumeName, p.Verbose); err != nil {
		return err
	}

	// Run cryptogen to generate MSP
	if err := docker.RunDockerCommand(blockchainDirectory, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", "-v", fmt.Sprintf("%s:/etc/template.yml", cryptogenYamlPath), "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "cryptogen", "generate", "--config", "/etc/template.yml", "--output", "/etc/firefly/organizations"); err != nil {
		return err
	}

	// Generate genesis block
	if err := docker.RunDockerCommand(blockchainDirectory, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "-v", fmt.Sprintf("%s:/etc/hyperledger/fabric/configtx.yaml", path.Join(blockchainDirectory, "configtx.yaml")), "hyperledger/fabric-tools:2.3", "configtxgen", "-outputBlock", "/etc/firefly/firefly.block", "-profile", "SingleOrgApplicationGenesis", "-channelID", "firefly"); err != nil {
		return err
	}

	return nil
}

func (p *FabricProvider) DeploySmartContracts() ([]byte, error) {
	// No config patch YAML required for Fabric, as the chaincode name is pre-determined
	return nil, p.deploySmartContracts()
}

func (p *FabricProvider) deploySmartContracts() error {
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

func (p *FabricProvider) GetFireflyConfig(stack *types.Stack, m *types.Member) (blockchainConfig *core.BlockchainConfig, orgConfig *core.OrgConfig) {
	orgConfig = &core.OrgConfig{
		Name:     m.OrgName,
		Identity: m.OrgName,
	}

	blockchainConfig = &core.BlockchainConfig{
		Type: "fabric",
		Fabric: &core.FabricConfig{
			Fabconnect: &core.FabconnectConfig{
				URL:       p.getFabconnectUrl(m),
				Chaincode: "firefly",
				Channel:   "firefly",
				Signer:    m.OrgName,
				Topic:     m.ID,
			},
		},
	}
	return
}

func (p *FabricProvider) Reset() error {
	return nil
}

func (p *FabricProvider) getFabconnectServiceDefinitions(members []*types.Member) []*docker.ServiceDefinition {
	blockchainDirectory := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")
	serviceDefinitions := make([]*docker.ServiceDefinition, len(members))
	for i, member := range members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "fabconnect_" + member.ID,
			Service: &docker.Service{
				Image:         p.Stack.VersionManifest.Fabconnect.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_fabconnect_%s", p.Stack.Name, member.ID),
				Command:       "-f /fabconnect/fabconnect.yaml",
				DependsOn: map[string]map[string]string{
					"fabric_ca":      {"condition": "service_started"},
					"fabric_peer":    {"condition": "service_started"},
					"fabric_orderer": {"condition": "service_started"},
				},
				Ports: []string{fmt.Sprintf("%d:3000", member.ExposedConnectorPort)},
				Volumes: []string{
					fmt.Sprintf("fabconnect_receipts_%s:/fabconnect/receipts", member.ID),
					fmt.Sprintf("fabconnect_events_%s:/fabconnect/events", member.ID),
					fmt.Sprintf("%s:/fabconnect/fabconnect.yaml", path.Join(blockchainDirectory, "fabconnect.yaml")),
					fmt.Sprintf("%s:/fabconnect/ccp.yaml", path.Join(blockchainDirectory, "ccp.yaml")),
					"firefly_fabric:/etc/firefly",
				},
				Logging: docker.StandardLogOptions,
			},
			VolumeNames: []string{
				"fabconnect_receipts_" + member.ID,
				"fabconnect_events_" + member.ID,
				"firefly_fabric",
			},
		}
	}
	return serviceDefinitions
}

func (p *FabricProvider) getFabconnectUrl(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://fabconnect_%s:3000", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}

func (p *FabricProvider) writeConfigtxYaml() error {
	filePath := path.Join(constants.StacksDir, p.Stack.Name, "blockchain", "configtx.yaml")
	return ioutil.WriteFile(filePath, []byte(configtxYaml), 0755)
}

func (p *FabricProvider) createChannel() error {
	p.Log.Info("creating channel")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "osnadmin", "channel", "join", "--channelID", "firefly", "--config-block", "/etc/firefly/firefly.block", "-o", "fabric_orderer:7053", "--ca-file", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/ca.crt", "--client-cert", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.crt", "--client-key", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.key")
}

func (p *FabricProvider) joinChannel() error {
	p.Log.Info("joining channel")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "hyperledger/fabric-tools:2.3", "peer", "channel", "join", "-b", "/etc/firefly/firefly.block")
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
			containerName = fmt.Sprintf("%s_firefly_core_%s", p.Stack.Name, member.ID)
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
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/contracts", path.Join(stackDir, "contracts")), "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "install", "/contracts/firefly_fabric.tar.gz")
}

func (p *FabricProvider) queryInstalled() (*QueryInstalledResponse, error) {
	p.Log.Info("querying installed chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	str, err := docker.RunDockerCommandBuffered(stackDir, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "queryinstalled", "--output", "json")
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
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "approveformyorg", "-o", "fabric_orderer:7050", "--ordererTLSHostnameOverride", "fabric_orderer", "--channelID", "firefly", "--name", "firefly", "--version", "1.0", "--package-id", packageId, "--sequence", "1", "--tls", "--cafile", "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem")
}

func (p *FabricProvider) commitChaincode() error {
	p.Log.Info("committing chaincode")
	stackDir := path.Join(constants.StacksDir, p.Stack.Name)
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.Stack.Name)
	return docker.RunDockerCommand(stackDir, p.Verbose, p.Verbose, "run", "--platform", getDockerPlatform(), "--rm", fmt.Sprintf("--network=%s_default", p.Stack.Name), "-e", "CORE_PEER_ADDRESS=fabric_peer:7051", "-e", "CORE_PEER_TLS_ENABLED=true", "-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt", "-e", "CORE_PEER_LOCALMSPID=Org1MSP", "-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp", "-v", fmt.Sprintf("%s:/etc/firefly", volumeName), "hyperledger/fabric-tools:2.3", "peer", "lifecycle", "chaincode", "commit", "-o", "fabric_orderer:7050", "--ordererTLSHostnameOverride", "fabric_orderer", "--channelID", "firefly", "--name", "firefly", "--version", "1.0", "--sequence", "1", "--tls", "--cafile", "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem")
}

func (p *FabricProvider) registerIdentities() error {
	p.Log.Info("registering identities")
	for _, m := range p.Stack.Members {
		res, err := fabconnect.CreateIdentity(fmt.Sprintf("http://127.0.0.1:%v", m.ExposedConnectorPort), m.OrgName)
		if err != nil {
			return err
		}
		_, err = fabconnect.EnrollIdentity(fmt.Sprintf("http://127.0.0.1:%v", m.ExposedConnectorPort), m.OrgName, res.Secret)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *FabricProvider) GetContracts(filename string) ([]string, error) {
	return []string{}, fmt.Errorf("deploying chaincode on a Fabric network is not supported yet")
}

func (p *FabricProvider) DeployContract(filename, contractName string, member types.Member) (string, error) {
	return "", fmt.Errorf("deploying chaincode on a Fabric network is not supported yet")
}

// As of release 2.4, Hyperledger Fabric only publishes amd64 images, but no arm64 specific images
func getDockerPlatform() string {
	return "linux/amd64"
}
