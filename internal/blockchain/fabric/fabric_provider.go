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

package fabric

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/blockchain/fabric/fabconnect"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	cp "github.com/otiai10/copy"
)

type Account struct {
	Name    string `json:"name"`
	OrgName string `json:"orgName"`
}

type FabricProvider struct {
	ctx   context.Context
	log   log.Logger
	stack *types.Stack
}

//go:embed configtx.yaml
var configtxYaml string

const chaincodeName = "firefly"
const chaincodeVersion = "1.0"
const channel = "firefly"

func NewFabricProvider(ctx context.Context, stack *types.Stack) *FabricProvider {
	return &FabricProvider{
		ctx:   ctx,
		stack: stack,
		log:   log.LoggerFromContext(ctx),
	}
}

func (p *FabricProvider) WriteConfig(options *types.InitOptions) error {
	blockchainDirectory := path.Join(p.stack.InitDir, "blockchain")

	if err := os.MkdirAll(blockchainDirectory, 0755); err != nil {
		return err
	}
	if p.stack.RemoteFabricNetwork {
		for i, member := range p.stack.Members {
			if err := cp.Copy(options.CCPYAMLPaths[i], path.Join(blockchainDirectory, fmt.Sprintf("%s_ccp.yaml", member.ID))); err != nil {
				return err
			}
			if err := cp.Copy(options.MSPPaths[i], path.Join(blockchainDirectory, fmt.Sprintf("%s_msp", member.ID))); err != nil {
				return err
			}
		}
	} else {
		cryptogenYamlPath := path.Join(blockchainDirectory, "cryptogen.yaml")

		if err := WriteCryptogenConfig(len(p.stack.Members), cryptogenYamlPath); err != nil {
			return err
		}
		if err := WriteNetworkConfig(path.Join(blockchainDirectory, "ccp.yaml")); err != nil {
			return err
		}
		if err := p.writeConfigtxYaml(); err != nil {
			return err
		}
	}

	if err := fabconnect.WriteFabconnectConfig(path.Join(blockchainDirectory, "fabconnect.yaml")); err != nil {
		return err
	}

	return nil
}

func (p *FabricProvider) FirstTimeSetup() error {
	if !p.stack.RemoteFabricNetwork {
		volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
		if err := docker.CreateVolume(p.ctx, volumeName); err != nil {
			return err
		}
		blockchainDirectory := path.Join(p.stack.RuntimeDir, "blockchain")
		cryptogenYamlPath := path.Join(blockchainDirectory, "cryptogen.yaml")

		// Run cryptogen to generate MSP
		if err := docker.RunDockerCommand(p.ctx, blockchainDirectory,
			"run",
			"--rm",
			"-v", fmt.Sprintf("%s:/etc/template.yml", cryptogenYamlPath),
			"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
			FabricToolsImageName,
			"cryptogen", "generate",
			"--config", "/etc/template.yml",
			"--output", "/etc/firefly/organizations",
		); err != nil {
			return err
		}

		// Generate genesis block
		if err := docker.RunDockerCommand(p.ctx, blockchainDirectory,
			"run",
			"--rm",
			"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
			"-v", fmt.Sprintf("%s:/etc/hyperledger/fabric/configtx.yaml", path.Join(blockchainDirectory, "configtx.yaml")),
			FabricToolsImageName,
			"configtxgen",
			"-outputBlock", "/etc/firefly/firefly.block",
			"-profile", "SingleOrgApplicationGenesis",
			"-channelID", "firefly",
		); err != nil {
			return err
		}
	}

	return nil
}

func (p *FabricProvider) DeployFireFlyContract() (*types.ContractDeploymentResult, error) {
	// No config patch YAML required for Fabric, as the chaincode name is pre-determined
	if p.stack.RemoteFabricNetwork {
		return &types.ContractDeploymentResult{
			DeployedContract: &types.DeployedContract{
				Name: p.stack.ChaincodeName,
				Location: map[string]string{
					"channel":   p.stack.ChannelName,
					"chaincode": p.stack.ChaincodeName,
				},
			},
		}, nil
	}
	result, err := p.deploySmartContracts()
	return result, err
}

func (p *FabricProvider) deploySmartContracts() (*types.ContractDeploymentResult, error) {
	if p.stack.RemoteFabricNetwork {
		return nil, fmt.Errorf("deploying chaincode is not supported with a remote Fabric network")
	}
	packageFilename := path.Join(p.stack.RuntimeDir, "contracts", "firefly_fabric.tar.gz")

	if err := p.extractChaincode(); err != nil {
		return nil, err
	}

	if err := p.installChaincode(packageFilename); err != nil {
		return nil, err
	}

	res, err := p.queryInstalled()
	if err != nil {
		return nil, err
	}
	if len(res.InstalledChaincodes) == 0 {
		return nil, fmt.Errorf("failed to find installed chaincode")
	}

	if err := p.approveChaincode(channel, chaincodeName, chaincodeVersion, res.InstalledChaincodes[0].PackageID); err != nil {
		return nil, err
	}

	if err := p.commitChaincode(channel, chaincodeName, chaincodeVersion); err != nil {
		return nil, err
	}

	result := &types.ContractDeploymentResult{
		DeployedContract: &types.DeployedContract{
			Name: "FireFly",
			Location: map[string]string{
				"channel":   channel,
				"chaincode": chaincodeName,
			},
		},
	}
	return result, nil
}

func (p *FabricProvider) PreStart() error {
	return nil
}

func (p *FabricProvider) PostStart(firstTimeSetup bool) error {
	if firstTimeSetup {
		if !p.stack.RemoteFabricNetwork {
			if err := p.createChannel(); err != nil {
				return err
			}

			if err := p.joinChannel(); err != nil {
				return err
			}
		}

		// Register pre-created identities
		p.log.Info("registering identities")
		for _, m := range p.stack.Members {
			_, err := p.registerIdentity(m, m.OrgName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *FabricProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	if p.stack.RemoteFabricNetwork {
		return p.getFabconnectServiceDefinitions(p.stack.Members)
	}
	serviceDefinitions := GenerateDockerServiceDefinitions(p.stack)
	serviceDefinitions = append(serviceDefinitions, p.getFabconnectServiceDefinitions(p.stack.Members)...)
	return serviceDefinitions
}

func (p *FabricProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
	var connectorURL string
	if m.External {
		connectorURL = p.GetConnectorExternalURL(m)
	} else {
		connectorURL = p.GetConnectorURL(m)
	}
	blockchainConfig = &types.BlockchainConfig{
		Type: "fabric",
		Fabric: &types.FabricConfig{
			Fabconnect: &types.FabconnectConfig{
				URL:       connectorURL,
				Chaincode: p.stack.ChaincodeName,
				Channel:   p.stack.ChannelName,
				Signer:    m.OrgName,
				Topic:     m.ID,
			},
		},
	}
	return
}

func (p *FabricProvider) GetOrgConfig(stack *types.Stack, m *types.Organization) (orgConfig *types.OrgConfig) {
	orgConfig = &types.OrgConfig{
		Name: m.OrgName,
		Key:  m.OrgName,
	}
	return
}

func (p *FabricProvider) Reset() error {
	return nil
}

func (p *FabricProvider) getFabconnectServiceDefinitions(members []*types.Organization) []*docker.ServiceDefinition {
	blockchainDirectory := path.Join(p.stack.RuntimeDir, "blockchain")
	serviceDefinitions := make([]*docker.ServiceDefinition, len(members))
	for i, member := range members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "fabconnect_" + member.ID,
			Service: &docker.Service{
				Image:         p.stack.VersionManifest.Fabconnect.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_fabconnect_%s", p.stack.Name, member.ID),
				Command:       "-f /fabconnect/fabconnect.yaml",
				Ports:         []string{fmt.Sprintf("%d:3000", member.ExposedConnectorPort)},
				Volumes: []string{
					fmt.Sprintf("fabconnect_receipts_%s:/fabconnect/receipts", member.ID),
					fmt.Sprintf("fabconnect_events_%s:/fabconnect/events", member.ID),
					fmt.Sprintf("%s:/fabconnect/fabconnect.yaml", path.Join(blockchainDirectory, "fabconnect.yaml")),
				},
				HealthCheck: &docker.HealthCheck{
					Test: []string{"CMD", "wget", "-O", "-", "http://localhost:3000/status"},
				},
				Logging:     docker.StandardLogOptions,
				Environment: p.stack.EnvironmentVars,
			},
			VolumeNames: []string{
				"fabconnect_receipts_" + member.ID,
				"fabconnect_events_" + member.ID,
			},
		}

		if p.stack.RemoteFabricNetwork {
			serviceDefinitions[i].Service.Volumes = append(serviceDefinitions[i].Service.Volumes,
				fmt.Sprintf("%s:/etc/firefly/organizations", path.Join(blockchainDirectory, fmt.Sprintf("%s_msp", member.ID))),
				fmt.Sprintf("%s:/fabconnect/ccp.yaml", path.Join(blockchainDirectory, fmt.Sprintf("%s_ccp.yaml", member.ID))),
			)
		} else {
			serviceDefinitions[i].Service.DependsOn = map[string]map[string]string{
				"fabric_ca":      {"condition": "service_started"},
				"fabric_peer":    {"condition": "service_started"},
				"fabric_orderer": {"condition": "service_started"},
			}
			serviceDefinitions[i].Service.Volumes = append(serviceDefinitions[i].Service.Volumes,
				"firefly_fabric:/etc/firefly",
				fmt.Sprintf("%s:/fabconnect/ccp.yaml", path.Join(blockchainDirectory, "ccp.yaml")),
			)
			serviceDefinitions[i].VolumeNames = append(serviceDefinitions[i].VolumeNames, "firefly_fabric")
		}
	}

	return serviceDefinitions
}

func (p *FabricProvider) writeConfigtxYaml() error {
	if !p.stack.RemoteFabricNetwork {
		filePath := path.Join(p.stack.InitDir, "blockchain", "configtx.yaml")
		return os.WriteFile(filePath, []byte(configtxYaml), 0755)
	}
	return nil
}

func (p *FabricProvider) createChannel() error {
	p.log.Info("creating channel")
	stackDir := p.stack.StackDir
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	return docker.RunDockerCommand(p.ctx, stackDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		FabricToolsImageName,
		"osnadmin", "channel", "join",
		"--channelID", "firefly",
		"--config-block", "/etc/firefly/firefly.block",
		"-o", "fabric_orderer:7053",
		"--ca-file", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/ca.crt",
		"--client-cert", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.crt",
		"--client-key", "/etc/firefly/organizations/ordererOrganizations/example.com/users/Admin@example.com/tls/client.key",
	)
}

func (p *FabricProvider) joinChannel() error {
	p.log.Info("joining channel")
	stackDir := p.stack.StackDir
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	return docker.RunDockerCommand(p.ctx, stackDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		"-e", "CORE_PEER_ADDRESS=fabric_peer:7051",
		"-e", "CORE_PEER_TLS_ENABLED=true",
		"-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
		"-e", "CORE_PEER_LOCALMSPID=Org1MSP",
		"-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		FabricToolsImageName,
		"peer", "channel", "join",
		"-b", "/etc/firefly/firefly.block")
}

func (p *FabricProvider) extractChaincode() error {
	contractsDir := path.Join(p.stack.RuntimeDir, "contracts")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return err
	}

	var containerName string
	for _, member := range p.stack.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_firefly_core_%s", p.stack.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	p.log.Info("extracting smart contracts")
	if err := docker.CopyFromContainer(p.ctx, containerName, "/firefly/contracts/firefly_fabric.tar.gz", path.Join(contractsDir, "firefly_fabric.tar.gz")); err != nil {
		return err
	}
	return nil
}

func (p *FabricProvider) installChaincode(packageFilename string) error {
	p.log.Info("installing chaincode")
	contractsDir := path.Join(p.stack.RuntimeDir, "contracts")
	if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
		if err := os.Mkdir(contractsDir, 0755); err != nil {
			return err
		}
	}
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	return docker.RunDockerCommand(p.ctx, contractsDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-e", "CORE_PEER_ADDRESS=fabric_peer:7051",
		"-e", "CORE_PEER_TLS_ENABLED=true",
		"-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
		"-e", "CORE_PEER_LOCALMSPID=Org1MSP",
		"-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		"-v", fmt.Sprintf("%s:/package.tar.gz", packageFilename),
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		FabricToolsImageName,
		"peer", "lifecycle", "chaincode", "install", "/package.tar.gz",
	)
}

func (p *FabricProvider) queryInstalled() (*QueryInstalledResponse, error) {
	p.log.Info("querying installed chaincode")
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	str, err := docker.RunDockerCommandBuffered(p.ctx, p.stack.RuntimeDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-e", "CORE_PEER_ADDRESS=fabric_peer:7051",
		"-e", "CORE_PEER_TLS_ENABLED=true",
		"-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
		"-e", "CORE_PEER_LOCALMSPID=Org1MSP",
		"-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		FabricToolsImageName,
		"peer", "lifecycle", "chaincode", "queryinstalled",
		"--output", "json",
	)
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

func (p *FabricProvider) approveChaincode(channel, chaincode, version, packageID string) error {
	p.log.Info("approving chaincode")
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	return docker.RunDockerCommand(p.ctx, p.stack.RuntimeDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-e", "CORE_PEER_ADDRESS=fabric_peer:7051",
		"-e", "CORE_PEER_TLS_ENABLED=true",
		"-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
		"-e", "CORE_PEER_LOCALMSPID=Org1MSP",
		"-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		FabricToolsImageName,
		"peer", "lifecycle", "chaincode", "approveformyorg",
		"-o", "fabric_orderer:7050",
		"--ordererTLSHostnameOverride", "fabric_orderer",
		"--channelID", channel,
		"--name", chaincode,
		"--version", version,
		"--package-id", packageID,
		"--sequence", "1",
		"--tls",
		"--cafile", "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem",
	)
}

func (p *FabricProvider) commitChaincode(channel, chaincode, version string) error {
	p.log.Info("committing chaincode")
	volumeName := fmt.Sprintf("%s_firefly_fabric", p.stack.Name)
	return docker.RunDockerCommand(p.ctx, p.stack.RuntimeDir,
		"run",
		"--rm",
		fmt.Sprintf("--network=%s_default", p.stack.Name),
		"-e", "CORE_PEER_ADDRESS=fabric_peer:7051",
		"-e", "CORE_PEER_TLS_ENABLED=true",
		"-e", "CORE_PEER_TLS_ROOTCERT_FILE=/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
		"-e", "CORE_PEER_LOCALMSPID=Org1MSP",
		"-e", "CORE_PEER_MSPCONFIGPATH=/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		"-v", fmt.Sprintf("%s:/etc/firefly", volumeName),
		FabricToolsImageName,
		"peer", "lifecycle", "chaincode", "commit",
		"-o", "fabric_orderer:7050",
		"--ordererTLSHostnameOverride", "fabric_orderer",
		"--channelID", channel,
		"--name", chaincode,
		"--version", version,
		"--sequence", "1",
		"--tls",
		"--cafile", "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem",
	)
}

func (p *FabricProvider) registerIdentity(member *types.Organization, name string) (*Account, error) {
	res, err := fabconnect.CreateIdentity(fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort), name)
	if err != nil {
		return nil, err
	}
	_, err = fabconnect.EnrollIdentity(fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort), name, res.Secret)
	if err != nil {
		return nil, err
	}
	return &Account{
		Name:    name,
		OrgName: member.OrgName,
	}, nil
}

func (p *FabricProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	return []string{filename}, nil
}

func (p *FabricProvider) DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	switch {
	case len(extraArgs) < 1:
		return nil, fmt.Errorf("channel not set")
	case len(extraArgs) < 2:
		return nil, fmt.Errorf("chaincode not set")
	case len(extraArgs) < 3:
		return nil, fmt.Errorf("version not set")
	}
	channel := extraArgs[0]
	chaincode := extraArgs[1]
	version := extraArgs[2]

	if err := p.installChaincode(filename); err != nil {
		if !strings.Contains(err.Error(), "chaincode already successfully installed") {
			return nil, err
		}
	}

	res, err := p.queryInstalled()
	if err != nil {
		return nil, err
	}

	chaincodeInstalled := false
	packageID := ""
	for _, installedChaincode := range res.InstalledChaincodes {
		validLabel := regexp.MustCompile(`^([^_]+)_.+$`)
		matches := validLabel.FindStringSubmatch(installedChaincode.Label)
		for _, match := range matches {
			if match == chaincode {
				chaincodeInstalled = true
				packageID = installedChaincode.PackageID
				break
			}
		}
	}

	if !chaincodeInstalled {
		return nil, fmt.Errorf("failed to find installed chaincode")
	}

	if err := p.approveChaincode(channel, chaincode, version, packageID); err != nil {
		return nil, err
	}

	if err := p.commitChaincode(channel, chaincode, version); err != nil {
		return nil, err
	}
	result := &types.ContractDeploymentResult{
		DeployedContract: &types.DeployedContract{
			Name: contractName,
			Location: map[string]string{
				"channel":   channel,
				"chaincode": chaincode,
			},
		},
	}
	return result, nil
}

func (p *FabricProvider) CreateAccount(args []string) (interface{}, error) {
	stackHasRunBefore, err := p.stack.HasRunBefore()
	if err != nil {
		return nil, err
	}
	switch {
	case len(args) < 1:
		return "", fmt.Errorf("org name not set")
	case len(args) < 2:
		return "", fmt.Errorf("account name not set")
	}
	orgName := args[0]
	accountName := args[1]

	if stackHasRunBefore {
		// Find the FireFly member by the org name
		for _, member := range p.stack.Members {
			if member.OrgName == orgName {
				return p.registerIdentity(member, accountName)
			}
		}
		return nil, fmt.Errorf("unable to find a FireFly org with name: '%s'", orgName)
	}
	return &Account{
		Name:    accountName,
		OrgName: orgName,
	}, nil
}

func (p *FabricProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &Account{
		Name:    accountMap["name"].(string),
		OrgName: accountMap["orgName"].(string),
	}
}

func (p *FabricProvider) GetConnectorName() string {
	return "fabconnect"
}

func (p *FabricProvider) GetConnectorURL(org *types.Organization) string {
	return fmt.Sprintf("http://fabconnect_%s:%v", org.ID, 3000)
}

func (p *FabricProvider) GetConnectorExternalURL(org *types.Organization) string {
	return fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort)
}
