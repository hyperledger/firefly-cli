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

package stacks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hyperledger/firefly-cli/internal/blockchain"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/besu"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethsigner"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/geth"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/remoterpc"
	"github.com/hyperledger/firefly-cli/internal/blockchain/fabric"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/tokens"
	"github.com/hyperledger/firefly-cli/internal/tokens/erc1155"
	"github.com/hyperledger/firefly-cli/internal/tokens/erc20erc721"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"

	"gopkg.in/yaml.v2"

	"github.com/hyperledger/firefly-cli/internal/log"

	"github.com/otiai10/copy"
)

type StackManager struct {
	Log                    log.Logger
	Stack                  *types.Stack
	blockchainProvider     blockchain.IBlockchainProvider
	tokenProviders         []tokens.ITokensProvider
	fireflyCoreEntrypoints [][]string
	IsOldFileStructure     bool
}

func ListStacks() ([]string, error) {
	files, err := ioutil.ReadDir(constants.StacksDir)
	if err != nil {
		return nil, err
	}

	stacks := make([]string, 0)
	i := 0
	for _, f := range files {
		if f.IsDir() {
			if exists, err := CheckExists(f.Name()); err == nil && exists {
				stacks = append(stacks, f.Name())
				i++
			}
		}
	}
	return stacks, nil
}

func NewStackManager(logger log.Logger) *StackManager {
	return &StackManager{
		Log: logger,
	}
}

func (s *StackManager) InitStack(stackName string, memberCount int, options *types.InitOptions) (err error) {
	s.Stack = &types.Stack{
		Name:                   stackName,
		Members:                make([]*types.Member, memberCount),
		SwarmKey:               GenerateSwarmKey(),
		ExposedBlockchainPort:  options.ServicesBasePort,
		Database:               options.DatabaseSelection.String(),
		BlockchainProvider:     options.BlockchainProvider.String(),
		BlockchainNodeProvider: options.BlockchainNodeProvider.String(),
		TokenProviders:         options.TokenProviders,
		ContractAddress:        options.ContractAddress,
		StackDir:               filepath.Join(constants.StacksDir, stackName),
		InitDir:                filepath.Join(constants.StacksDir, stackName, "init"),
		RuntimeDir:             filepath.Join(constants.StacksDir, stackName, "runtime"),
		State: &types.StackState{
			DeployedContracts: make([]*types.DeployedContract, 0),
			Accounts:          make([]interface{}, memberCount),
		},
		SandboxEnabled: options.SandboxEnabled,
		FFTMEnabled:    options.FFTMEnabled,
		ChainIDPtr:     &options.ChainID,
	}

	if options.PrometheusEnabled {
		s.Stack.PrometheusEnabled = true
		s.Stack.ExposedPrometheusPort = options.PrometheusPort
	}

	var manifest *types.VersionManifest

	if options.ManifestPath != "" {
		// If a path to a manifest file is set, read the existing file
		manifest, err = core.ReadManifestFile(options.ManifestPath)
		if err != nil {
			return err
		}
	} else {
		// Otherwise, fetch the manifest file from GitHub for the specified version
		if options.FireFlyVersion == "" || strings.ToLower(options.FireFlyVersion) == "latest" {
			manifest, err = core.GetLatestReleaseManifest()
			if err != nil {
				return err
			}
		} else {
			manifest, err = core.GetReleaseManifest(options.FireFlyVersion)
			if err != nil {
				return err
			}
		}
	}

	s.Stack.VersionManifest = manifest
	s.blockchainProvider = s.getBlockchainProvider(false)
	s.tokenProviders = s.getITokenProviders(false)

	for i := 0; i < memberCount; i++ {
		externalProcess := i < options.ExternalProcesses
		s.Stack.Members[i] = createMember(fmt.Sprint(i), i, options, externalProcess)
		if s.Stack.Members[i].Account != nil {
			s.Stack.State.Accounts[i] = s.Stack.Members[i].Account
		}
	}
	compose := docker.CreateDockerCompose(s.Stack)
	extraServices := s.blockchainProvider.GetDockerServiceDefinitions()
	for i, tp := range s.tokenProviders {
		extraServices = append(extraServices, tp.GetDockerServiceDefinitions(i)...)
	}

	for _, serviceDefinition := range extraServices {
		// Add each service definition to the docker compose file
		compose.Services[serviceDefinition.ServiceName] = serviceDefinition.Service
		// Add the volume name for each volume used by this service
		for _, volumeName := range serviceDefinition.VolumeNames {
			compose.Volumes[volumeName] = struct{}{}
		}

		// Add a dependency so each firefly core container won't start up until dependencies are up
		for _, member := range s.Stack.Members {
			if service, ok := compose.Services[fmt.Sprintf("firefly_core_%v", *member.Index)]; ok {
				condition := "service_started"
				if serviceDefinition.Service.HealthCheck != nil {
					condition = "service_healthy"
				}
				service.DependsOn[serviceDefinition.ServiceName] = map[string]string{"condition": condition}
			}
		}
	}

	if err := s.ensureInitDirectories(); err != nil {
		return err
	}
	if err := s.writeDockerCompose(s.Stack.InitDir, compose); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %s", err)
	}
	return s.writeConfig(options)
}

func CheckExists(stackName string) (bool, error) {
	_, err := os.Stat(filepath.Join(constants.StacksDir, stackName, "stack.json"))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func isOldFileStructure(stackDir string) (bool, error) {
	_, err := os.Stat(filepath.Join(stackDir, "init"))
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}
}

func (s *StackManager) LoadStack(stackName string, verbose bool) error {
	stackDir := filepath.Join(constants.StacksDir, stackName)
	exists, err := CheckExists(stackName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("stack '%s' does not exist", stackName)
	}
	d, err := ioutil.ReadFile(filepath.Join(stackDir, "stack.json"))
	if err != nil {
		return err
	}
	var stack *types.Stack
	if err := json.Unmarshal(d, &stack); err != nil {
		return err
	}
	s.Stack = stack
	s.Stack.StackDir = stackDir
	s.blockchainProvider = s.getBlockchainProvider(verbose)
	s.tokenProviders = s.getITokenProviders(verbose)

	isOldFileStructure, err := isOldFileStructure(s.Stack.StackDir)
	if err != nil {
		return err
	}

	if !isOldFileStructure {
		s.Stack.InitDir = filepath.Join(s.Stack.StackDir, "init")
		s.Stack.RuntimeDir = filepath.Join(s.Stack.StackDir, "runtime")
	} else {
		s.IsOldFileStructure = true
		s.Stack.InitDir = s.Stack.StackDir
		s.Stack.RuntimeDir = s.Stack.StackDir
	}

	for _, member := range stack.Members {
		if member.Account != nil {
			member.Account = s.blockchainProvider.ParseAccount(member.Account)
		}
	}

	// For backwards compatability, add a "default" VersionManifest
	// in memory for stacks that were created with old CLI versions
	if s.Stack.VersionManifest == nil {
		s.Stack.VersionManifest = &types.VersionManifest{
			FireFly: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly",
				Tag:   "latest",
			},
			Ethconnect: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly-ethconnect",
				Tag:   "latest",
			},
			Fabconnect: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly-fabconnect",
				Tag:   "latest",
			},
			DataExchange: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly-dataexchange-https",
				Tag:   "latest",
			},
			TokensERC1155: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly-tokens-erc1155",
				Tag:   "latest",
			},
			TokensERC20ERC721: &types.ManifestEntry{
				Image: "ghcr.io/hyperledger/firefly-tokens-erc20-erc721",
				Tag:   "latest",
			},
		}
	}

	stackHasRunBefore, err := s.StackHasRunBefore()
	if err != nil {
		return nil
	}
	if stackHasRunBefore {
		return s.loadStackStateJSON()
	} else {
		s.Stack.State = &types.StackState{}
	}
	return nil
}

func (s *StackManager) loadStackStateJSON() error {
	stackStatePath := filepath.Join(s.Stack.RuntimeDir, "stackState.json")
	_, err := os.Stat(stackStatePath)
	if os.IsNotExist(err) {
		// Initialize with an empty StackState
		s.Stack.State = &types.StackState{}
		return nil
	} else if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(stackStatePath)
	if err != nil {
		return err
	}
	var stackState *types.StackState
	if err := json.Unmarshal(b, &stackState); err != nil {
		return err
	}

	for i, account := range stackState.Accounts {
		if account != nil {
			stackState.Accounts[i] = s.blockchainProvider.ParseAccount(account)
		}
	}

	s.Stack.State = stackState
	return nil
}

func (s *StackManager) writeStackStateJSON(directory string) error {
	stackStateBytes, err := json.MarshalIndent(s.Stack.State, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(directory, "stackState.json"), stackStateBytes, 0755)
}

func (s *StackManager) ensureInitDirectories() error {
	configDir := filepath.Join(s.Stack.InitDir, "config")

	if err := os.MkdirAll(filepath.Join(configDir), 0755); err != nil {
		return err
	}

	for _, member := range s.Stack.Members {
		if err := os.MkdirAll(filepath.Join(configDir, "dataexchange_"+member.ID, "peer-certs"), 0755); err != nil {
			return err
		}
	}

	return nil
}

func (s *StackManager) writeDockerCompose(workingDir string, compose *docker.DockerComposeConfig) error {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(workingDir, "docker-compose.yml"), bytes, 0755)
}

func (s *StackManager) writeStackConfig() error {
	stackConfigBytes, err := json.MarshalIndent(s.Stack, "", " ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(s.Stack.StackDir, "stack.json"), stackConfigBytes, 0755); err != nil {
		return err
	}
	return s.writeStackStateJSON(s.Stack.InitDir)
}

func (s *StackManager) writeConfig(options *types.InitOptions) error {
	if err := s.writeDataExchangeCerts(options.Verbose); err != nil {
		return err
	}

	for _, member := range s.Stack.Members {
		config := core.NewFireflyConfig(s.Stack, member)
		config.Blockchain, config.Org = s.blockchainProvider.GetFireflyConfig(s.Stack, member)
		for iTok, tp := range s.tokenProviders {
			config.Tokens = append(config.Tokens, tp.GetFireflyConfig(member, iTok))
		}
		coreConfigFilename := filepath.Join(s.Stack.InitDir, "config", fmt.Sprintf("firefly_core_%s.yml", member.ID))
		if err := core.WriteFireflyConfig(config, coreConfigFilename, options.ExtraCoreConfigPath); err != nil {
			return err
		}
		if s.Stack.FFTMEnabled {
			fftmConfig := NewFFTMConfig(s.Stack, member)
			fftmConfigFilename := filepath.Join(s.Stack.InitDir, "config", fmt.Sprintf("firefly_fftm_%s.yml", member.ID))
			if err := WriteFFTMConfig(fftmConfig, fftmConfigFilename); err != nil {
				return err
			}
		}
	}

	if err := s.writeStackConfig(); err != nil {
		return err
	}

	if err := s.blockchainProvider.WriteConfig(options); err != nil {
		return err
	}

	if s.Stack.PrometheusEnabled {
		promConfig := s.GeneratePrometheusConfig()
		configBytes, err := yaml.Marshal(promConfig)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(s.Stack.InitDir, "config", "prometheus.yml"), configBytes, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (s *StackManager) writeDataExchangeCerts(verbose bool) error {
	configDir := filepath.Join(s.Stack.InitDir, "config")
	for _, member := range s.Stack.Members {

		memberDXDir := path.Join(configDir, "dataexchange_"+member.ID)

		// TODO: remove dependency on openssl here
		opensslCmd := exec.Command("openssl", "req", "-new", "-x509", "-nodes", "-days", "365", "-subj", fmt.Sprintf("/CN=dataexchange_%s/O=member_%s", member.ID, member.ID), "-keyout", "key.pem", "-out", "cert.pem")
		opensslCmd.Dir = filepath.Join(configDir, "dataexchange_"+member.ID)
		if err := opensslCmd.Run(); err != nil {
			return err
		}

		dataExchangeConfig := s.GenerateDataExchangeHTTPSConfig(member.ID)
		configBytes, err := json.Marshal(dataExchangeConfig)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(memberDXDir, "config.json"), configBytes, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) copyDataExchangeConfigToVolumes(verbose bool) error {
	configDir := filepath.Join(s.Stack.RuntimeDir, "config")
	for _, member := range s.Stack.Members {
		// Copy files into docker volumes
		memberDXDir := path.Join(configDir, "dataexchange_"+member.ID)
		volumeName := fmt.Sprintf("%s_dataexchange_%s", s.Stack.Name, member.ID)
		docker.MkdirInVolume(volumeName, "peer-certs", verbose)
		if err := docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "config.json"), "/config.json", verbose); err != nil {
			return err
		}
		if err := docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "cert.pem"), "/cert.pem", verbose); err != nil {
			return err
		}
		if err := docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "key.pem"), "/key.pem", verbose); err != nil {
			return err
		}
	}
	return nil
}

func createMember(id string, index int, options *types.InitOptions, external bool) *types.Member {
	serviceBase := options.ServicesBasePort + (index * 100)
	member := &types.Member{
		ID:                      id,
		Index:                   &index,
		ExposedFireflyPort:      options.FireFlyBasePort + index,
		ExposedFireflyAdminPort: serviceBase + 1, // note shared blockchain node is on zero
		ExposedConnectorPort:    serviceBase + 2,
		ExposedUIPort:           serviceBase + 3,
		ExposedDatabasePort:     serviceBase + 4,
		ExposedDataexchangePort: serviceBase + 5,
		ExposedIPFSApiPort:      serviceBase + 6,
		ExposedIPFSGWPort:       serviceBase + 7,
		External:                external,
		OrgName:                 options.OrgNames[index],
		NodeName:                options.NodeNames[index],
	}
	nextPort := serviceBase + 8
	if options.PrometheusEnabled {
		member.ExposedFireflyMetricsPort = nextPort
		nextPort++
	}
	for range options.TokenProviders {
		member.ExposedTokensPorts = append(member.ExposedTokensPorts, nextPort)
		nextPort++
	}

	switch options.BlockchainProvider {
	case types.Ethereum:
		address, privateKey := ethereum.GenerateAddressAndPrivateKey()
		member.Account = &ethereum.Account{
			Address:    address,
			PrivateKey: privateKey,
		}
	case types.HyperledgerFabric:
		// This will be filled in by the Fabric blockchain provider
	}
	if options.SandboxEnabled {
		member.ExposedSandboxPort = nextPort
		nextPort++
	}
	if options.FFTMEnabled {
		member.ExposedFFTMPort = nextPort
		nextPort++
	}
	return member
}

func (s *StackManager) StartStack(verbose bool, options *types.StartOptions) (messages []string, err error) {
	fmt.Printf("starting FireFly stack '%s'... ", s.Stack.Name)
	// Check to make sure all of our ports are available
	err = s.checkPortsAvailable()
	if err != nil {
		return messages, err
	}
	hasBeenRun, err := s.StackHasRunBefore()
	if err != nil {
		return messages, err
	}
	if !hasBeenRun {
		setupMessages, err := s.runFirstTimeSetup(verbose, options)
		messages = append(messages, setupMessages...)
		if err != nil {
			// Something bad happened during setup
			if options.NoRollback {
				return messages, err
			} else {
				// Rollback changes
				s.Log.Error(fmt.Errorf("an error occurred - rolling back changes"))
				resetErr := s.ResetStack(verbose)

				var finalErr error

				if resetErr != nil {
					finalErr = fmt.Errorf("%s - error resetting stack: %s", err.Error(), resetErr.Error())
				} else {
					finalErr = fmt.Errorf("%s - all changes rolled back", err.Error())
				}

				return messages, finalErr
			}
		}
	} else {
		err = s.runStartupSequence(s.Stack.RuntimeDir, verbose, false)
		if err != nil {
			return messages, err
		}
	}
	return messages, s.ensureFireflyNodesUp(true)
}

func (s *StackManager) PullStack(verbose bool, options *types.PullOptions) error {
	var images []string
	manifestImages := make(map[string]bool)

	// Collect FireFly docker image names
	for _, entry := range s.Stack.VersionManifest.Entries() {
		fullImage := entry.GetDockerImageString()
		s.Log.Info(fmt.Sprintf("Manifest entry image='%s' local=%t", fullImage, entry.Local))
		manifestImages[fullImage] = true
		if entry.Local {
			continue
		}
		images = append(images, fullImage)
	}

	// IPFS is the only one that we always use in every stack
	images = append(images, constants.IPFSImageName)

	// Also pull postgres if we're using it
	if s.Stack.Database == types.PostgreSQL.String() {
		images = append(images, constants.PostgresImageName)
	}

	// Also pull the Sandbox if we're using it
	if s.Stack.SandboxEnabled {
		images = append(images, constants.SandboxImageName)
	}

	// Also pull the FFTM if we're using it
	// if s.Stack.FFTMEnabled {
	// 	images = append(images, constants.FFTMImageName)
	// }

	// Iterate over all images used by the blockchain provider
	for _, service := range s.blockchainProvider.GetDockerServiceDefinitions() {
		if !manifestImages[service.Service.Image] {
			images = append(images, service.Service.Image)
		}
	}

	// Iterate over all images used by the tokens provider
	for iTok, tp := range s.tokenProviders {
		for _, service := range tp.GetDockerServiceDefinitions(iTok) {
			if !manifestImages[service.Service.Image] {
				images = append(images, service.Service.Image)
			}
		}
	}

	// Use docker to pull every image - retry on failure
	for _, image := range images {
		s.Log.Info(fmt.Sprintf("pulling '%s'", image))
		if err := docker.RunDockerCommandRetry(s.Stack.InitDir, verbose, verbose, options.Retries, "pull", image); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) removeVolumes(verbose bool) {
	var volumes []string
	for _, service := range s.blockchainProvider.GetDockerServiceDefinitions() {
		volumes = append(volumes, service.VolumeNames...)
	}
	for iTok, tp := range s.tokenProviders {
		for _, service := range tp.GetDockerServiceDefinitions(iTok) {
			volumes = append(volumes, service.VolumeNames...)
		}
	}
	for volumeName := range docker.CreateDockerCompose(s.Stack).Volumes {
		volumes = append(volumes, volumeName)
	}
	for _, volumeName := range volumes {
		docker.RunDockerCommand("", verbose, verbose, "volume", "remove", fmt.Sprintf("%s_%s", s.Stack.Name, volumeName))
	}
}

func (s *StackManager) runStartupSequence(workingDir string, verbose bool, firstTimeSetup bool) error {
	if err := s.blockchainProvider.PreStart(); err != nil {
		return err
	}

	s.Log.Info("starting FireFly dependencies")
	if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "-p", s.Stack.Name, "up", "-d"); err != nil {
		return err
	}

	if err := s.blockchainProvider.PostStart(); err != nil {
		return err
	}

	return nil
}

func (s *StackManager) StopStack(verbose bool) error {
	return docker.RunDockerComposeCommand(s.Stack.InitDir, verbose, verbose, "-p", s.Stack.Name, "stop")
}

func (s *StackManager) ResetStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(s.Stack.InitDir, verbose, verbose, "-p", s.Stack.Name, "down"); err != nil {
		return err
	}
	if err := os.RemoveAll(s.Stack.RuntimeDir); err != nil {
		return err
	}
	if err := s.blockchainProvider.Reset(); err != nil {
		return err
	}
	s.removeVolumes(verbose)
	return nil
}

func (s *StackManager) RemoveStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(s.Stack.InitDir, verbose, verbose, "-p", s.Stack.Name, "down"); err != nil {
		return err
	}
	s.removeVolumes(verbose)
	return os.RemoveAll(s.Stack.StackDir)
}

func (s *StackManager) checkPortsAvailable() error {
	ports := make([]int, 1)
	ports[0] = s.Stack.ExposedBlockchainPort
	for _, member := range s.Stack.Members {
		ports = append(ports, member.ExposedDataexchangePort)
		ports = append(ports, member.ExposedConnectorPort)
		if !member.External {
			ports = append(ports, member.ExposedFireflyAdminPort)
			ports = append(ports, member.ExposedFireflyPort)
			ports = append(ports, member.ExposedFireflyMetricsPort)
		}
		ports = append(ports, member.ExposedIPFSApiPort)
		ports = append(ports, member.ExposedIPFSGWPort)
		ports = append(ports, member.ExposedDatabasePort)
		ports = append(ports, member.ExposedUIPort)
		ports = append(ports, member.ExposedTokensPorts...)

		if s.Stack.SandboxEnabled {
			ports = append(ports, member.ExposedSandboxPort)
		}
		if s.Stack.FFTMEnabled {
			ports = append(ports, member.ExposedFFTMPort)
		}
	}

	if s.Stack.PrometheusEnabled {
		ports = append(ports, s.Stack.ExposedPrometheusPort)
	}

	for _, port := range ports {
		available, err := checkPortAvailable(port)
		if err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("port %d is unavailable. please check to see if another process is listening on that port", port)
		}
	}
	return nil
}

func checkPortAvailable(port int) (bool, error) {
	timeout := time.Millisecond * 500
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)), timeout)

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true, nil
	}

	switch t := err.(type) {

	case *net.OpError:
		switch t := t.Unwrap().(type) {
		case *os.SyscallError:
			if t.Syscall == "connect" {
				return true, nil
			}
		}
		if t.Op == "dial" {
			return false, err
		} else if t.Op == "read" {
			return true, nil
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return true, nil
		}
	}

	if conn != nil {
		defer conn.Close()
		return false, nil
	}
	return true, nil
}

func (s *StackManager) copyFireflyConfigToContainer(verbose bool, workingDir string, member *types.Member) error {
	if !member.External {
		s.Log.Info(fmt.Sprintf("copying firefly.core to firefly_core_%s", member.ID))
		volumeName := fmt.Sprintf("%s_firefly_core_%s", s.Stack.Name, member.ID)
		if err := docker.CopyFileToVolume(volumeName, filepath.Join(workingDir, fmt.Sprintf("firefly_core_%s.yml", member.ID)), "/firefly.core.yml", verbose); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) copyFFTMConfigToContainer(verbose bool, workingDir string, member *types.Member) error {
	if s.Stack.FFTMEnabled {
		s.Log.Info(fmt.Sprintf("copying firefly.fftm to fftm_%s", member.ID))
		volumeName := fmt.Sprintf("%s_fftm_%s", s.Stack.Name, member.ID)
		if err := docker.CopyFileToVolume(volumeName, filepath.Join(workingDir, fmt.Sprintf("firefly_fftm_%s.yml", member.ID)), "/firefly.fftm", verbose); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) runFirstTimeSetup(verbose bool, options *types.StartOptions) (messages []string, err error) {
	configDir := filepath.Join(s.Stack.RuntimeDir, "config")

	for i := 0; i < len(s.Stack.Members); i++ {
		if s.Stack.Members[i].Account != nil {
			s.Stack.State.Accounts = append(s.Stack.State.Accounts, s.Stack.Members[i].Account)
		}
	}

	if err := copy.Copy(s.Stack.InitDir, s.Stack.RuntimeDir); err != nil {
		return messages, err
	}

	if err := s.disableFireflyCoreContainers(verbose, s.Stack.RuntimeDir); err != nil {
		return messages, err
	}

	s.Log.Info("initializing blockchain node")
	if err := s.blockchainProvider.FirstTimeSetup(); err != nil {
		return messages, err
	}

	if s.Stack.PrometheusEnabled {
		s.Log.Info("copying prometheus.yml to prometheus_config")
		volumeName := fmt.Sprintf("%s_prometheus_config", s.Stack.Name)
		if err := docker.CopyFileToVolume(volumeName, path.Join(configDir, "prometheus.yml"), "/prometheus.yml", verbose); err != nil {
			return messages, err
		}
	}

	if err := s.copyDataExchangeConfigToVolumes(verbose); err != nil {
		return messages, err
	}

	pullOptions := &types.PullOptions{
		Retries: 2,
	}
	if err := s.PullStack(verbose, pullOptions); err != nil {
		return messages, err
	}

	if err := s.runStartupSequence(s.Stack.RuntimeDir, verbose, true); err != nil {
		return messages, err
	}

	for i, tp := range s.tokenProviders {
		result, err := tp.DeploySmartContracts(i)
		if err != nil {
			return messages, err
		}
		if result != nil {
			if result.Message != "" {
				messages = append(messages, result.Message)
			}
			s.Stack.State.DeployedContracts = append(s.Stack.State.DeployedContracts, result.DeployedContract)
		}
	}

	if s.Stack.ContractAddress == "" {
		s.Log.Info("deploying FireFly smart contracts")
		blockchainConfig, result, err := s.blockchainProvider.DeployFireFlyContract()
		if err != nil {
			return messages, err
		}
		if result != nil {
			if result.Message != "" {
				messages = append(messages, result.Message)
			}
			s.Stack.State.DeployedContracts = append(s.Stack.State.DeployedContracts, result.DeployedContract)
		}
		newConfig := &core.FireflyConfig{
			Blockchain: blockchainConfig,
		}
		s.patchFireFlyCoreConfigs(verbose, configDir, newConfig)
	}

	for _, member := range s.Stack.Members {
		if err := s.copyFireflyConfigToContainer(verbose, configDir, member); err != nil {
			return messages, err
		}
		if err := s.copyFFTMConfigToContainer(verbose, configDir, member); err != nil {
			return messages, err
		}
	}

	if err := s.enableFireflyCoreContainers(verbose, s.Stack.RuntimeDir); err != nil {
		return messages, err
	}

	// Bring up the FireFly Core containers now that we've finalized the runtime config
	docker.RunDockerComposeCommand(s.Stack.InitDir, verbose, verbose, "-p", s.Stack.Name, "up", "-d")

	if err := s.ensureFireflyNodesUp(true); err != nil {
		return messages, err
	}

	s.Log.Info("registering FireFly identities")
	if err := s.registerFireflyIdentities(verbose); err != nil {
		return messages, err
	}

	s.Log.Info("initializing token providers")
	for iTok, tp := range s.tokenProviders {
		if err := tp.FirstTimeSetup(iTok); err != nil {
			return messages, err
		}
	}

	// Update the stack state with any new state that was created as a part of the setup process
	return messages, s.writeStackStateJSON(s.Stack.RuntimeDir)
}

func (s *StackManager) ensureFireflyNodesUp(firstTimeSetup bool) error {
	for _, member := range s.Stack.Members {
		if member.External {
			configFilename := filepath.Join(s.Stack.RuntimeDir, "config", fmt.Sprintf("firefly_core_%v.yml", member.ID))
			var port int
			if firstTimeSetup {
				port = member.ExposedFireflyAdminPort
			} else {
				port = member.ExposedFireflyPort
			}
			// Check process running
			available, err := checkPortAvailable(port)
			if err != nil {
				return err
			}
			if available {
				s.Log.Info(fmt.Sprintf("please start your firefly core with the config file for this stack: firefly -f %s  ", configFilename))
				if err := s.waitForFireflyStart(port); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *StackManager) waitForFireflyStart(port int) error {
	retries := 120
	retryPeriod := 1000 // ms
	retriesRemaining := retries
	for retriesRemaining > 0 {
		time.Sleep(time.Duration(retryPeriod) * time.Millisecond)
		available, err := checkPortAvailable(port)
		if err != nil {
			return err
		}
		if !available {
			return nil
		}
		retriesRemaining--
	}
	return fmt.Errorf("waited for %v seconds for firefly to start on port %v but it was never available", retries*retryPeriod/1000, port)
}

func (s *StackManager) UpgradeStack(verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "-p", s.Stack.Name, "down"); err != nil {
		return err
	}
	return docker.RunDockerComposeCommand(workingDir, verbose, verbose, "pull")
}

func (s *StackManager) PrintStackInfo(verbose bool) error {
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(s.Stack.InitDir, verbose, true, "-p", s.Stack.Name, "images"); err != nil {
		return err
	}
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(s.Stack.RuntimeDir, verbose, true, "-p", s.Stack.Name, "ps"); err != nil {
		return err
	}
	fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(s.Stack.InitDir, "docker-compose.yml"))
	return nil
}

func (s *StackManager) disableFireflyCoreContainers(verbose bool, workingDir string) error {
	d, err := ioutil.ReadFile(filepath.Join(workingDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	var dockerComposeYAML *docker.DockerComposeConfig
	if err := yaml.Unmarshal(d, &dockerComposeYAML); err != nil {
		return err
	}

	// Cache any currently set entrypoint
	s.fireflyCoreEntrypoints = make([][]string, len(s.Stack.Members))
	for _, member := range s.Stack.Members {
		if !member.External {
			s.fireflyCoreEntrypoints[*member.Index] = dockerComposeYAML.Services[fmt.Sprintf("firefly_core_%v", *member.Index)].EntryPoint
			// Temporarily set the entrypoint to not run anything
			dockerComposeYAML.Services[fmt.Sprintf("firefly_core_%v", *member.Index)].EntryPoint = []string{"/bin/sh", "-c", "exit", "0"}
		}
	}
	return s.writeDockerCompose(workingDir, dockerComposeYAML)
}

func (s *StackManager) enableFireflyCoreContainers(verbose bool, workingDir string) error {
	d, err := ioutil.ReadFile(filepath.Join(workingDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	var dockerComposeYAML *docker.DockerComposeConfig
	if err := yaml.Unmarshal(d, &dockerComposeYAML); err != nil {
		return err
	}

	for _, member := range s.Stack.Members {
		if !member.External {
			// If there was a cached entrypoint for this service, restore it
			// Otherwise default to the entrypoint of "firefly" as defined in the Dockerfile for FireFly Core
			entrypoint := []string{"firefly"}
			if len(s.fireflyCoreEntrypoints[*member.Index]) > 0 {
				entrypoint = s.fireflyCoreEntrypoints[*member.Index]
			}
			dockerComposeYAML.Services[fmt.Sprintf("firefly_core_%v", *member.Index)].EntryPoint = entrypoint
		}
	}
	return s.writeDockerCompose(workingDir, dockerComposeYAML)
}

func (s *StackManager) patchFireFlyCoreConfigs(verbose bool, workingDir string, newConfig *core.FireflyConfig) error {
	if newConfig != nil {
		newConfigBytes, err := yaml.Marshal(newConfig)
		if err != nil {
			return err
		}
		for _, member := range s.Stack.Members {
			s.Log.Debug(fmt.Sprintf("patching config for %s: %v", member.ID, newConfig))
			configFile := path.Join(workingDir, fmt.Sprintf("firefly_core_%s.yml", member.ID))
			merger := conflate.New()
			if err := merger.AddFiles(configFile); err != nil {
				return fmt.Errorf("failed merging config %s", configFile)
			}
			if err := merger.AddData(newConfigBytes); err != nil {
				return fmt.Errorf("failed merging YAML '%v' into config: %s", newConfig, err)
			}
			s.Log.Info(fmt.Sprintf("updating %s config for new smart contract address", member.ID))
			configData, err := merger.MarshalYAML()
			if err != nil {
				return err
			}
			if err = ioutil.WriteFile(configFile, configData, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *StackManager) StackHasRunBefore() (bool, error) {
	if s.IsOldFileStructure {
		dataDir := filepath.Join(constants.StacksDir, s.Stack.Name, "data")
		_, err := os.Stat(dataDir)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		runtimeDir := filepath.Join(s.Stack.RuntimeDir)
		_, err := os.Stat(runtimeDir)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

func (s *StackManager) GetContracts(filename string, extraArgs []string) ([]string, error) {
	return s.blockchainProvider.GetContracts(filename, extraArgs)
}

func (s *StackManager) DeployContract(filename, contractName string, memberIndex int, extraArgs []string) (string, error) {
	result, err := s.blockchainProvider.DeployContract(filename, contractName, s.Stack.Members[memberIndex], extraArgs)
	if err != nil {
		return "", err
	}
	// Update the stackState.json file with the newly deployed contract
	deployedContract := &types.DeployedContract{
		Name:     contractName,
		Location: result.DeployedContract.Location,
	}
	s.Stack.State.DeployedContracts = append(s.Stack.State.DeployedContracts, deployedContract)
	if err = s.writeStackStateJSON(s.Stack.RuntimeDir); err != nil {
		return "", err
	}

	// Serialize the contract location to JSON to print on the command line
	b, err := json.MarshalIndent(result.DeployedContract.Location, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *StackManager) CreateAccount(args []string) (string, error) {
	newAccount, err := s.blockchainProvider.CreateAccount(args)
	if err != nil {
		return "", err
	}
	s.Stack.State.Accounts = append(s.Stack.State.Accounts, newAccount)
	if err = s.writeStackStateJSON(s.Stack.RuntimeDir); err != nil {
		return "", err
	}

	// Serialize the account to JSON to print on the command line
	b, err := json.MarshalIndent(newAccount, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *StackManager) getBlockchainProvider(verbose bool) blockchain.IBlockchainProvider {

	if s.Stack.BlockchainProvider == types.GoEthereum.String() {
		s.Stack.BlockchainProvider = types.Ethereum.String()
		s.Stack.BlockchainNodeProvider = types.GoEthereum.String()
	}

	if s.Stack.BlockchainProvider == types.HyperledgerBesu.String() {
		s.Stack.BlockchainProvider = types.Ethereum.String()
		s.Stack.BlockchainNodeProvider = types.HyperledgerBesu.String()
	}

	switch s.Stack.BlockchainProvider {
	case types.Ethereum.String():
		switch s.Stack.BlockchainNodeProvider {
		case types.GoEthereum.String():
			return &geth.GethProvider{
				Verbose: verbose,
				Log:     s.Log,
				Stack:   s.Stack,
			}
		case types.HyperledgerBesu.String():
			return &besu.BesuProvider{
				Verbose: verbose,
				Log:     s.Log,
				Stack:   s.Stack,
				Signer: &ethsigner.EthSignerProvider{
					Verbose: verbose,
					Log:     s.Log,
					Stack:   s.Stack,
				},
			}
		case types.RemoteRPC.String():
			return &remoterpc.RemoteRPCProvider{
				Verbose: verbose,
				Log:     s.Log,
				Stack:   s.Stack,
				Signer: &ethsigner.EthSignerProvider{
					Verbose: verbose,
					Log:     s.Log,
					Stack:   s.Stack,
				},
			}
		default:
			return nil
		}
	case types.HyperledgerFabric.String():
		return &fabric.FabricProvider{
			Verbose: verbose,
			Log:     s.Log,
			Stack:   s.Stack,
		}
	default:
		return nil
	}
}

func (s *StackManager) getITokenProviders(verbose bool) []tokens.ITokensProvider {
	tps := make([]tokens.ITokensProvider, len(s.Stack.TokenProviders))
	for i, tp := range s.Stack.TokenProviders {
		switch tp {
		case types.ERC1155:
			tps[i] = &erc1155.ERC1155Provider{
				Verbose: verbose,
				Log:     s.Log,
				Stack:   s.Stack,
			}
		case types.ERC20_ERC721:
			tps[i] = &erc20erc721.ERC20ERC721Provider{
				Verbose: verbose,
				Log:     s.Log,
				Stack:   s.Stack,
			}
		default:
			return nil
		}
	}
	return tps
}
