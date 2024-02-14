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
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/docker"
)

func (p *EthSignerProvider) writeTomlKeyFile(walletFilePath string) (string, error) {
	outputDirectory := filepath.Dir(walletFilePath)
	keyFile := filepath.Base(walletFilePath)
	toml := fmt.Sprintf(`[metadata]
createdAt = 2019-11-05T08:15:30-05:00
description = "File based configuration"

[signing]
type = "file-based-signer"
key-file = "/data/keystore/%s"
password-file = "/data/password"
`, keyFile)
	filename := filepath.Join(outputDirectory, fmt.Sprintf("%s.toml", keyFile))
	return filename, os.WriteFile(filename, []byte(toml), 0755)
}

func (p *EthSignerProvider) copyTomlFileToVolume(ctx context.Context, tomlFilePath, volumeName string) error {
	if err := docker.MkdirInVolume(ctx, volumeName, "/keystore"); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(ctx, volumeName, tomlFilePath, "/keystore"); err != nil {
		return err
	}
	return nil
}
