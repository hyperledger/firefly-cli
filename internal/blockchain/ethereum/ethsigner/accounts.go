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
	"path/filepath"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/docker"
)

func (p *EthSignerProvider) writeTomlKeyFile(directory, address string) error {
	outputDirectory := filepath.Join(directory, "blockchain", "keystore")
	address = strings.TrimPrefix(address, "0x")
	toml := fmt.Sprintf(`[metadata]
createdAt = 2019-11-05T08:15:30-05:00
description = "File based configuration"

[signing]
type = "file-based-signer"
key-file = "/data/keystore/%s"
password-file = "/data/password"
`, address)
	filename := filepath.Join(outputDirectory, fmt.Sprintf("%s.toml", address))
	return ioutil.WriteFile(filename, []byte(toml), 0755)
}

func (p *EthSignerProvider) copyTomlFileToVolume(directory, address, volumeName string, verbose bool) error {
	address = strings.TrimPrefix(address, "0x")
	filename := filepath.Join(directory, fmt.Sprintf("%s.toml", address))
	if err := docker.MkdirInVolume(volumeName, "/keystore", verbose); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(volumeName, filename, "/keystore", verbose); err != nil {
		return err
	}
	return nil
}
