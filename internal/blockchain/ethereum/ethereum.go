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
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"golang.org/x/crypto/sha3"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethtypes"
	"github.com/hyperledger/firefly-cli/internal/log"
)

type Account struct {
	Address      string `json:"address"`
	PrivateKey   string `json:"privateKey"`
	PtmPublicKey string `json:"ptmPublicKey"` // Public key used for Tessera
}

func GenerateAddressAndPrivateKey() (address string, privateKey string) {
	newPrivateKey, _ := secp256k1.NewPrivateKey()
	privateKeyBytes := newPrivateKey.Serialize()
	encodedPrivateKey := "0x" + hex.EncodeToString(privateKeyBytes)
	// Remove the "04" Suffix byte when computing the address. This byte indicates that it is an uncompressed public key.
	publicKeyBytes := newPrivateKey.PubKey().SerializeUncompressed()[1:]
	// Take the hash of the public key to generate the address
	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes)
	// Ethereum addresses only use the lower 20 bytes, so toss the rest away
	encodedAddress := "0x" + hex.EncodeToString(hash.Sum(nil)[12:32])

	return encodedAddress, encodedPrivateKey
}

func ReadFireFlyContract(ctx context.Context, s *types.Stack) (*ethtypes.CompiledContract, error) {
	log := log.LoggerFromContext(ctx)
	var containerName string
	for _, member := range s.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_firefly_core_%s", s.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return nil, errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	log.Info("extracting smart contracts")

	if err := ExtractContracts(ctx, containerName, "/firefly/contracts", s.RuntimeDir); err != nil {
		return nil, err
	}

	var fireflyContract *ethtypes.CompiledContract
	contracts, err := ReadContractJSON(filepath.Join(s.RuntimeDir, "contracts", "Firefly.json"))
	if err != nil {
		return nil, err
	}

	fireflyContract, ok := contracts.Contracts["Firefly.sol:Firefly"]
	if !ok {
		fireflyContract, ok = contracts.Contracts["FireFly"]
		if !ok {
			return nil, fmt.Errorf("unable to find compiled FireFly contract")
		}
	}

	return fireflyContract, nil
}
