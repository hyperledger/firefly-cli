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
	if err := p.Signer.FirstTimeSetup(); err != nil {

	}

	for i := range p.Stack.Members {
		// Copy ethconnect config to each member's volume
		ethconnectConfigPath := filepath.Join(p.Stack.StackDir, "runtime", "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		ethconnectConfigVolumeName := fmt.Sprintf("%s_ethconnect_config_%v", p.Stack.Name, i)
		docker.CopyFileToVolume(ethconnectConfigVolumeName, ethconnectConfigPath, "config.yaml", p.Verbose)
	}

	return nil
}

func (p *RemoteRPCProvider) PreStart() error {
	return nil
}

func (p *RemoteRPCProvider) PostStart() error {
	return nil
}

func (p *RemoteRPCProvider) DeployFireFlyContract() (*core.BlockchainConfig, *types.ContractDeploymentResult, error) {
	return nil, nil, fmt.Errorf("You must pre-deploy your FireFly contract when using a remote RPC endpoint")
}

func (p *RemoteRPCProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	defs := []*docker.ServiceDefinition{
		p.Signer.GetDockerServiceDefinition(p.Stack.RemoteNodeURL),
	}
	defs = append(defs, ethconnect.GetEthconnectServiceDefinitions(p.Stack, map[string]string{"ethsigner": "service_healthy"})...)
	return defs
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
	if stack.FFTMEnabled {
		blockchainConfig.Ethereum.FFTM = &core.FFTMConfig{
			URL: fmt.Sprintf("http://fftm_%s:5008", m.ID),
		}
	}
	return
}

func (p *RemoteRPCProvider) Reset() error {
	return nil
}

func (p *RemoteRPCProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	return []string{}, nil
}

func (p *RemoteRPCProvider) DeployContract(filename, contractName string, member *types.Member, extraArgs []string) (*types.ContractDeploymentResult, error) {
	return nil, fmt.Errorf("Contract deployment not supported for Remote RPC URL connections")
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
