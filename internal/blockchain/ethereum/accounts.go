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

package ethereum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-signer/pkg/keystorev3"
	"github.com/hyperledger/firefly-signer/pkg/secp256k1"
)

func CreateWalletFile(outputDirectory, prefix, password string) (*secp256k1.KeyPair, string, error) {
	keyPair, err := secp256k1.GenerateSecp256k1KeyPair()
	if err != nil {
		return nil, "", err
	}
	wallet := keystorev3.NewWalletFileStandard(password, keyPair)

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return nil, "", err
	}

	var filename string
	if prefix != "" {
		filename = filepath.Join(outputDirectory, fmt.Sprintf("%v_%s", prefix, keyPair.Address.String()[2:]))
	} else {
		filename = filepath.Join(outputDirectory, keyPair.Address.String()[2:])
	}
	err = os.WriteFile(filename, wallet.JSON(), 0755)
	if err != nil {
		return nil, "", err
	}
	return keyPair, filename, nil
}

func CopyWalletFileToVolume(ctx context.Context, walletFilePath, volumeName string) error {
	if err := docker.MkdirInVolume(ctx, volumeName, "/keystore"); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(ctx, volumeName, walletFilePath, "/keystore"); err != nil {
		return err
	}
	return nil
}
