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
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var besuImage = "hyperledger/besu:22.4"
var ethsignerImage = "consensys/ethsigner:22.1"
var gethImage = "ethereum/client-go:release-1.10"

type BesuProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *BesuProvider) WriteConfig(options *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.Stack.Name, "init")
	for i, member := range p.Stack.Members {
		// Write the private key to disk for each member
		if err := p.writeAccountToDisk(p.Stack.InitDir, member.Address, member.PrivateKey); err != nil {
			return err
		}

		// Generate the ethconnect config for each member
		ethconnectConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		if err := ethconnect.GenerateEthconnectConfig(member, "ethsigner").WriteConfig(ethconnectConfigPath, options.ExtraEthconnectConfigPath); err != nil {
			return nil
		}

		if err := p.writeTomlKeyFile(p.Stack.InitDir, member.Address); err != nil {
			return err
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
	genesis := CreateGenesis([]string{nodeAddress[2:]}, options.BlockPeriod)
	if err := genesis.WriteGenesisJson(filepath.Join(initDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	// Write the password that will be used to encrypt the private key
	// TODO: Probably randomize this and make it differnet per member?
	if err := ioutil.WriteFile(filepath.Join(initDir, "blockchain", "password"), []byte("correcthorsebatterystaple"), 0755); err != nil {
		return err
	}

	return nil
}

func (p *BesuProvider) FirstTimeSetup() error {
	ethsignerVolumeName := fmt.Sprintf("%s_ethsigner", p.Stack.Name)
	besuVolumeName := fmt.Sprintf("%s_besu", p.Stack.Name)
	blockchainDir := filepath.Join(p.Stack.RuntimeDir, "blockchain")
	contractsDir := filepath.Join(p.Stack.RuntimeDir, "contracts")

	if err := docker.CreateVolume(ethsignerVolumeName, p.Verbose); err != nil {
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

	// Mount the directory containing all members' private keys and password, and import the accounts using the geth CLI
	// Note: This is needed because of licensing issues with the Go Ethereum library that could do this step
	for _, member := range p.Stack.Members {
		if err := p.importAccountToEthsigner(member.Address); err != nil {
			return err
		}
	}

	// Copy the password (to be used for decrypting private keys)
	if err := docker.CopyFileToVolume(ethsignerVolumeName, path.Join(blockchainDir, "password"), "password", p.Verbose); err != nil {
		return err
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

func (p *BesuProvider) PostStart() error {
	return nil
}

func (p *BesuProvider) DeployFireFlyContract() (*core.BlockchainConfig, error) {
	return ethconnect.DeployFireFlyContract(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	addresses := ""
	for i, member := range p.Stack.Members {
		addresses = addresses + member.Address
		if i+1 < len(p.Stack.Members) {
			addresses = addresses + ","
		}
	}
	besuCommand := `--genesis-file=/data/genesis.json --network-id 2021 --rpc-http-enabled --rpc-http-api=ETH,NET,CLIQUE --host-allowlist="*" --rpc-http-cors-origins="all" --sync-mode=FULL --discovery-enabled=false --node-private-key-file=/data/nodeKey --min-gas-price=0`
	ethsignerCommand := `--chain-id=2021 --downstream-http-host="besu" --downstream-http-port=8545 multikey-signer --directory=/data/keystore`

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
	serviceDefinitions[1] = &docker.ServiceDefinition{
		ServiceName: "ethsigner",
		Service: &docker.Service{
			Image:         ethsignerImage,
			ContainerName: fmt.Sprintf("%s_ethsigner", p.Stack.Name),
			User:          "root",
			Command:       ethsignerCommand,
			Volumes:       []string{"ethsigner:/data"},
			Logging:       docker.StandardLogOptions,
			HealthCheck: &docker.HealthCheck{
				Test:     []string{"CMD", "curl", "http://besu:8545/livenes"},
				Interval: "4s",
				Retries:  30,
			},
			Ports: []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"ethsigner"},
	}
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack, "ethsigner")...)
	return serviceDefinitions
}

func (p *BesuProvider) GetFireflyConfig(stack *types.Stack, m *types.Member) (blockchainConfig *core.BlockchainConfig, orgConfig *core.OrgConfig) {
	orgConfig = &core.OrgConfig{
		Name: m.OrgName,
		Key:  m.Address,
	}

	blockchainConfig = &core.BlockchainConfig{
		Type: "ethereum",
		Ethereum: &core.EthereumConfig{
			Ethconnect: &core.EthconnectConfig{
				URL:      p.getEthconnectURL(m),
				Instance: stack.ContractAddress,
				Topic:    m.ID,
			},
		},
	}
	return
}

func (p *BesuProvider) getSmartContractAddressPatchJSON(contractAddress string) []byte {
	return []byte(fmt.Sprintf(`{"blockchain":{"ethereum":{"ethconnect":{"instance":"%s"}}}}`, contractAddress))
}

func (p *BesuProvider) Reset() error {
	return nil
}

func (p *BesuProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	contracts, err := ethereum.ReadCombinedABIJSON(filename)
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

func (p *BesuProvider) DeployContract(filename, contractName string, member *types.Member, extraArgs []string) (interface{}, error) {
	contractAddres, err := ethconnect.DeployCustomContract(member, filename, contractName)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"address": contractAddres,
	}, nil
}

func (p *BesuProvider) CreateAccount() (interface{}, error) {
	address, privateKey := ethereum.GenerateAddressAndPrivateKey()

	if err := p.writeAccountToDisk(p.Stack.RuntimeDir, address, privateKey); err != nil {
		return nil, err
	}

	if err := p.writeTomlKeyFile(p.Stack.RuntimeDir, address); err != nil {
		return nil, err
	}

	if err := p.importAccountToEthsigner(address); err != nil {
		return nil, err
	}

	return map[string]string{
		"address":    address,
		"privateKey": privateKey,
	}, nil
}

func (p *BesuProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}
