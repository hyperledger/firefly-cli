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

package besu

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/evmconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethsigner"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var besuImage = "hyperledger/besu:22.4"

type BesuProvider struct {
	ctx       context.Context
	stack     *types.Stack
	signer    *ethsigner.EthSignerProvider
	connector connector.Connector
}

func NewBesuProvider(ctx context.Context, stack *types.Stack) *BesuProvider {
	var connector connector.Connector
	switch stack.BlockchainConnector {
	case types.BlockchainConnectorEthconnect:
		connector = ethconnect.NewEthconnect(ctx)
	case types.BlockchainConnectorEvmconnect:
		connector = evmconnect.NewEvmconnect(ctx)
	}

	return &BesuProvider{
		ctx:       ctx,
		stack:     stack,
		connector: connector,
		signer:    ethsigner.NewEthSignerProvider(ctx, stack),
	}
}

func (p *BesuProvider) WriteConfig(options *types.InitOptions) error {
	if err := p.signer.WriteConfig(options, "http://besu:8545"); err != nil {
		return err
	}

	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	for i, member := range p.stack.Members {

		// Generate the connector config for each member
		connectorConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("%s_%v.yaml", p.connector.Name(), i))
		if err := p.connector.GenerateConfig(p.stack, member, "ethsigner").WriteConfig(connectorConfigPath, options.ExtraConnectorConfigPath); err != nil {
			return nil
		}

	}

	// Create genesis.json
	// Generate node key
	nodeAddress, nodeKey := ethereum.GenerateAddressAndPrivateKey()
	// Write the node key to disk
	if err := os.WriteFile(filepath.Join(initDir, "blockchain", "nodeKey"), []byte(nodeKey), 0755); err != nil {
		return err
	}
	// Drop the 0x on the front of the address here because that's what is expected in the genesis.json
	genesis := CreateGenesis([]string{nodeAddress[2:]}, options.BlockPeriod, p.stack.ChainID())
	if err := genesis.WriteGenesisJSON(filepath.Join(initDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	return nil
}

func (p *BesuProvider) FirstTimeSetup() error {
	besuVolumeName := fmt.Sprintf("%s_besu", p.stack.Name)
	blockchainDir := filepath.Join(p.stack.RuntimeDir, "blockchain")
	contractsDir := filepath.Join(p.stack.RuntimeDir, "contracts")

	if err := p.signer.FirstTimeSetup(); err != nil {
		return err
	}

	if err := p.connector.FirstTimeSetup(p.stack); err != nil {
		return err
	}

	if err := docker.CreateVolume(p.ctx, besuVolumeName); err != nil {
		return err
	}

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return err
	}

	for i := range p.stack.Members {
		// Copy connector config to each member's volume
		connectorConfigPath := filepath.Join(p.stack.StackDir, "runtime", "config", fmt.Sprintf("%s_%v.yaml", p.connector.Name(), i))
		connectorConfigVolumeName := fmt.Sprintf("%s_%s_config_%v", p.stack.Name, p.connector.Name(), i)
		if err := docker.CopyFileToVolume(p.ctx, connectorConfigVolumeName, connectorConfigPath, "config.yaml"); err != nil {
			return err
		}
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(p.ctx, besuVolumeName, path.Join(blockchainDir, "genesis.json"), "genesis.json"); err != nil {
		return err
	}

	// Copy the node key
	if err := docker.CopyFileToVolume(p.ctx, besuVolumeName, path.Join(blockchainDir, "nodeKey"), "nodeKey"); err != nil {
		return err
	}

	return nil
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart(firstTimeSetup bool) error {
	return nil
}

func (p *BesuProvider) DeployFireFlyContract() (*types.ContractDeploymentResult, error) {
	contract, err := ethereum.ReadFireFlyContract(p.ctx, p.stack)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contract, "FireFly", p.stack.Members[0], nil)
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	addresses := ""
	for i, member := range p.stack.Members {
		account := member.Account.(*ethereum.Account)
		addresses += account.Address
		if i+1 < len(p.stack.Members) {
			addresses += ","
		}
	}
	besuCommand := fmt.Sprintf(`--genesis-file=/data/genesis.json --network-id %d --rpc-http-enabled --rpc-http-api=ETH,NET,CLIQUE --host-allowlist="*" --rpc-http-cors-origins="all" --sync-mode=FULL --discovery-enabled=false --node-private-key-file=/data/nodeKey --min-gas-price=0`, p.stack.ChainID())

	serviceDefinitions := make([]*docker.ServiceDefinition, 2)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "besu",
		Service: &docker.Service{
			Image:         besuImage,
			ContainerName: fmt.Sprintf("%s_besu", p.stack.Name),
			User:          "root",
			Command:       besuCommand,
			Volumes: []string{
				"besu:/data",
			},
			Logging:     docker.StandardLogOptions,
			Environment: p.stack.EnvironmentVars,
		},

		VolumeNames: []string{"besu"},
	}
	serviceDefinitions[1] = p.signer.GetDockerServiceDefinition("http://besu:8545")
	serviceDefinitions = append(serviceDefinitions, p.connector.GetServiceDefinitions(p.stack, map[string]string{"ethsigner": "service_healthy"})...)
	return serviceDefinitions
}

func (p *BesuProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
	var connectorURL string
	if m.External {
		connectorURL = p.GetConnectorExternalURL(m)
	} else {
		connectorURL = p.GetConnectorURL(m)
	}

	blockchainConfig = &types.BlockchainConfig{
		Type: "ethereum",
		Ethereum: &types.EthereumConfig{
			Ethconnect: &types.EthconnectConfig{
				URL:   connectorURL,
				Topic: m.ID,
			},
		},
	}
	return
}

func (p *BesuProvider) GetOrgConfig(stack *types.Stack, m *types.Organization) (orgConfig *types.OrgConfig) {
	account := m.Account.(*ethereum.Account)
	orgConfig = &types.OrgConfig{
		Name: m.OrgName,
		Key:  account.Address,
	}
	return
}

func (p *BesuProvider) Reset() error {
	return nil
}

func (p *BesuProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	contracts, err := ethereum.ReadContractJSON(filename)
	if err != nil {
		return []string{}, err
	}
	contractNames := make([]string, len(contracts.Contracts))
	i := 0
	for contractName := range contracts.Contracts {
		contractNames[i] = contractName
		i++
	}
	return contractNames, err
}

func (p *BesuProvider) DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	contracts, err := ethereum.ReadContractJSON(filename)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contracts.Contracts[contractName], instanceName, member, extraArgs)
}

func (p *BesuProvider) CreateAccount(args []string) (interface{}, error) {
	return p.signer.CreateAccount(args)
}

func (p *BesuProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &ethereum.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}

func (p *BesuProvider) GetConnectorName() string {
	return p.connector.Name()
}

func (p *BesuProvider) GetConnectorURL(org *types.Organization) string {
	return fmt.Sprintf("http://%s_%s:%v", p.connector.Name(), org.ID, p.connector.Port())
}

func (p *BesuProvider) GetConnectorExternalURL(org *types.Organization) string {
	return fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort)
}
