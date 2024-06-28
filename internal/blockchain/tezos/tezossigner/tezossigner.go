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

package tezossigner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/tezos"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type TezosSignerProvider struct {
	ctx   context.Context
	stack *types.Stack
}

func NewTezosSignerProvider(ctx context.Context, stack *types.Stack) *TezosSignerProvider {
	return &TezosSignerProvider{
		ctx:   ctx,
		stack: stack,
	}
}

func (p *TezosSignerProvider) WriteConfig(_ *types.InitOptions) error {
	initDir := filepath.Join(constants.StacksDir, p.stack.Name, "init")
	signerConfigPath := filepath.Join(initDir, "config", "tezossigner.yaml")

	memberAccounts := p.getMembersAccounts()
	if err := GenerateSignerConfig(memberAccounts).WriteConfig(signerConfigPath); err != nil {
		return nil
	}

	return nil
}

func (p *TezosSignerProvider) getMembersAccounts() []string {
	accounts := make([]string, 0, len(p.stack.Members))
	for _, member := range p.stack.Members {
		if member.Account != nil {
			account := member.Account.(*tezos.Account)
			accounts = append(accounts, account.Address)
		}
	}
	return accounts
}

func (p *TezosSignerProvider) FirstTimeSetup() error {
	tezossignerVolumeName := fmt.Sprintf("%s_tezossigner", p.stack.Name)
	blockchainDir := filepath.Join(p.stack.RuntimeDir, "blockchain")

	if err := docker.CreateVolume(p.ctx, tezossignerVolumeName); err != nil {
		return err
	}

	// Copy the signer config to the volume
	signerConfigPath := filepath.Join(p.stack.StackDir, "runtime", "config", "tezossigner.yaml")
	signerConfigVolumeName := fmt.Sprintf("%s_tezossigner_config", p.stack.Name)
	if err := docker.CopyFileToVolume(p.ctx, signerConfigVolumeName, signerConfigPath, "signatory.yaml"); err != nil {
		return err
	}

	// Copy the members wallets to the volume
	if err := docker.CopyFileToVolume(p.ctx, signerConfigVolumeName, filepath.Join(blockchainDir, "keystore", "secret.json"), "secret.json"); err != nil {
		return err
	}

	return nil
}

func (p *TezosSignerProvider) GetDockerServiceDefinition(rpcURL string) *docker.ServiceDefinition {
	return &docker.ServiceDefinition{
		ServiceName: "tezossigner",
		Service: &docker.Service{
			Image:         "ecadlabs/signatory",
			ContainerName: fmt.Sprintf("%s_tezossigner", p.stack.Name),
			User:          "root",
			Command:       "-c /etc/signatory.yaml --base-dir /data",
			Volumes: []string{
				"tezossigner:/data",
				"tezossigner_config:/etc",
			},
			Logging: docker.StandardLogOptions,
			HealthCheck: &docker.HealthCheck{
				Test: []string{
					"CMD",
					"curl",
					"--fail",
					"http://localhost:9583/healthz",
				},
				Interval: "15s", // 6000 requests in a day
				Retries:  30,
			},
			Ports: []string{
				fmt.Sprintf("%d:6732", p.stack.ExposedBlockchainPort),
				"9583:9583",
			},
			Environment: p.stack.EnvironmentVars,
		},
		VolumeNames: []string{
			"tezossigner",
			"tezossigner_config",
		},
	}
}

func (p *TezosSignerProvider) CreateAccount(args []string) (interface{}, error) {
	tezossignerConfigVolumeName := fmt.Sprintf("%s_tezossigner_config", p.stack.Name)
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

	address, pk, err := tezos.GenerateAddressAndPrivateKey()
	if err != nil {
		return nil, err
	}

	// TODO: add support for several accounts at the same time
	json := fmt.Sprintf(`[
	{
		"name": "%s",
		"value": "unencrypted:%s"
	}
]`, address, pk)

	filename := filepath.Join(outputDirectory, "secret.json")
	if err := os.WriteFile(filename, []byte(json), 0755); err != nil {
		return nil, err
	}

	if stackHasRunBefore {
		// Copy the signer secret to the volume
		signerSecretPath := filepath.Join(outputDirectory, "secret.json")
		if err := docker.CopyFileToVolume(p.ctx, tezossignerConfigVolumeName, signerSecretPath, "secret.json"); err != nil {
			return nil, err
		}
	}

	return &tezos.Account{
		Address:    address,
		PrivateKey: pk,
	}, nil
}
