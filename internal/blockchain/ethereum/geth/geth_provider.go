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

package geth

import (
	"fmt"
	"io/ioutil"
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

type GethProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *GethProvider) WriteConfig() error {
	stackDir := filepath.Join(constants.StacksDir, p.Stack.Name)
	for _, member := range p.Stack.Members {
		// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
		if err := ioutil.WriteFile(filepath.Join(stackDir, "blockchain", member.ID, "keyfile"), []byte(member.PrivateKey[2:]), 0755); err != nil {
			return err
		}
	}

	// Create genesis.json
	addresses := make([]string, len(p.Stack.Members))
	for i, member := range p.Stack.Members {
		// Drop the 0x on the front of the address here because that's what geth is expecting in the genesis.json
		addresses[i] = member.Address[2:]
	}
	genesis := ethereum.CreateGenesisJson(addresses)
	if err := genesis.WriteGenesisJson(filepath.Join(stackDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	// Write the password that will be used to encrypt the private key
	// TODO: Probably randomize this and make it differnet per member?
	if err := ioutil.WriteFile(filepath.Join(stackDir, "blockchain", "password"), []byte("correcthorsebatterystaple"), 0755); err != nil {
		return err
	}

	return nil
}

func (p *GethProvider) FirstTimeSetup() error {
	volumeName := fmt.Sprintf("%s_geth", p.Stack.Name)
	gethConfigDir := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")

	// Mount the directory containing all members' private keys and password, and import the accounts using the geth CLI
	for _, member := range p.Stack.Members {
		if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/geth", gethConfigDir), "-v", fmt.Sprintf("%s:/data", volumeName), "ethereum/client-go:release-1.9", "--nousb", "account", "import", "--password", "/geth/password", "--keystore", "/data/keystore", fmt.Sprintf("/geth/%s/keyfile", member.ID)); err != nil {
			return err
		}
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(volumeName, path.Join(gethConfigDir, "genesis.json"), "genesis.json", p.Verbose); err != nil {
		return err
	}

	// Copy the password (to be used for decrypting private keys)
	if err := docker.CopyFileToVolume(volumeName, path.Join(gethConfigDir, "password"), "password", p.Verbose); err != nil {
		return err
	}

	// Initialize the genesis block
	if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/data", volumeName), "ethereum/client-go:release-1.9", "--datadir", "/data", "--nousb", "init", "/data/genesis.json"); err != nil {
		return err
	}

	return nil
}

func (p *GethProvider) PreStart() error {
	return nil
}

func (p *GethProvider) PostStart() error {
	// Unlock accounts
	gethClient := NewGethClient(fmt.Sprintf("http://127.0.0.1:%v", p.Stack.ExposedBlockchainPort))
	for _, m := range p.Stack.Members {
		retries := 10
		p.Log.Info(fmt.Sprintf("unlocking account for member %s", m.ID))
		for {
			if err := gethClient.UnlockAccount(m.Address, "correcthorsebatterystaple"); err != nil {
				if retries == 0 {
					return fmt.Errorf("unable to unlock account %s for member %s", m.Address, m.ID)
				}
				time.Sleep(time.Second * 1)
				retries--
			} else {
				break
			}
		}
	}
	return nil
}

func (p *GethProvider) DeploySmartContracts() error {
	return ethereum.DeployContracts(p.Stack, p.Log, p.Verbose)
}

func (p *GethProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	addresses := ""
	for i, member := range p.Stack.Members {
		addresses = addresses + member.Address
		if i+1 < len(p.Stack.Members) {
			addresses = addresses + ","
		}
	}
	gethCommand := fmt.Sprintf(`--datadir /data --syncmode 'full' --port 30311 --rpcvhosts=* --rpccorsdomain "*" --miner.gastarget 804247552 --rpc --rpcaddr "0.0.0.0" --rpcport 8545 --rpcapi 'admin,personal,db,eth,net,web3,txpool,miner,clique' --networkid 2021 --miner.gasprice 0 --unlock '%s' --password /data/password --mine --nousb --allow-insecure-unlock --nodiscover`, addresses)

	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "geth",
		Service: &docker.Service{
			Image:         "ethereum/client-go:release-1.9",
			ContainerName: fmt.Sprintf("%s_geth", p.Stack.Name),
			Command:       gethCommand,
			Volumes:       []string{"geth:/data"},
			Logging:       docker.StandardLogOptions,
			Ports:         []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"geth"},
	}
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack)...)
	return serviceDefinitions
}

func (p *GethProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
	return &core.BlockchainConfig{
		Type: "ethereum",
		Ethereum: &core.EthereumConfig{
			Ethconnect: &core.EthconnectConfig{
				URL:      p.getEthconnectURL(m),
				Instance: "/contracts/firefly",
				Topic:    m.ID,
			},
		},
	}
}

func (p *GethProvider) Reset() error {
	return nil
}

func (p *GethProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	}
}
