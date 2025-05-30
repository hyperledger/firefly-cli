// Copyright © 2025 IOG Singapore and SundaeSwap, Inc.
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
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano"
	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano/cardanosigner"
	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano/connector"
	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano/connector/cardanoconnect"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type RemoteRPCProvider struct {
	ctx       context.Context
	stack     *types.Stack
	connector connector.Connector
	signer    *cardanosigner.CardanoSignerProvider
}

func NewRemoteRPCProvider(ctx context.Context, stack *types.Stack) *RemoteRPCProvider {
	return &RemoteRPCProvider{
		ctx:       ctx,
		stack:     stack,
		connector: cardanoconnect.NewCardanoconnect(ctx),
		signer:    cardanosigner.NewCardanoSignerProvider(ctx, stack),
	}
}

func (p *RemoteRPCProvider) WriteConfig(options *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	for i, member := range p.stack.Members {
		// Generate the connector config for each member
		connectorConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("%s_%v.yaml", p.connector.Name(), i))
		if err := p.connector.GenerateConfig(p.stack, member).WriteConfig(connectorConfigPath, options.ExtraConnectorConfigPath); err != nil {
			return err
		}
	}

	return p.signer.WriteConfig(options)
}

func (p *RemoteRPCProvider) FirstTimeSetup() error {
	if err := p.signer.FirstTimeSetup(); err != nil {
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
	return nil
}

func (p *RemoteRPCProvider) DeployFireFlyContract() (*types.ContractDeploymentResult, error) {
	return nil, errors.New("DeployFireFlyContract not supported")
}

func (p *RemoteRPCProvider) PreStart() error {
	return nil
}

func (p *RemoteRPCProvider) PostStart(firstTimeSetup bool) error {
	return nil
}

func (p *RemoteRPCProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	defs := []*docker.ServiceDefinition{
		p.signer.GetDockerServiceDefinition(p.stack.RemoteNodeURL),
	}
	defs = append(defs, p.connector.GetServiceDefinitions(p.stack, map[string]string{})...)

	return defs
}

func (p *RemoteRPCProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
	var connectorURL string
	if m.External {
		connectorURL = p.GetConnectorExternalURL(m)
	} else {
		connectorURL = p.GetConnectorURL(m)
	}
	blockchainConfig = &types.BlockchainConfig{
		Type: "cardano",
		Cardano: &types.CardanoConfig{
			Cardanoconnect: &types.CardanoconnectConfig{
				URL:   connectorURL,
				Topic: m.ID,
			},
		},
	}
	return
}

func (p *RemoteRPCProvider) GetOrgConfig(stack *types.Stack, m *types.Organization) (orgConfig *types.OrgConfig) {
	account := m.Account.(*cardano.Account)
	orgConfig = &types.OrgConfig{
		Name: m.OrgName,
		Key:  account.Address,
	}
	return
}

func (p *RemoteRPCProvider) Reset() error {
	return nil
}

func (p *RemoteRPCProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	return nil, errors.New("GetContracts not supported")
}

func (p *RemoteRPCProvider) DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	return nil, errors.New("DeployContract not supported")
}

func (p *RemoteRPCProvider) CreateAccount(args []string) (interface{}, error) {
	return p.signer.CreateAccount(args)
}

func (p *RemoteRPCProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	return &cardano.Account{
		Address:    accountMap["address"].(string),
		PrivateKey: accountMap["privateKey"].(string),
	}
}

func (p *RemoteRPCProvider) GetConnectorName() string {
	return p.connector.Name()
}

func (p *RemoteRPCProvider) GetConnectorURL(org *types.Organization) string {
	return fmt.Sprintf("http://cardanoconnect_%s:%v", org.ID, 3000)
}

func (p *RemoteRPCProvider) GetConnectorExternalURL(org *types.Organization) string {
	return fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort)
}
