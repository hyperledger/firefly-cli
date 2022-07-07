// Copyright © 2022 Kaleido, Inc.
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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethsigner"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var besuImage = "hyperledger/besu:22.4"
var gethImage = "ethereum/client-go:release-1.10"

type BesuProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
	Signer  *ethsigner.EthSignerProvider
}

func (p *BesuProvider) WriteConfig(options *types.InitOptions) error {
	if err := p.Signer.WriteConfig(options, "http://besu:8545"); err != nil {
		return err
	}

	initDir := filepath.Join(constants.StacksDir, p.Stack.Name, "init")
	for i, member := range p.Stack.Members {

		// Generate the ethconnect config for each member
		ethconnectConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		if err := ethconnect.GenerateEthconnectConfig(member, "ethsigner").WriteConfig(ethconnectConfigPath, options.ExtraEthconnectConfigPath); err != nil {
			return nil
		}

	}

	// Create genesis.json
	// Generate node key
	nodeAddress, nodeKey := ethereum.GenerateAddressAndPrivateKey()
	// Write the node key to disk
	if err := ioutil.WriteFile(filepath.Join(initDir, "blockchain", "nodeKey"), []byte(nodeKey), 0755); err != nil {
		return err
	}
	// Drop the 0x on the front of the address here because that's what is expected in the genesis.json
	genesis := CreateGenesis([]string{nodeAddress[2:]}, options.BlockPeriod, p.Stack.ChainID())
	if err := genesis.WriteGenesisJson(filepath.Join(initDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	return nil
}

func (p *BesuProvider) FirstTimeSetup() error {
	besuVolumeName := fmt.Sprintf("%s_besu", p.Stack.Name)
	blockchainDir := filepath.Join(p.Stack.RuntimeDir, "blockchain")
	contractsDir := filepath.Join(p.Stack.RuntimeDir, "contracts")

	if err := p.Signer.FirstTimeSetup(); err != nil {
		return err
	}

	if err := docker.CreateVolume(besuVolumeName, p.Verbose); err != nil {
		return err
	}

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return err
	}

	for i := range p.Stack.Members {
		// Copy ethconnect config to each member's volume
		ethconnectConfigPath := filepath.Join(p.Stack.StackDir, "runtime", "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		ethconnectConfigVolumeName := fmt.Sprintf("%s_ethconnect_config_%v", p.Stack.Name, i)
		docker.CopyFileToVolume(ethconnectConfigVolumeName, ethconnectConfigPath, "config.yaml", p.Verbose)
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(besuVolumeName, path.Join(blockchainDir, "genesis.json"), "genesis.json", p.Verbose); err != nil {
		return err
	}

	// Copy the node key
	if err := docker.CopyFileToVolume(besuVolumeName, path.Join(blockchainDir, "nodeKey"), "nodeKey", p.Verbose); err != nil {
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
	return ethconnect.DeployFireFlyContract(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	addresses := ""
	for i, member := range p.Stack.Members {
		account := member.Account.(*ethereum.Account)
		addresses = addresses + account.Address
		if i+1 < len(p.Stack.Members) {
			addresses = addresses + ","
		}
	}
	besuCommand := fmt.Sprintf(`--genesis-file=/data/genesis.json --network-id %d --rpc-http-enabled --rpc-http-api=ETH,NET,CLIQUE --host-allowlist="*" --rpc-http-cors-origins="all" --sync-mode=FULL --discovery-enabled=false --node-private-key-file=/data/nodeKey --min-gas-price=0`, p.Stack.ChainID())

	serviceDefinitions := make([]*docker.ServiceDefinition, 2)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "besu",
		Service: &docker.Service{
			Image:         besuImage,
			ContainerName: fmt.Sprintf("%s_besu", p.Stack.Name),
			User:          "root",
			Command:       besuCommand,
			Volumes: []string{
				"besu:/data",
			},
			Logging: docker.StandardLogOptions,
		},

		VolumeNames: []string{"besu"},
	}
	serviceDefinitions[1] = p.Signer.GetDockerServiceDefinition("http://besu:8545")
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack, map[string]string{"ethsigner": "service_healthy"})...)
	return serviceDefinitions
}

func (p *BesuProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
	blockchainConfig = &types.BlockchainConfig{
		Type: "ethereum",
		Ethereum: &types.EthereumConfig{
			Ethconnect: &types.EthconnectConfig{
				URL:   p.getEthconnectURL(m),
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

func (p *BesuProvider) DeployContract(filename, contractName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	contractAddres, err := ethconnect.DeployCustomContract(member, filename, contractName)
	if err != nil {
		return nil, err
	}
	result := &types.ContractDeploymentResult{
		DeployedContract: &types.DeployedContract{
			Name: contractName,
			Location: map[string]string{
				"address": contractAddres,
			},
		},
	}
	return result, nil
}

func (p *BesuProvider) CreateAccount(args []string) (interface{}, error) {
	return p.Signer.CreateAccount(args)
}

func (p *BesuProvider) getEthconnectURL(member *types.Organization) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}

func (p *BesuProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &ethereum.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}
