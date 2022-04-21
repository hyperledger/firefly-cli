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

package remoterpc

import (
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethsigner"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type RemoteRPCProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
	Signer  *ethsigner.EthSignerProvider
}

func (p *RemoteRPCProvider) WriteConfig(options *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.Stack.Name, "init")
	for i, member := range p.Stack.Members {

		// Generate the ethconnect config for each member
		ethconnectConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		if err := ethconnect.GenerateEthconnectConfig(member, "ethsigner").WriteConfig(ethconnectConfigPath, options.ExtraEthconnectConfigPath); err != nil {
			return nil
		}

	}

	return p.Signer.WriteConfig(options)
}

func (p *RemoteRPCProvider) FirstTimeSetup() error {
	return p.Signer.FirstTimeSetup()
}

func (p *RemoteRPCProvider) PreStart() error {
	return nil
}

func (p *RemoteRPCProvider) PostStart() error {
	return nil
}

func (p *RemoteRPCProvider) DeployFireFlyContract() (*core.BlockchainConfig, error) {
	return nil, fmt.Errorf("You must pre-deploy your FireFly contract when using a remote RPC endpoint")
}

func (p *RemoteRPCProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	return []*docker.ServiceDefinition{
		p.Signer.GetDockerServiceDefinition(p.Stack.RemoteNodeURL),
	}
}

func (p *RemoteRPCProvider) GetFireflyConfig(stack *types.Stack, m *types.Member) (blockchainConfig *core.BlockchainConfig, orgConfig *core.OrgConfig) {
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

func (p *RemoteRPCProvider) Reset() error {
	return nil
}

func (p *RemoteRPCProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
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

func (p *RemoteRPCProvider) DeployContract(filename, contractName string, member *types.Member, extraArgs []string) (interface{}, error) {
	contractAddres, err := ethconnect.DeployCustomContract(member, filename, contractName)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"address": contractAddres,
	}, nil
}

func (p *RemoteRPCProvider) CreateAccount(args []string) (interface{}, error) {
	return p.Signer.CreateAccount(args)
}

func (p *RemoteRPCProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}

func (p *RemoteRPCProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &ethereum.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}
