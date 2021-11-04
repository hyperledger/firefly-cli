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
	_ "embed"
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

//go:embed besuCliqueConfig/besu/networkFiles/member1/keys/key.pub
var mem1key_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/member1/keys/key
var mem1key_priv []byte

//go:embed besuCliqueConfig/besu/networkFiles/validator1/keys/key.pub
var val1key_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/validator1/keys/key
var val1key_priv []byte

//go:embed besuCliqueConfig/besu/networkFiles/rpcnode/keys/key.pub
var rpckey_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/rpcnode/keys/key
var rpckey_priv []byte

//go:embed besuCliqueConfig/besu/.env
var besu_env []byte

//go:embed besuCliqueConfig/besu/config.toml
var besu_config []byte

//go:embed besuCliqueConfig/besu/permissions_config.toml
var besu_perm_config []byte

//go:embed besuCliqueConfig/besu/static-nodes.json
var static_nodes []byte

//go:embed besuCliqueConfig/besu/log-config-splunk.xml
var log_config_splunk []byte

//go:embed besuCliqueConfig/besu/log-config.xml
var log_config []byte

//go:embed besuCliqueConfig/EthConnect/eth_connect.sh
var eth_connect_sh []byte

//go:embed besuCliqueConfig/ethsigner/createKeyFile.js
var createKeyFile []byte

//go:embed besuCliqueConfig/ethsigner/ethsigner.sh
var ethsigner_sh []byte

//go:embed besuCliqueConfig/ethsigner/Nodejs.sh
var nodejs_sh []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member1/tm.key
var tessera_mem1_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member1/tm.pub
var tessera_mem1_tmpub []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member2/tm.key
var tessera_mem2_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member2/tm.pub
var tessera_mem2_tmpub []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member3/tm.key
var tessera_mem3_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member3/tm.pub
var tessera_mem3_tmpub []byte

//go:embed besuCliqueConfig/besu_mem1_def.sh
var mem1_entrypt_sh []byte

//go:embed besuCliqueConfig/bootnode_def.sh
var bootnode_def []byte

//go:embed besuCliqueConfig/validator_node_def.sh
var validator_def []byte

//go:embed besuCliqueConfig/tessera_def.sh
var tessera_def []byte

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
	GetPath := func(elem ...string) string { return filepath.Join(append([]string{stackDir, "config"}, elem...)...) }
	if err := os.Mkdir(filepath.Join(stackDir, "config"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(stackDir, "logs"), 0755); err != nil {
		return err
	}
	log_members := []string{"besu", "tessera"}
	for _, members := range log_members {
		if err := os.Mkdir(filepath.Join(stackDir, "logs", members), 0755); err != nil {
			return err
		}
	}
	if err := os.Mkdir(GetPath("besu"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("besu", "networkFiles"), 0755); err != nil {
		return err
	}
	member_directories := []string{"member1", "rpcnode", "validator1"}
	for _, file := range member_directories {
		if err := os.Mkdir(GetPath("besu", "networkFiles", file), 0755); err != nil {
			return err
		}
		if err := os.Mkdir(GetPath("besu", "networkFiles", file, "keys"), 0755); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "member1", "keys", "key.pub"), mem1key_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "member1", "keys", "key"), mem1key_priv, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "validator1", "keys", "key.pub"), val1key_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "validator1", "keys", "key"), val1key_priv, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "rpcnode", "keys", "key.pub"), rpckey_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "rpcnode", "keys", "key"), rpckey_priv, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", ".env"), besu_env, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "config.toml"), besu_config, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "permissions_config.toml"), besu_perm_config, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "static-nodes.json"), static_nodes, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "log-config-splunk.xml"), log_config_splunk, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "log-config.xml"), log_config, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("EthConnect"), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("EthConnect", "eth_connect.sh"), eth_connect_sh, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("ethsigner"), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "createKeyFile.js"), createKeyFile, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "ethsigner.sh"), ethsigner_sh, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "Nodejs.sh"), nodejs_sh, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("tessera"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("tessera", "networkFiles"), 0755); err != nil {
		return err
	}
	tessera_members := []string{"member1", "member2", "member3"}
	for _, member := range tessera_members {
		if err := os.Mkdir(GetPath("tessera", "networkFiles", member), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member1", "tm.key"), tessera_mem1_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member1", "tm.pub"), tessera_mem1_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member2", "tm.key"), tessera_mem2_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member2", "tm.pub"), tessera_mem2_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member3", "tm.key"), tessera_mem3_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member3", "tm.pub"), tessera_mem3_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu_mem1_def.sh"), mem1_entrypt_sh, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("bootnode_def.sh"), bootnode_def, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("validator_node_def.sh"), validator_def, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera_def.sh"), tessera_def, 0755); err != nil {
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
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "PassFile", "passwordFile"), []byte(`SomeSüper$trÖngPäs$worD!`), 0755); err != nil {
		return err
	}

	// Create genesis.json
	genesis := ethereum.CreateBesuCliqueGenesis(addresses)

	if err := genesis.WriteBesuCliqueGenesisJson(GetPath("besu", "CliqueGenesis.json")); err != nil {
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

	serviceDefinitions := make([]*docker.ServiceDefinition, 5)
	// Define bootNode validator container
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
				fmt.Sprintf("%s_default", p.Stack.Name): &docker.IPMapping{
					IPAddress: "172.16.239.11",
				},
			},
		},
	}
	// RPC Node Definition, this container is the JSON-RPC endpoint for Besu
	serviceDefinitions[1] = &docker.ServiceDefinition{
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
	// Tessera Container for enabling Private Transaction Support
	serviceDefinitions[2] = &docker.ServiceDefinition{
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
	// Besu Container depends on Tessera
	serviceDefinitions[3] = &docker.ServiceDefinition{
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
	// EthSigner Container needs to be defined as,
	// eth_sendTransaction cannot be used to send Transactions on Besu (Besu only accepts raw Transactions)
	serviceDefinitions[4] = &docker.ServiceDefinition{
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
