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

package geth

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/docker"
)

func (p *GethProvider) writeAccountToDisk(directory, address, privateKey string) error {
	outputDirectory := filepath.Join(directory, "blockchain", "accounts", address[2:])
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return err
	}
	filename := filepath.Join(outputDirectory, "keyfile")
	// Drop the 0x on the front of the private key here because that's what geth is expecting in the keyfile
	return ioutil.WriteFile(filename, []byte(privateKey[2:]), 0755)
}

func (p *GethProvider) importAccountToGeth(address string) error {
	address = address[2:]
	gethVolumeName := fmt.Sprintf("%s_geth", p.Stack.Name)
	blockchainDir := filepath.Join(p.Stack.RuntimeDir, "blockchain")
	if err := docker.RunDockerCommand(p.Stack.StackDir, p.Verbose, p.Verbose,
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/blockchain", blockchainDir),
		"-v", fmt.Sprintf("%s:/data", gethVolumeName),
		gethImage,
		"account",
		"import",
		"--password", "/blockchain/password",
		"--keystore", "/data/keystore",
		fmt.Sprintf("/blockchain/accounts/%s/keyfile", address),
	); err != nil {
		return err
	}
	return nil
}
