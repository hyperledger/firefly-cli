// Copyright © 2022 Kaleido, Inc.
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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/docker"
)

func (p *EthSignerProvider) writeAccountToDisk(directory, address, privateKey string) error {
	outputDirectory := filepath.Join(directory, "blockchain", "accounts", address[2:])
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return err
	}
	filename := filepath.Join(outputDirectory, "keyfile")
	// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
	return ioutil.WriteFile(filename, []byte(privateKey[2:]), 0755)
}

func (p *EthSignerProvider) writeTomlKeyFile(directory, address string) error {
	outputDirectory := filepath.Join(directory, "blockchain", "accounts", address[2:])
	address = address[2:]
	toml := fmt.Sprintf(`[metadata]
createdAt = 2019-11-05T08:15:30-05:00
description = "File based configuration"

[signing]
type = "file-based-signer"
key-file = "/data/keystore/%s.key"
password-file = "/data/password"
`, address)
	filename := filepath.Join(outputDirectory, fmt.Sprintf("%s.toml", address))
	return ioutil.WriteFile(filename, []byte(toml), 0755)
}

func (p *EthSignerProvider) importAccountToEthsigner(address string) error {
	blockchainDir := filepath.Join(p.Stack.RuntimeDir, "blockchain")
	ethsignerVolumeName := fmt.Sprintf("%s_ethsigner", p.Stack.Name)
	address = address[2:]
	if err := docker.RunDockerCommand(p.Stack.RuntimeDir, p.Verbose, p.Verbose,
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/ethsigner", blockchainDir),
		"-v", fmt.Sprintf("%s:/data", ethsignerVolumeName),
		gethImage,
		"account",
		"import",
		"--password", "/ethsigner/password",
		"--keystore", "/data/keystore/output",
		fmt.Sprintf("/ethsigner/accounts/%s/keyfile", address),
	); err != nil {
		return err
	}

	// Move the file so we can reference it by name in the toml file and copy the toml file
	if err := docker.RunDockerCommand(p.Stack.RuntimeDir, p.Verbose, p.Verbose,
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/data", ethsignerVolumeName),
		"alpine",
		"/bin/sh",
		"-c",
		fmt.Sprintf("mv /data/keystore/output/*%s  /data/keystore/%s.key", address, address),
	); err != nil {
		return err
	}

	if err := docker.RunDockerCommand(p.Stack.RuntimeDir, p.Verbose, p.Verbose,
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/ethsigner", blockchainDir),
		"-v", fmt.Sprintf("%s:/data", ethsignerVolumeName),
		"alpine",
		"cp",
		fmt.Sprintf("/ethsigner/accounts/%s/%s.toml", address, address),
		fmt.Sprintf("/data/keystore/%s.toml", address),
	); err != nil {
		return err
	}
	return nil
}
