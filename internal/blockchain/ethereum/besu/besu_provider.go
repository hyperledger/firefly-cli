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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger-labs/firefly-cli/internal/constants"
	"github.com/hyperledger-labs/firefly-cli/internal/core"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/internal/log"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type BesuClient struct {
	rpcUrl string
}

type BesuProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

type RpcRequest struct {
	JsonRPC string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

func NewBesuClient(rpcUrl string) *BesuClient {
	return &BesuClient{
		rpcUrl: rpcUrl,
	}
}

func (p *BesuProvider) WriteConfig() error {

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

func (p *BesuProvider) RunFirstTimeSetup() error {
	volumeName := fmt.Sprintf("%s_besu", p.Stack.Name)
	besuConfigDir := path.Join(constants.StacksDir, p.Stack.Name, "blockchain")

	// Mount the directory containing all members' private keys and password, and import the accounts using the geth CLI
	for _, member := range p.Stack.Members {
		if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/besu", besuConfigDir), "-v", fmt.Sprintf("%s:/data", volumeName), "ethereum/client-go:release-1.9", "--nousb", "account", "import", "--password", "/besu/password", "--keystore", "/data/keystore", fmt.Sprintf("/besu/%s/keyfile", member.ID)); err != nil {
			return err
		}
	}

	// Copy the genesis block information
	if err := docker.CopyFileToVolume(volumeName, path.Join(besuConfigDir, "genesis.json"), "genesis.json", p.Verbose); err != nil {
		return err
	}

	// Copy the password (to be used for decrypting private keys)
	if err := docker.CopyFileToVolume(volumeName, path.Join(besuConfigDir, "password"), "password", p.Verbose); err != nil {
		return err
	}

	// Initialize the genesis block
	if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "run", "--rm", "-v", fmt.Sprintf("%s:/data", volumeName), "hyperledger/besu:latest", "--datadir", "/data", "--nousb", "init", "/data/genesis.json"); err != nil {
		return err
	}
	return nil
}

func (p *BesuProvider) DeploySmartContracts() error {
	return ethereum.DeployContracts(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart() error {
	// Unlock accounts
	gethClient := NewBesuClient(fmt.Sprintf("http://127.0.0.1:%v", p.Stack.ExposedBlockchainPort))
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

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	addresses := ""
	for i, member := range p.Stack.Members {
		addresses = addresses + member.Address
		if i+1 < len(p.Stack.Members) {
			addresses = addresses + ","
		}
	}
	besuCommand := fmt.Sprintf(`--datadir /data --syncmode 'full' --port 30311 --rpcvhosts=* --rpccorsdomain "*" --miner.gastarget 804247552 --rpc --rpcaddr "127.0.0.1" --rpcport 8545 --rpcapi 'admin,personal,db,eth,net,web3,txpool,miner,clique' --networkid 1337 --miner.gasprice 0 --unlock '%s' --password /data/password --mine --nousb --allow-insecure-unlock --nodiscover`, addresses)
	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "besu",
		Service: &docker.Service{
			Image:   "hyperledger/besu:latest",
			Command: besuCommand,
			Volumes: []string{"besu:/data"},
			Logging: docker.StandardLogOptions,
			Ports:   []string{fmt.Sprintf("%d:8545", p.Stack.ExposedBlockchainPort)},
		},
		VolumeNames: []string{"besu"},
	}
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack.Members)...)
	return serviceDefinitions
}

func (p *BesuProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
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

func (p *BesuProvider) getEthconnectURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	}
}

func (g *BesuClient) UnlockAccount(address string, password string) error {
	requestBody, err := json.Marshal(&RpcRequest{
		JsonRPC: "2.0",
		ID:      0,
		Method:  "personal_unlockAccount",
		Params:  []string{address, password},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", g.rpcUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%d %s", resp.StatusCode, responseBody)
	}
	return nil
}
