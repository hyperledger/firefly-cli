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

package quorum

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/evmconnect"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/tessera"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var quorumImage = "quorumengineering/quorum:24.4"
var tesseraImage = "quorumengineering/tessera:24.4"
var ExposedBlockchainPortMultiplier = 10

// TODO: Probably randomize this and make it different per member?
var keyPassword = "correcthorsebatterystaple"

type QuorumProvider struct {
	ctx       context.Context
	stack     *types.Stack
	connector connector.Connector
	dockerMgr docker.IDockerManager
}

func NewQuorumProvider(ctx context.Context, stack *types.Stack) *QuorumProvider {
	var connector connector.Connector
	switch stack.BlockchainConnector {
	case types.BlockchainConnectorEthconnect:
		connector = ethconnect.NewEthconnect(ctx)
	case types.BlockchainConnectorEvmconnect:
		connector = evmconnect.NewEvmconnect(ctx)
	}

	return &QuorumProvider{
		ctx:       ctx,
		stack:     stack,
		connector: connector,
		dockerMgr: docker.NewDockerManager(),
	}
}

func (p *QuorumProvider) WriteConfig(options *types.InitOptions) error {
	l := log.LoggerFromContext(p.ctx)
	initDir := p.stack.InitDir
	for i, member := range p.stack.Members {
		// Generate the connector config for each member
		connectorConfigPath := filepath.Join(initDir, "config", fmt.Sprintf("%s_%v.yaml", p.connector.Name(), i))
		if err := p.connector.GenerateConfig(p.stack, member, fmt.Sprintf("quorum_%d", i)).WriteConfig(connectorConfigPath, options.ExtraConnectorConfigPath); err != nil {
			return nil
		}

		// Generate tessera docker-entrypoint for each member
		if p.stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
			l.Info(fmt.Sprintf("generating tessera docker-entrypoint file for member %d", i))
			tesseraEntrypointOutputDirectory := filepath.Join(initDir, "tessera", fmt.Sprintf("tessera_%d", i))
			if err := tessera.CreateTesseraEntrypoint(p.ctx, tesseraEntrypointOutputDirectory, p.stack.Name, len(p.stack.Members)); err != nil {
				return err
			}
		}

		// Generate quorum docker-entrypoint for each member
		l.Info(fmt.Sprintf("generating quorum docker-entrypoint file for member %d", i))
		quorumEntrypointOutputDirectory := filepath.Join(initDir, "blockchain", fmt.Sprintf("quorum_%d", i))
		if err := CreateQuorumEntrypoint(p.ctx, quorumEntrypointOutputDirectory, p.stack.Consensus.String(), p.stack.Name, i, int(p.stack.ChainID()), options.BlockPeriod, p.stack.PrivateTransactionManager); err != nil {
			return err
		}
	}

	// Create genesis.json
	addresses := make([]string, len(p.stack.Members))
	for i, member := range p.stack.Members {
		address := member.Account.(*ethereum.Account).Address
		// Drop the 0x on the front of the address here because that's what quorum is expecting in the genesis.json
		addresses[i] = address[2:]
	}
	genesis := CreateGenesis(addresses, options.BlockPeriod, p.stack.ChainID())
	if err := genesis.WriteGenesisJSON(filepath.Join(initDir, "blockchain", "genesis.json")); err != nil {
		return err
	}

	return nil
}

func (p *QuorumProvider) FirstTimeSetup() error {
	quorumVolumeName := fmt.Sprintf("%s_quorum", p.stack.Name)
	tesseraVolumeName := fmt.Sprintf("%s_tessera", p.stack.Name)
	blockchainDir := path.Join(p.stack.RuntimeDir, "blockchain")
	tesseraDir := path.Join(p.stack.RuntimeDir, "tessera")
	contractsDir := path.Join(p.stack.RuntimeDir, "contracts")
	rootDir := "/"

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
		if err := p.dockerMgr.CopyFileToVolume(p.ctx, connectorConfigVolumeName, connectorConfigPath, "config.yaml"); err != nil {
			return err
		}

		// Volume name instantiation
		quorumVolumeNameMember := fmt.Sprintf("%s_%d", quorumVolumeName, i)
		tesseraVolumeNameMember := fmt.Sprintf("%s_%d", tesseraVolumeName, i)

		// Copy the wallet files of each member to their respective blockchain volume
		keystoreDirectory := filepath.Join(blockchainDir, fmt.Sprintf("quorum_%d", i), "keystore")
		if err := p.dockerMgr.CopyFileToVolume(p.ctx, quorumVolumeNameMember, keystoreDirectory, "/"); err != nil {
			return err
		}

		if p.stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
			// Copy member specific tessera key files
			if err := p.dockerMgr.MkdirInVolume(p.ctx, tesseraVolumeNameMember, rootDir); err != nil {
				return err
			}
			tmKeystoreDirectory := filepath.Join(tesseraDir, fmt.Sprintf("tessera_%d", i), "keystore")
			if err := p.dockerMgr.CopyFileToVolume(p.ctx, tesseraVolumeNameMember, tmKeystoreDirectory, rootDir); err != nil {
				return err
			}
			// Copy tessera docker-entrypoint file
			tmEntrypointPath := filepath.Join(tesseraDir, fmt.Sprintf("tessera_%d", i), tessera.DockerEntrypoint)
			if err := p.dockerMgr.CopyFileToVolume(p.ctx, tesseraVolumeNameMember, tmEntrypointPath, rootDir); err != nil {
				return err
			}
		}

		// Copy quorum docker-entrypoint file
		quorumEntrypointPath := filepath.Join(blockchainDir, fmt.Sprintf("quorum_%d", i), tessera.DockerEntrypoint)
		if err := p.dockerMgr.CopyFileToVolume(p.ctx, quorumVolumeNameMember, quorumEntrypointPath, rootDir); err != nil {
			return err
		}

		// Copy the genesis block information
		if err := p.dockerMgr.CopyFileToVolume(p.ctx, quorumVolumeNameMember, path.Join(blockchainDir, "genesis.json"), "genesis.json"); err != nil {
			return err
		}

		// Initialize the genesis block
		if err := p.dockerMgr.RunDockerCommand(p.ctx, p.stack.StackDir, "run", "--rm", "-v", fmt.Sprintf("%s:/data", quorumVolumeNameMember), quorumImage, "--datadir", "/data", "init", "/data/genesis.json"); err != nil {
			return err
		}
	}

	return nil
}

func (p *QuorumProvider) PreStart() error {
	return nil
}

func (p *QuorumProvider) PostStart(firstTimeSetup bool) error {
	l := log.LoggerFromContext(p.ctx)
	// Unlock accounts
	for _, account := range p.stack.State.Accounts {
		address := account.(*ethereum.Account).Address
		l.Info(fmt.Sprintf("unlocking account %s", address))
		// Check which member the account belongs to
		var memberIndex int
		for _, member := range p.stack.Members {
			if member.Account.(*ethereum.Account).Address == address {
				memberIndex = *member.Index
				break
			}
		}
		if err := p.unlockAccount(address, keyPassword, memberIndex); err != nil {
			return err
		}
	}

	return nil
}

func (p *QuorumProvider) unlockAccount(address, password string, memberIndex int) error {
	l := log.LoggerFromContext(p.ctx)
	verbose := log.VerbosityFromContext(p.ctx)
	// exposed blockchain port is the default for node 0, we need to add the port multiplier to get the right rpc for the correct node
	quorumClient := NewQuorumClient(fmt.Sprintf("http://127.0.0.1:%v", p.stack.ExposedBlockchainPort+(memberIndex*ExposedBlockchainPortMultiplier)))
	retries := 10
	for {
		if err := quorumClient.UnlockAccount(address, password); err != nil {
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

func (p *QuorumProvider) DeployFireFlyContract() (*types.ContractDeploymentResult, error) {
	contract, err := ethereum.ReadFireFlyContract(p.ctx, p.stack)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contract, "FireFly", p.stack.Members[0], nil)
}

func (p *QuorumProvider) buildTesseraServiceDefinition(memberIdx int) *docker.ServiceDefinition {
	return &docker.ServiceDefinition{
		ServiceName: fmt.Sprintf("tessera_%d", memberIdx),
		Service: &docker.Service{
			Image:         tesseraImage,
			ContainerName: fmt.Sprintf("%s_member%dtessera", p.stack.Name, memberIdx),
			Volumes:       []string{fmt.Sprintf("tessera_%d:/data", memberIdx)},
			Logging:       docker.StandardLogOptions,
			Ports:         []string{fmt.Sprintf("%d:%s", p.stack.ExposedPtmPort+(memberIdx*ExposedBlockchainPortMultiplier), tessera.TmTpPort)}, // defaults 4100, 4110, 4120, 4130
			Environment:   p.stack.EnvironmentVars,
			EntryPoint:    []string{"/bin/sh", "-c", "/data/docker-entrypoint.sh"},
			Deploy:        map[string]interface{}{"restart_policy": map[string]interface{}{"condition": "on-failure", "max_attempts": int64(3)}},
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD",
					"curl",
					"--fail",
					fmt.Sprintf("http://localhost:%s/upcheck", tessera.TmTpPort),
				},
				Interval: "15s", // 6000 requests in a day
				Retries:  30,
			},
		},
		VolumeNames: []string{fmt.Sprintf("tessera_%d", memberIdx)},
	}
}

func (p *QuorumProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	memberCount := len(p.stack.Members)
	serviceDefinitionsCount := memberCount
	if p.stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
		serviceDefinitionsCount *= 2
	}
	serviceDefinitions := make([]*docker.ServiceDefinition, serviceDefinitionsCount)
	connectorDependents := map[string]string{}
	for i := 0; i < memberCount; i++ {
		var quorumDependsOn map[string]map[string]string
		if p.stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
			quorumDependsOn = map[string]map[string]string{fmt.Sprintf("tessera_%d", i): {"condition": "service_healthy"}}
			serviceDefinitions[i+memberCount] = p.buildTesseraServiceDefinition(i)
			// No arm64 images for Tessera
			if runtime.GOARCH == "arm64" {
				serviceDefinitions[i+memberCount].Service.Platform = "linux/amd64"
			}
		}
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: fmt.Sprintf("quorum_%d", i),
			Service: &docker.Service{
				Image:         quorumImage,
				ContainerName: fmt.Sprintf("%s_quorum_%d", p.stack.Name, i),
				Volumes:       []string{fmt.Sprintf("quorum_%d:/data", i)},
				Logging:       docker.StandardLogOptions,
				Ports:         []string{fmt.Sprintf("%d:8545", p.stack.ExposedBlockchainPort+(i*ExposedBlockchainPortMultiplier))}, // defaults 5100, 5110, 5120, 5130
				Environment:   p.stack.EnvironmentVars,
				EntryPoint:    []string{"/bin/sh", "-c", "/data/docker-entrypoint.sh"},
				DependsOn:     quorumDependsOn,
			},
			VolumeNames: []string{fmt.Sprintf("quorum_%d", i)},
		}
		connectorDependents[fmt.Sprintf("quorum_%d", i)] = "service_started"
	}
	serviceDefinitions = append(serviceDefinitions, p.connector.GetServiceDefinitions(p.stack, connectorDependents)...)
	return serviceDefinitions
}

func (p *QuorumProvider) GetBlockchainPluginConfig(stack *types.Stack, m *types.Organization) (blockchainConfig *types.BlockchainConfig) {
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

func (p *QuorumProvider) GetOrgConfig(stack *types.Stack, m *types.Organization) (orgConfig *types.OrgConfig) {
	account := m.Account.(*ethereum.Account)
	orgConfig = &types.OrgConfig{
		Name: m.OrgName,
		Key:  account.Address,
	}
	return
}

func (p *QuorumProvider) Reset() error {
	return nil
}

func (p *QuorumProvider) GetContracts(filename string, extraArgs []string) ([]string, error) {
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

func (p *QuorumProvider) DeployContract(filename, contractName, instanceName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	contracts, err := ethereum.ReadContractJSON(filename)
	if err != nil {
		return nil, err
	}
	return p.connector.DeployContract(contracts.Contracts[contractName], instanceName, member, extraArgs)
}

func (p *QuorumProvider) CreateAccount(args []string) (interface{}, error) {
	l := log.LoggerFromContext(p.ctx)
	memberIndex := args[2]
	quorumVolumeName := fmt.Sprintf("%s_quorum_%s", p.stack.Name, memberIndex)
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
	outputDirectory := filepath.Join(directory, "blockchain", fmt.Sprintf("quorum_%s", memberIndex), "keystore")
	keyPair, walletFilePath, err := ethereum.CreateWalletFile(outputDirectory, prefix, keyPassword)
	if err != nil {
		return nil, err
	}

	// Tessera is an optional add-on to the quorum blockchain node provider
	var tesseraPubKey, tesseraKeysPath string
	if p.stack.PrivateTransactionManager.Equals(types.PrivateTransactionManagerTessera) {
		tesseraKeysOutputDirectory := filepath.Join(directory, "tessera", fmt.Sprintf("tessera_%s", memberIndex), "keystore")
		_, tesseraPubKey, tesseraKeysPath, err = tessera.CreateTesseraKeys(p.ctx, tesseraImage, tesseraKeysOutputDirectory, "", "tm")
		if err != nil {
			return nil, err
		}
		l.Info(fmt.Sprintf("keys generated in %s", tesseraKeysPath))
	}

	if stackHasRunBefore {
		if err := ethereum.CopyWalletFileToVolume(p.ctx, walletFilePath, quorumVolumeName); err != nil {
			return nil, err
		}
		if memberIndexInt, err := strconv.Atoi(memberIndex); err != nil {
			return nil, err
		} else {
			if err := p.unlockAccount(keyPair.Address.String(), keyPassword, memberIndexInt); err != nil {
				return nil, err
			}
		}
	}

	return &ethereum.Account{
		Address:      keyPair.Address.String(),
		PrivateKey:   hex.EncodeToString(keyPair.PrivateKeyBytes()),
		PtmPublicKey: tesseraPubKey,
	}, nil
}

func (p *QuorumProvider) ParseAccount(account interface{}) interface{} {
	accountMap := account.(map[string]interface{})
	ptmPublicKey := "" // if we start quorum without tessera, no public key will be generated
	v, ok := accountMap["ptmPublicKey"]
	if ok {
		ptmPublicKey = v.(string)
	}

	return &ethereum.Account{
		Address:      accountMap["address"].(string),
		PrivateKey:   accountMap["privateKey"].(string),
		PtmPublicKey: ptmPublicKey,
	}
}

func (p *QuorumProvider) GetConnectorName() string {
	return p.connector.Name()
}

func (p *QuorumProvider) GetConnectorURL(org *types.Organization) string {
	return fmt.Sprintf("http://%s_%s:%v", p.connector.Name(), org.ID, p.connector.Port())
}

func (p *QuorumProvider) GetConnectorExternalURL(org *types.Organization) string {
	return fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort)
}
