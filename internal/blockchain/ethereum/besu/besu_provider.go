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
	"io"
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
	GetPath := func(elem ...string) string { return filepath.Join(append([]string{stackDir}, elem...)...) }

	source, err := filepath.Abs("internal/blockchain/ethereum/besu/besuConfig")
	if err != nil {
		return err
	}
	if err := CopyDir(source, GetPath("config")); err != nil {
		return err
	}
	// try to make a simplified version by reusing code snippets if possible
	file_names := []string{"accounts", "SignerConfig"}
	for _, file := range file_names {
		if err := os.Mkdir(GetPath("config", "ethsigner", file), 0755); err != nil {
			return err
		}
	}

	addresses := make([]string, len(p.Stack.Members))
	if err := os.Mkdir(GetPath("config", "ethsigner", "PassFile"), 0755); err != nil {
		return err
	}
	for i, member := range p.Stack.Members {
		addresses[i] = member.Address[2:]
		// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
		if err := os.Mkdir(GetPath("config", "ethsigner", "accounts", member.ID), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("config", "ethsigner", "accounts", member.ID, "privateKey"), []byte(member.PrivateKey), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("config", "ethsigner", "accounts", member.ID, "address"), []byte(member.Address[2:]), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(GetPath("config", "ethsigner", "SignerConfig", fmt.Sprintf("%s.toml", member.Address[2:])), []byte(fmt.Sprintf(
			`[metadata]
description = "File based configuration"
[signing]
type = "file-based-signer"
key-file = "%s_keyFile"
password-file = "%s"`, filepath.Join("/keyFiles", member.ID), filepath.Join("/PassFile", "passwordFile"))), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(GetPath("config", "ethsigner", "PassFile", "passwordFile"), []byte(`SomeSüper$trÖngPäs$worD!`), 0755); err != nil {
		return err
	}

	// Create genesis.json
	genesis := ethereum.CreateIBFTGenesis(addresses)

	if err := genesis.WriteIBFTGenesisJson(GetPath("config", "besu", "ibft2Genesis.json")); err != nil {
		return err
	}

	// Write the password that will be used to encrypt the private key
	// TODO: Probably randomize this and make it differnet per member?

	return nil
}
func (p *BesuProvider) FirstTimeSetup() error {
	stackDir := filepath.Join(constants.StacksDir, p.Stack.Name)
	EthSignerConfigPath := filepath.Join(stackDir, "config", "ethsigner")
	if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "run", "--name", "NodeEthSign", "-v", EthSignerConfigPath+":/ethSigner", "--entrypoint", "/ethSigner/Nodejs.sh", "node:latest"); err != nil {
		return err
	}
	if err := docker.RunDockerCommand(constants.StacksDir, p.Verbose, p.Verbose, "cp", "NodeEthSign:/usr/local/bin/keyFiles", EthSignerConfigPath+"/keyFiles"); err != nil {
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

	return nil
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {

	ethConnectConfig := filepath.Join(constants.StacksDir, p.Stack.Name, "config", "EthConnect")

	serviceDefinitions := make([]*docker.ServiceDefinition, 12)

	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "validator1",
		Service: &docker.Service{
			Restart: "on-failure",
			Image:   "hyperledger/besu:latest",
			EnvFile: "./config/besu/.env",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=validator1,service.version=${BESU_VERSION:-latest}",
			},
			Volumes: []string{"public-keys:/tmp/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/validator1/keys:/opt/besu/keys",
			},
			EntryPoint: []string{"/config/bootnode_def.sh"},
			Networks: &docker.Network{
				// NetworkName: &docker.IPMapping{
				// 	IPAddress: "172.16.239.11",
				// },
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.11",
				},
			},
		},
	}
	serviceDefinitions[1] = &docker.ServiceDefinition{
		ServiceName: "validator2",
		Service: &docker.Service{
			Restart: "on-failure",
			Image:   "hyperledger/besu:latest",
			EnvFile: "./config/besu/.env",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=validator2,service.version=${BESU_VERSION:-latest}",
			},
			Volumes: []string{"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/validator2/keys:/opt/besu/keys",
			},
			EntryPoint: []string{"/config/validator_node_def.sh"},
			DependsOn:  map[string]map[string]string{"validator1": {"condition": "service_started"}},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.12",
				},
			},
		},
	}
	serviceDefinitions[2] = &docker.ServiceDefinition{
		ServiceName: "validator3",
		Service: &docker.Service{
			Restart: "on-failure",
			Image:   "hyperledger/besu:latest",
			EnvFile: "./config/besu/.env",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=validator3,service.version=${BESU_VERSION:-latest}",
			},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/validator3/keys:/opt/besu/keys",
			},
			EntryPoint: []string{"/config/validator_node_def.sh"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.13",
				},
			},
			DependsOn: map[string]map[string]string{"validator1": {"condition": "service_started"}},
		},
	}
	serviceDefinitions[3] = &docker.ServiceDefinition{
		ServiceName: "validator4",
		Service: &docker.Service{
			Restart: "on-failure",
			Image:   "hyperledger/besu:latest",
			EnvFile: "./config/besu/.env",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=validator4,service.version=${BESU_VERSION:-latest}",
			},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/validator4/keys:/opt/besu/keys",
			},
			EntryPoint: []string{"/config/validator_node_def.sh"},
			DependsOn:  map[string]map[string]string{"validator1": {"condition": "service_started"}},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.14",
				},
			},
		},
	}
	serviceDefinitions[4] = &docker.ServiceDefinition{
		ServiceName: "rpcnode",
		Service: &docker.Service{
			Restart: "on-failure",
			Image:   "hyperledger/besu:latest",
			EnvFile: "./config/besu/.env",
			HealthCheck: &docker.HealthCheck{
				Test:     []string{"CMD", "curl", "http://localhost:8555/liveness"},
				Interval: "2s",
				Retries:  25,
				Timeout:  "2s",
			},
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=rpcnode,service.version=${BESU_VERSION:-latest}",
			},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/rpcnode/keys:/opt/besu/keys",
			},
			EntryPoint: []string{"/config/validator_node_def.sh"},
			DependsOn:  map[string]map[string]string{"validator1": {"condition": "service_started"}},
			Ports:      []string{"8555:8555/tcp", "8556:8556/tcp"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.15",
				},
			},
		},
	}
	serviceDefinitions[5] = &docker.ServiceDefinition{
		ServiceName: "member1tessera",
		Service: &docker.Service{
			Image:   "quorumengineering/tessera:21.7.0",
			Expose:  []int{9000, 9080, 9101},
			Restart: "no",
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD", "wget", "--spider", "--proxy", "off", "http://localhost:9000/upcheck"},
				Interval: "3s",
				Timeout:  "3s",
				Retries:  20,
			},
			Ports:       []string{"9081:9080"},
			Environment: map[string]string{"TESSERA_CONFIG_TYPE": `"-09"`},
			Volumes: []string{
				"./config:/config",
				"./config/tessera/networkFiles/member1:/config/keys",
				"member1tessera:/data",
				"./logs/tessera:/var/log/tessera/"},
			EntryPoint: []string{"/config/tessera_def.sh"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.26",
				},
			},
		},
	}
	serviceDefinitions[6] = &docker.ServiceDefinition{
		ServiceName: "member1besu",
		Service: &docker.Service{
			Image:   "hyperledger/besu:latest",
			Restart: "on-failure",
			Environment: map[string]string{"OTEL_RESOURCE_ATTRIBUTES": "service.name=member1besu,service.version=${BESU_VERSION:-latest}",
				"NODE_ID": "6"},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./config/besu/networkFiles/member1/keys:/opt/besu/keys",
				"./config/tessera/networkFiles/member1/tm.pub:/config/tessera/tm.pub"},
			EntryPoint: []string{"/config/besu_mem1_def.sh"},
			DependsOn: map[string]map[string]string{
				"validator1":     {"condition": "service_started"},
				"member1tessera": {"condition": "service_started"}},
			Ports: []string{"20000:8545/tcp", "20001:8546/tcp"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.16",
				},
			},
		},
	}
	serviceDefinitions[7] = &docker.ServiceDefinition{
		ServiceName: "member2tessera",
		Service: &docker.Service{
			Image:   "quorumengineering/tessera:21.7.0",
			Expose:  []int{9000, 9080, 9101},
			Restart: "no",
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD", "wget", "--spider", "--proxy", "off", "http://localhost:9000/upcheck"},
				Interval: "3s",
				Timeout:  "3s",
				Retries:  20,
			},
			Environment: map[string]string{"TESSERA_CONFIG_TYPE": `"-09"`},
			Ports:       []string{"9082:9080"},
			Volumes: []string{
				"./config:/config",
				"./config/tessera/networkFiles/member2:/config/keys",
				"member2tessera:/data",
				"./logs/tessera:/var/log/tessera/"},
			EntryPoint: []string{"/config/tessera_def.sh"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.27",
				},
			},
		},
	}
	serviceDefinitions[8] = &docker.ServiceDefinition{
		ServiceName: "member2besu",
		Service: &docker.Service{
			Image:   "hyperledger/besu:latest",
			Restart: "on-failure",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=member2besu,service.version=${BESU_VERSION:-latest}",
				"NODE_ID":                  "7",
			},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/member2/keys:/opt/besu/keys",
				"./config/tessera/networkFiles/member2/tm.pub:/config/tessera/tm.pub"},
			EntryPoint: []string{"/config/besu_mem2_def.sh"},
			DependsOn: map[string]map[string]string{
				"validator1":     {"condition": "service_started"},
				"member2tessera": {"condition": "service_started"}},
			Ports: []string{
				"20002:8555/tcp",
				"20003:8556/tcp"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.17",
				},
			},
		},
	}
	serviceDefinitions[9] = &docker.ServiceDefinition{
		ServiceName: "member3tessera",
		Service: &docker.Service{
			Image:   "quorumengineering/tessera:21.7.0",
			Expose:  []int{9000, 9080, 9101},
			Restart: "no",
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD", "wget", "--spider", "--proxy", "off", "http://localhost:9000/upcheck"},
				Interval: "3s",
				Timeout:  "3s",
				Retries:  20,
			},
			Environment: map[string]string{"TESSERA_CONFIG_TYPE": `"-09"`},
			Ports:       []string{"9083:9080"},
			Volumes: []string{
				"./config:/config",
				"./config/tessera/networkFiles/member3:/config/keys",
				"member3tessera:/data",
				"./logs/tessera:/var/log/tessera/"},
			EntryPoint: []string{"/config/tessera_def.sh"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.28",
				},
			},
		},
	}
	serviceDefinitions[10] = &docker.ServiceDefinition{
		ServiceName: "member3besu",
		Service: &docker.Service{
			Image:   "hyperledger/besu:latest",
			Restart: "on-failure",
			Environment: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=member3besu,service.version=${BESU_VERSION:-latest}",
				"NODE_ID":                  "8",
			},
			Volumes: []string{
				"public-keys:/opt/besu/public-keys/",
				"./config:/config",
				"./logs/besu:/tmp/besu",
				"./config/besu/networkFiles/member3/keys:/opt/besu/keys",
				"./config/tessera/networkFiles/member3/tm.pub:/config/tessera/tm.pub"},
			EntryPoint: []string{"/config/besu_mem3_def.sh"},
			DependsOn: map[string]map[string]string{
				"validator1":     {"condition": "service_started"},
				"member3tessera": {"condition": "service_started"}},
			Ports: []string{
				"20004:8555/tcp",
				"20005:8556/tcp"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.18",
				},
			},
		},
	}
	serviceDefinitions[11] = &docker.ServiceDefinition{
		ServiceName: "ethsigner",
		Service: &docker.Service{
			Image: "consensys/ethsigner:develop",
			EntryPoint: []string{
				"/entryPoint/ethsigner.sh",
			},
			Expose: []int{8545},
			Volumes: []string{
				"./config/ethsigner/keyFiles:/keyFiles",
				"./config/ethsigner/PassFile:/PassFile",
				"./config/ethsigner/SignerConfig:/SignerConfig",
				"./config/ethsigner/ethsigner.sh:/entryPoint/ethsigner.sh",
			},
			DependsOn: map[string]map[string]string{
				"validator1": {"condition": "service_started"},
				"rpcnode":    {"condition": "service_healthy"}},
			Ports: []string{"18545:8545/tcp"},
			Networks: &docker.Network{
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.40",
				},
			},
			HealthCheck: &docker.HealthCheck{
				Test:     []string{"CMD", "curl", "http://localhost:8545/upcheck"},
				Interval: "5s",
				Retries:  20,
				Timeout:  "5s",
			},
		},
	}
	serviceDefinitions = append(serviceDefinitions, ethconnect.GetEthconnectServiceDefinitions(p.Stack, "besu", ethConnectConfig)...)
	return serviceDefinitions
}

func (p *BesuProvider) GetFireflyConfig(m *types.Member) (blockchainConfig *core.BlockchainConfig, orgConfig *core.OrgConfig) {
	orgConfig = &core.OrgConfig{
		Name:     m.OrgName,
		Identity: m.Address,
	}
	blockchainConfig = &core.BlockchainConfig{
		Type: "ethereum",
		Ethereum: &core.EthereumConfig{
			Ethconnect: &core.EthconnectConfig{
				URL:      p.getEthconnectURL(m),
				Instance: "/contracts/firefly",
				Topic:    m.ID,
			},
		},
	}
	return
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

func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
