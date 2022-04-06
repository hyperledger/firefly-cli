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

package besu

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type BesuProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

func (p *BesuProvider) WriteConfig(options *types.InitOptions) error {
	GetPath := func(elem ...string) string {
		return filepath.Join(append([]string{p.Stack.InitDir, "config"}, elem...)...)
	}

	if err := p.writeStaticFiles(); err != nil {
		return err
	}

	// try to make a simplified version by reusing code snippets if possible
	file_names := []string{"accounts", "SignerConfig"}
	for _, file := range file_names {
		if err := os.Mkdir(GetPath("ethsigner", file), 0755); err != nil {
			return err
		}
	}

	addresses := make([]string, len(p.Stack.Members))
	if err := os.Mkdir(GetPath("ethsigner", "PassFile"), 0755); err != nil {
		return err
	}
	for i, member := range p.Stack.Members {
		addresses[i] = member.Address[2:]
		// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
		if err := os.Mkdir(GetPath("ethsigner", "accounts", member.ID), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("ethsigner", "accounts", member.ID, "privateKey"), []byte(member.PrivateKey), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("ethsigner", "accounts", member.ID, "address"), []byte(member.Address[2:]), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("ethsigner", "SignerConfig", fmt.Sprintf("%s.toml", member.Address[2:])), []byte(fmt.Sprintf(
			`[metadata]
description = "File based configuration"
[signing]
type = "file-based-signer"
key-file = "%s_keyFile"
password-file = "%s"`, filepath.Join("/keyFiles", member.ID), filepath.Join("/PassFile", "passwordFile"))), 0755); err != nil {
			return err
		}

		// Generate the ethconnect config for each member
		ethconnectConfigPath := filepath.Join(p.Stack.InitDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		if err := ethconnect.GenerateEthconnectConfig(member, "ethsigner").WriteConfig(ethconnectConfigPath, options.ExtraEthconnectConfigPath); err != nil {
			return nil
		}
	}

	// Write the password that will be used to encrypt the private key
	// TODO: Probably randomize this and make it differnet per member?
	if err := ioutil.WriteFile(GetPath("ethsigner", "PassFile", "passwordFile"), []byte(`SomeSüper$trÖngPäs$worD!`), 0755); err != nil {
		return err
	}

	// Create genesis.json
	genesis := CreateGenesis(addresses, options.BlockPeriod)

	if err := genesis.WriteGenesisJson(GetPath("besu", "CliqueGenesis.json")); err != nil {
		return err
	}

	return nil
}

func (p *BesuProvider) FirstTimeSetup() error {
	EthSignerConfigPath := filepath.Join(p.Stack.RuntimeDir, "config", "ethsigner")

	ethSignerKeysVolume := fmt.Sprintf("%s_ethsigner_keys", p.Stack.Name)
	docker.CreateVolume(ethSignerKeysVolume, p.Verbose)

	if err := docker.RunDockerCommand(p.Stack.RuntimeDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/ethSigner", EthSignerConfigPath), "-v", fmt.Sprintf("%s:/usr/local/bin/keyFiles", ethSignerKeysVolume), "--entrypoint", "/ethSigner/Nodejs.sh", "node:latest"); err != nil {
		return err
	}

	for i := range p.Stack.Members {
		// Copy ethconnect config to each member's volume
		ethconnectConfigPath := filepath.Join(p.Stack.RuntimeDir, "config", fmt.Sprintf("ethconnect_%v.yaml", i))
		ethconnectConfigVolumeName := fmt.Sprintf("%s_ethconnect_config_%v", p.Stack.Name, i)
		docker.CopyFileToVolume(ethconnectConfigVolumeName, ethconnectConfigPath, "config.yaml", p.Verbose)
	}

	return nil
}

func (p *BesuProvider) DeployFireFlyContract() (*core.BlockchainConfig, error) {
	return ethconnect.DeployFireFlyContract(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart() error {

	return nil
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {

	serviceDefinitions := make([]*docker.ServiceDefinition, 0)
	
	serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
		ServiceName: "member1besu",
		Service: &docker.Service{
			Image: "hyperledger/besu:latest",
			Environment: map[string]string{"OTEL_RESOURCE_ATTRIBUTES": "service.name=member1besu,service.version=${BESU_VERSION:-latest}",
				"NODE_ID": "6"},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./config/besu/networkFiles/member1/keys:/opt/besu/keys",
				// "./config/tessera/networkFiles/member1/tm.pub:/config/tessera/tm.pub"
			},
			EntryPoint: []string{"/config/besu_mem1_def.sh"},
			// DependsOn: map[string]map[string]string{
			// "validator1":     {"condition": "service_healthy"},
			// "member1tessera": {"condition": "service_healthy"}},
			Ports: []string{"20000:8545/tcp", "20001:8546/tcp"},
		},
	})
	// EthSigner Container needs to be defined as,
	// eth_sendTransaction cannot be used to send Transactions on Besu (Besu only accepts raw Transactions)
	serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
		ServiceName: "ethsigner",
		Service: &docker.Service{
			Image: "consensys/ethsigner:develop",
			EntryPoint: []string{
				"/entryPoint/ethsigner.sh",
			},
			Expose: []int{8545},
			Volumes: []string{
				"ethsigner_keys:/keyFiles",
				"./config/ethsigner/PassFile:/PassFile",
				"./config/ethsigner/SignerConfig:/SignerConfig",
				"./config/ethsigner/ethsigner.sh:/entryPoint/ethsigner.sh",
			},
			DependsOn: map[string]map[string]string{
				"validator1": {"condition": "service_healthy"},
				"rpcnode":    {"condition": "service_healthy"}},
			Ports: []string{"18545:8545/tcp"},
			HealthCheck: &docker.HealthCheck{
				Test:     []string{"CMD", "curl", "http://localhost:8545/upcheck"},
				Interval: "5s",
				Retries:  20,
				Timeout:  "5s",
			},
		},
		VolumeNames: []string{"ethsigner_keys"},
	})
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

func (p *BesuProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
	contracts, err := ethereum.ReadCombinedABIJSON(filename)
	if err != nil {
		return []string{}, err
	}
	contractNames := make([]string, len(contracts.Contracts))
	i := 0
	for contractName := range contracts.Contracts {
		contractNames[i] = contractName
	}
	return contractNames, err
}

func (p *BesuProvider) DeployContract(filename, contractName string, member *types.Member, extraArgs []string) (string, error) {
	return ethconnect.DeployCustomContract(member, filename, contractName)
}

func (p *BesuProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	}
}

func (p *BesuProvider) Reset() error {
	return nil
}
