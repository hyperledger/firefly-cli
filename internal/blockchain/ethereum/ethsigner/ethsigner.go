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

package ethsigner

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

// TODO: Probably randomize this and make it different per member?
var keyPassword = "correcthorsebatterystaple"

const useJavaSigner = false // also need to change the image appropriately if you recompile to use the Java signer

type EthSignerProvider struct {
	ctx   context.Context
	stack *types.Stack
}

func NewEthSignerProvider(ctx context.Context, stack *types.Stack) *EthSignerProvider {
	return &EthSignerProvider{
		ctx:   ctx,
		stack: stack,
	}
}

func (p *EthSignerProvider) WriteConfig(options *types.InitOptions, rpcURL string) error {

	// Write the password that will be used to encrypt the private key
	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	blockchainDirectory := filepath.Join(initDir, "blockchain")
	if err := os.MkdirAll(blockchainDirectory, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(initDir, "blockchain", "password"), []byte(keyPassword), 0755); err != nil {
		return err
	}

	signerConfigPath := filepath.Join(initDir, "config", "ethsigner.yaml")
	if err := GenerateSignerConfig(options.ChainID, rpcURL).WriteConfig(signerConfigPath); err != nil {
		return nil
	}

	return nil
}

func (p *EthSignerProvider) FirstTimeSetup() error {
	ethsignerVolumeName := fmt.Sprintf("%s_ethsigner", p.stack.Name)
	blockchainDir := filepath.Join(p.stack.RuntimeDir, "blockchain")
	contractsDir := filepath.Join(p.stack.RuntimeDir, "contracts")

	if err := docker.CreateVolume(p.ctx, ethsignerVolumeName); err != nil {
		return err
	}

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return err
	}

	// Copy the signer config to the volume
	signerConfigPath := filepath.Join(p.stack.StackDir, "runtime", "config", "ethsigner.yaml")
	signerConfigVolumeName := fmt.Sprintf("%s_ethsigner_config", p.stack.Name)
	if err := docker.CopyFileToVolume(p.ctx, signerConfigVolumeName, signerConfigPath, "firefly.ffsigner"); err != nil {
		return err
	}

	// Copy the wallet files all members to the blockchain volume
	if err := docker.CopyFileToVolume(p.ctx, ethsignerVolumeName, filepath.Join(blockchainDir, "keystore"), "/"); err != nil {
		return err
	}

	// Copy the password (to be used for decrypting private keys)
	if err := docker.CopyFileToVolume(p.ctx, ethsignerVolumeName, path.Join(blockchainDir, "password"), "password"); err != nil {
		return err
	}

	return nil
}

func (p *EthSignerProvider) getCommand(rpcURL string) string {
	if !useJavaSigner {
		return ""
	}

	// The Java based signing runtime if swapped in, requires these command line parameters
	u, err := url.Parse(rpcURL)
	if err != nil || rpcURL == "" {
		panic(fmt.Errorf("RPC URL invalid '%s': %s", rpcURL, err))
	}
	ethsignerCommand := []string{}
	ethsignerCommand = append(ethsignerCommand, "--logging=DEBUG")
	ethsignerCommand = append(ethsignerCommand, fmt.Sprintf(`--chain-id=%d`, p.stack.ChainID()))
	ethsignerCommand = append(ethsignerCommand, fmt.Sprintf(`--downstream-http-host=%s`, u.Hostname()))
	port := u.Port()
	if u.Scheme == "https" {
		ethsignerCommand = append(ethsignerCommand, `--downstream-http-tls-enabled`)
		if port == "" {
			port = "443"
		}
	}
	if u.Path != "" && u.Path != "/" {
		ethsignerCommand = append(ethsignerCommand, fmt.Sprintf(`--downstream-http-path=%s`, u.Path))
	}
	ethsignerCommand = append(ethsignerCommand, fmt.Sprintf(`--downstream-http-port=%s`, port))
	ethsignerCommand = append(ethsignerCommand, `multikey-signer`)
	ethsignerCommand = append(ethsignerCommand, `--directory=/data/keystore`)
	return strings.Join(ethsignerCommand, " ")
}

func (p *EthSignerProvider) GetDockerServiceDefinition(rpcURL string) *docker.ServiceDefinition {
	return &docker.ServiceDefinition{
		ServiceName: "ethsigner",
		Service: &docker.Service{
			Image:         p.stack.VersionManifest.Signer.GetDockerImageString(),
			ContainerName: fmt.Sprintf("%s_ethsigner", p.stack.Name),
			User:          "root",
			Command:       p.getCommand(rpcURL),
			Volumes: []string{
				"ethsigner:/data",
				"ethsigner_config:/etc/firefly",
			},
			Logging: docker.StandardLogOptions,
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD",
					"curl",
					"-X", "POST",
					"-H", "Content-Type: application/json",
					"-d", `{"jsonrpc":"2.0","method":"net_version","params":[],"id":"1"}`,
					"-w", "%{http_code}",
					"-sS",
					"--fail",
					"http://localhost:8545/",
				},
				Interval: "15s", // 6000 requests in a day
				Retries:  60,
			},
			Ports:       []string{fmt.Sprintf("%d:8545", p.stack.ExposedBlockchainPort)},
			Environment: p.stack.EnvironmentVars,
		},
		VolumeNames: []string{
			"ethsigner",
			"ethsigner_config",
		},
	}
}

func (p *EthSignerProvider) CreateAccount(args []string) (interface{}, error) {
	ethsignerVolumeName := fmt.Sprintf("%s_ethsigner", p.stack.Name)
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

	outputDirectory := filepath.Join(directory, "blockchain", "keystore")
	keyPair, walletFilePath, err := ethereum.CreateWalletFile(outputDirectory, "", keyPassword)
	if err != nil {
		return nil, err
	}

	tomlFilePath, err := p.writeTomlKeyFile(walletFilePath)
	if err != nil {
		return nil, err
	}

	if stackHasRunBefore {
		if err := ethereum.CopyWalletFileToVolume(p.ctx, walletFilePath, ethsignerVolumeName); err != nil {
			return nil, err
		}

		if err := p.copyTomlFileToVolume(p.ctx, tomlFilePath, ethsignerVolumeName); err != nil {
			return nil, err
		}

	}

	return &ethereum.Account{
		Address:    keyPair.Address.String(),
		PrivateKey: hex.EncodeToString(keyPair.PrivateKeyBytes()),
	}, nil
}
