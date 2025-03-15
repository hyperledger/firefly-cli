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

package cardanosigner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blinklabs-io/bursa"
	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type CardanoSignerProvider struct {
	ctx   context.Context
	stack *types.Stack
}

func NewCardanoSignerProvider(ctx context.Context, stack *types.Stack) *CardanoSignerProvider {
	return &CardanoSignerProvider{
		ctx:   ctx,
		stack: stack,
	}
}

func (p *CardanoSignerProvider) WriteConfig(_ *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	signerConfigPath := filepath.Join(initDir, "config", "cardanosigner.yaml")

	if err := GenerateSignerConfig().WriteConfig(signerConfigPath); err != nil {
		return nil
	}

	return nil
}

func (p *CardanoSignerProvider) FirstTimeSetup() error {
	cardanosignerVolumeName := fmt.Sprintf("%s_cardanosigner", p.stack.Name)
	blockchainDir := filepath.Join(p.stack.RuntimeDir, "blockchain", "keystore")

	if err := docker.CreateVolume(p.ctx, cardanosignerVolumeName); err != nil {
		return err
	}

	// Copy the signer config to the volume
	signerConfigPath := filepath.Join(p.stack.StackDir, "runtime", "config", "cardanosigner.yaml")
	signerConfigVolumeName := fmt.Sprintf("%s_cardanosigner_config", p.stack.Name)
	if err := docker.CopyFileToVolume(p.ctx, signerConfigVolumeName, signerConfigPath, "cardanosigner.yaml"); err != nil {
		return err
	}

	// Copy the members wallets to the volume
	if err := docker.MkdirInVolume(p.ctx, cardanosignerVolumeName, "wallet"); err != nil {
		return err
	}

	for _, member := range p.stack.Members {
		account := member.Account.(*cardano.Account)
		filename := filepath.Join(blockchainDir, fmt.Sprintf("%s.skey", account.Address))
		if err := docker.CopyFileToVolume(p.ctx, cardanosignerVolumeName, filename, fmt.Sprintf("wallet/%s.skey", account.Address)); err != nil {
			return err
		}
	}

	return nil
}

func (p *CardanoSignerProvider) GetDockerServiceDefinition(rpcURL string) *docker.ServiceDefinition {
	return &docker.ServiceDefinition{
		ServiceName: "cardanosigner",
		Service: &docker.Service{
			Image:         p.stack.VersionManifest.Cardanosigner.GetDockerImageString(),
			ContainerName: fmt.Sprintf("%s_cardanosigner", p.stack.Name),
			User:          "root",
			Command:       "./firefly-cardanosigner -f /etc/config/cardanosigner.yaml",
			Volumes: []string{
				"cardanosigner:/data",
				"cardanosigner_config:/etc/config",
			},
			Logging: docker.StandardLogOptions,
			Ports: []string{
				fmt.Sprintf("%d:8555", p.stack.ExposedBlockchainPort),
				"9583:9583",
			},
			Environment: p.stack.EnvironmentVars,
		},
		VolumeNames: []string{
			"cardanosigner",
			"cardanosigner_config",
		},
	}
}

func (p *CardanoSignerProvider) CreateAccount(args []string) (interface{}, error) {
	cardanosignerVolumeName := fmt.Sprintf("%s_cardanosigner", p.stack.Name)
	network := p.stack.Network
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
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return nil, err
	}

	mnemonic, err := bursa.NewMnemonic()
	if err != nil {
		return nil, err
	}
	wallet, err := bursa.NewWallet(mnemonic, network, 0, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	contents, err := json.Marshal(wallet.PaymentExtendedSKey)
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(outputDirectory, fmt.Sprintf("%s.skey", wallet.PaymentAddress))
	if err := os.WriteFile(filename, contents, 0755); err != nil {
		return nil, err
	}

	if stackHasRunBefore {
		// Copy the signer secret to the volume
		if err := docker.CopyFileToVolume(p.ctx, cardanosignerVolumeName, filename, fmt.Sprintf("wallet/%s.skey", wallet.PaymentAddress)); err != nil {
			return nil, err
		}
	}

	return &cardano.Account{
		Address:    wallet.PaymentAddress,
		PrivateKey: wallet.PaymentExtendedSKey.CborHex,
	}, nil
}
