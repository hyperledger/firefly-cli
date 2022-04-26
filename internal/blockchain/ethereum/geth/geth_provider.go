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

package geth

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var gethImage = "ethereum/client-go:release-1.10"

type GethProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *GethProvider) WriteConfig(options *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.Stack.Name, "init")
	for i, member := range p.Stack.Members {
		account := member.Account.(*ethereum.Account)
		// Write the private key to disk for each member
		// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
		if err := p.writeAccountToDisk(p.Stack.InitDir, account.Address, account.PrivateKey); err != nil {
			return err
		}

		// Generate the ethconnect config for each member
		ethconnectConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		if err := ethconnect.GenerateEthconnectConfig(member, "geth").WriteConfig(ethconnectConfigPath, options.ExtraEthconnectConfigPath); err != nil {
			return nil
		}
	}

	// Create genesis.json
	addresses := make([]string, len(p.Stack.Members))
	for i, member := range p.Stack.Members {
		address := member.Account.(*ethereum.Account).Address
		// Drop the 0x on the front of the address here because that's what geth is expecting in the genesis.json
		addresses[i] = address[2:]
	}
	genesis := CreateGenesis(addresses, options.BlockPeriod)
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

func (p *GethProvider) FirstTimeSetup() error {
	gethVolumeName := fmt.Sprintf("%s_geth", p.Stack.Name)
	blockchainDir := path.Join(p.Stack.RuntimeDir, "blockchain")
	contractsDir := path.Join(p.Stack.RuntimeDir, "contracts")

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
	for _, member := range p.Stack.Members {
		address := member.Account.(*ethereum.Account).Address
		p.importAccountToGeth(address)
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(gethVolumeName, path.Join(blockchainDir, "genesis.json"), "genesis.json", p.Verbose); err != nil {
		return err
	}

	// Copy the password (to be used for decrypting private keys)
	if err := docker.CopyFileToVolume(gethVolumeName, path.Join(blockchainDir, "password"), "password", p.Verbose); err != nil {
		return err
	}

	// Initialize the genesis block
	if err := docker.RunDockerCommand(p.Stack.StackDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/data", gethVolumeName), gethImage, "--datadir", "/data", "init", "/data/genesis.json"); err != nil {
		return err
	}

	return nil
}

func (p *GethProvider) PreStart() error {
	return nil
}

func (p *GethProvider) PostStart() error {
	// Unlock accounts
	for _, account := range p.Stack.State.Accounts {
		address := account.(*ethereum.Account).Address
		p.Log.Info(fmt.Sprintf("unlocking account %s", address))
		if err := p.unlockAccount(address, "correcthorsebatterystaple"); err != nil {
			return err
		}
	}

	return nil
}

func (p *GethProvider) unlockAccount(address, password string) error {
	gethClient := NewGethClient(fmt.Sprintf("http://127.0.0.1:%v", p.Stack.ExposedBlockchainPort))
	retries := 10
	for {
		if err := gethClient.UnlockAccount(address, password); err != nil {
			if p.Verbose {
				p.Log.Debug(err.Error())
			}
			if retries == 0 {
				return fmt.Errorf("unable to unlock account %s", address)
			}
			time.Sleep(time.Second * 1)
			retries--
		} else {
			break
		}
	}
	return nil
}

func (p *GethProvider) DeployFireFlyContract() (*core.BlockchainConfig, *types.ContractDeploymentResult, error) {
	return ethconnect.DeployFireFlyContract(p.Stack, p.Log, p.Verbose)
}

func (p *GethProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	gethCommand := fmt.Sprintf(`--datadir /data --syncmode 'full' --port 30311 --http --http.addr "0.0.0.0" --http.port 8545 --http.vhosts "*" --http.api 'admin,personal,eth,net,web3,txpool,miner,clique' --networkid 2021 --miner.gasprice 0 --password /data/password --mine --allow-insecure-unlock --nodiscover`)

	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "geth",
		Service: &docker.Service{
			Image:         gethImage,
			ContainerName: fmt.Sprintf("%s_geth", p.Stack.Name),
			Command:       gethCommand,
			Volumes:       []string{"geth:/data"},
			Logging:       docker.StandardLogOptions,
			Ports:         []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"geth"},
	}
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack, "geth")...)
	return serviceDefinitions
}

func (p *GethProvider) GetFireflyConfig(stack *types.Stack, m *types.Member) (blockchainConfig *core.BlockchainConfig, orgConfig *core.OrgConfig) {
	account := m.Account.(*ethereum.Account)
	orgConfig = &core.OrgConfig{
		Name: m.OrgName,
		Key:  account.Address,
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

func (p *GethProvider) Reset() error {
	return nil
}

func (p *GethProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
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

func (p *GethProvider) DeployContract(filename, contractName string, member *types.Member, extraArgs []string) (*types.ContractDeploymentResult, error) {
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

func (p *GethProvider) CreateAccount(args []string) (interface{}, error) {
	address, privateKey := ethereum.GenerateAddressAndPrivateKey()

	if err := p.writeAccountToDisk(p.Stack.RuntimeDir, address, privateKey); err != nil {
		return nil, err
	}
	if err := p.importAccountToGeth(address); err != nil {
		return nil, err
	}
	if err := p.unlockAccount(address, "correcthorsebatterystaple"); err != nil {
		return nil, err
	}

	return map[string]string{
		"address":    address,
		"privateKey": privateKey,
	}, nil
}

func (p *GethProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}

func (p *GethProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &ethereum.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}
