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

package geth

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/evmconnect"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var gethImage = "ethereum/client-go:release-1.10"

// TODO: Probably randomize this and make it different per member?
var keyPassword = "correcthorsebatterystaple"

type GethProvider struct {
	ctx       context.Context
	stack     *types.Stack
	connector connector.Connector
}

func NewGethProvider(ctx context.Context, stack *types.Stack) *GethProvider {
	var connector connector.Connector
	switch stack.BlockchainConnector {
	case types.BlockchainConnectorEthconnect:
		connector = ethconnect.NewEthconnect(ctx)
	case types.BlockchainConnectorEvmconnect:
		connector = evmconnect.NewEvmconnect(ctx)
	}

	return &GethProvider{
		ctx:       ctx,
		stack:     stack,
		connector: connector,
	}
}

func (p *GethProvider) WriteConfig(options *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	for i, member := range p.stack.Members {
		// Generate the connector config for each member
		connectorConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("%s_%v.yaml", p.connector.Name(), i))
		if err := p.connector.GenerateConfig(p.stack, member, "geth").WriteConfig(connectorConfigPath, options.ExtraConnectorConfigPath); err != nil {
			return nil
		}
	}

	// Create genesis.json
	addresses := make([]string, len(p.stack.Members))
	for i, member := range p.stack.Members {
		address := member.Account.(*ethereum.Account).Address
		// Drop the 0x on the front of the address here because that's what geth is expecting in the genesis.json
		addresses[i] = address[2:]
	}
	genesis := CreateGenesis(addresses, options.BlockPeriod, p.stack.ChainID())
	if err := genesis.WriteGenesisJSON(filepath.Join(initDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	return nil
}

func (p *GethProvider) FirstTimeSetup() error {
	gethVolumeName := fmt.Sprintf("%s_geth", p.stack.Name)
	blockchainDir := path.Join(p.stack.RuntimeDir, "blockchain")
	contractsDir := path.Join(p.stack.RuntimeDir, "contracts")

	if err := p.connector.FirstTimeSetup(p.stack); err != nil {
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

	// Copy the wallet files all members to the blockchain volume
	keystoreDirectory := filepath.Join(blockchainDir, "keystore")
	if err := docker.CopyFileToVolume(p.ctx, gethVolumeName, keystoreDirectory, "/"); err != nil {
		return err
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(p.ctx, gethVolumeName, path.Join(blockchainDir, "genesis.json"), "genesis.json"); err != nil {
		return err
	}

	// Initialize the genesis block
	if err := docker.RunDockerCommand(p.ctx, p.stack.StackDir, "run", "--rm", "-v", fmt.Sprintf("%s:/data", gethVolumeName), gethImage, "--datadir", "/data", "init", "/data/genesis.json"); err != nil {
		return err
	}

	return nil
}

func (p *GethProvider) PreStart() error {
	return nil
}

func (p *GethProvider) PostStart(firstTimeSetup bool) error {
	l := log.LoggerFromContext(p.ctx)
	// Unlock accounts
	for _, account := range p.stack.State.Accounts {
		address := account.(*ethereum.Account).Address
		l.Info(fmt.Sprintf("unlocking account %s", address))
		if err := p.unlockAccount(address, keyPassword); err != nil {
			return err
		}
	}

	return nil
}

func (p *GethProvider) unlockAccount(address, password string) error {
	l := log.LoggerFromContext(p.ctx)
	verbose := log.VerbosityFromContext(p.ctx)
	gethClient := NewGethClient(fmt.Sprintf("http://127.0.0.1:%v", p.stack.ExposedBlockchainPort))
	retries := 10
	for {
		if err := gethClient.UnlockAccount(address, password); err != nil {
			if verbose {
				l.Debug(err.Error())
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

func (p *GethProvider) DeployFireFlyContract() (*types.ContractDeploymentResult, error) {
	contract, err := ethereum.ReadFireFlyContract(p.ctx, p.stack)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contract, "FireFly", p.stack.Members[0], nil)
}

func (p *GethProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	gethCommand := fmt.Sprintf(`--datadir /data --syncmode 'full' --port 30311 --http --http.addr "0.0.0.0" --http.corsdomain="*"  -http.port 8545 --http.vhosts "*" --http.api 'admin,personal,eth,net,web3,txpool,miner,clique,debug' --networkid %d --miner.gasprice 0 --password /data/password --mine --allow-insecure-unlock --nodiscover --verbosity 4 --miner.gaslimit 16777215`, p.stack.ChainID())

	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "geth",
		Service: &docker.Service{
			Image:         gethImage,
			ContainerName: fmt.Sprintf("%s_geth", p.stack.Name),
			Command:       gethCommand,
			Volumes:       []string{"geth:/data"},
			Logging:       docker.StandardLogOptions,
			Ports:         []string{fmt.Sprintf("%d:8545", p.stack.ExposedBlockchainPort)},
			Environment:   p.stack.EnvironmentVars,
		},
		VolumeNames: []string{"geth"},
	}
	serviceDefinitions = append(serviceDefinitions, p.connector.GetServiceDefinitions(p.stack, map[string]string{"geth": "service_started"})...)
	return serviceDefinitions
}

func (p *GethProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
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

func (p *GethProvider) GetOrgConfig(stack *types.Stack, m *types.Organization) (orgConfig *types.OrgConfig) {
	account := m.Account.(*ethereum.Account)
	orgConfig = &types.OrgConfig{
		Name: m.OrgName,
		Key:  account.Address,
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

func (p *GethProvider) DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	contracts, err := ethereum.ReadContractJSON(filename)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contracts.Contracts[contractName], instanceName, member, extraArgs)
}

func (p *GethProvider) CreateAccount(args []string) (interface{}, error) {
	gethVolumeName := fmt.Sprintf("%s_geth", p.stack.Name)
	var directory string
	stackHasRunBefore, err := p.stack.HasRunBefore()
	if err != nil {
		return nil, err
	}
	if stackHasRunBefore {
		directory = p.stack.RuntimeDir
	} else {
		directory = p.stack.InitDir
	}

	prefix := strconv.FormatInt(time.Now().UnixNano(), 10)
	outputDirectory := filepath.Join(directory, "blockchain", "keystore")
	keyPair, walletFilePath, err := ethereum.CreateWalletFile(outputDirectory, prefix, keyPassword)
	if err != nil {
		return nil, err
	}

	if stackHasRunBefore {
		if err := ethereum.CopyWalletFileToVolume(p.ctx, walletFilePath, gethVolumeName); err != nil {
			return nil, err
		}
		if err := p.unlockAccount(keyPair.Address.String(), keyPassword); err != nil {
			return nil, err
		}
	}

	return &ethereum.Account{
		Address:    keyPair.Address.String(),
		PrivateKey: hex.EncodeToString(keyPair.PrivateKeyBytes()),
	}, nil
}

func (p *GethProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &ethereum.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}

func (p *GethProvider) GetConnectorName() string {
	return p.connector.Name()
}

func (p *GethProvider) GetConnectorURL(org *types.Organization) string {
	return fmt.Sprintf("http://%s_%s:%v", p.connector.Name(), org.ID, p.connector.Port())
}

func (p *GethProvider) GetConnectorExternalURL(org *types.Organization) string {
	return fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort)
}
