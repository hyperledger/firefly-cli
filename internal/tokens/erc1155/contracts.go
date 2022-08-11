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

package erc1155

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

// TODO:
//
// REMOVE THIS FILE ONCE ERC-1155 SUPPORTS DEPLOYING WITH
// ETHCONNECT AND EVMCONNECT THE SAME WAY
//

const TOKEN_URI_PATTERN = "firefly://token/{id}"

func DeployContracts(ctx context.Context, s *types.Stack, tokenIndex int) (*types.ContractDeploymentResult, error) {
	l := log.LoggerFromContext(ctx)
	var containerName string
	for _, member := range s.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_tokens_%s_%d", s.Name, member.ID, tokenIndex)
			break
		}
	}
	if containerName == "" {
		return nil, errors.New("unable to extract contracts from container - no valid tokens containers found in stack")
	}
	l.Info("extracting smart contracts")

	if err := ethereum.ExtractContracts(ctx, containerName, "/root/contracts", s.RuntimeDir); err != nil {
		return nil, err
	}

	contracts, err := ethereum.ReadTruffleCompiledContract(filepath.Join(s.RuntimeDir, "contracts", "ERC1155MixedFungible.json"))
	if err != nil {
		return nil, err
	}
	tokenContract, ok := contracts.Contracts["ERC1155MixedFungible"]
	if !ok {
		return nil, fmt.Errorf("unable to find ERC1155MixedFungible in compiled contracts")
	}

	var tokenContractAddress string
	for _, member := range s.Members {
		// TODO: move to address based contract deployment, once ERC-1155 connector is updated to not require an EthConnect REST API registration
		if tokenContractAddress == "" {
			l.Info(fmt.Sprintf("deploying ERC1155 contract on '%s'", member.ID))
			tokenContractAddress, err = ethconnect.DeprecatedDeployContract(member, tokenContract, "erc1155", map[string]string{"uri": TOKEN_URI_PATTERN})
			if err != nil {
				return nil, err
			}
		} else {
			l.Info(fmt.Sprintf("registering ERC1155 contract on '%s'", member.ID))
			err = ethconnect.DeprecatedRegisterContract(member, tokenContract, tokenContractAddress, "erc1155", map[string]string{"uri": TOKEN_URI_PATTERN})
			if err != nil {
				return nil, err
			}
		}
	}

	result := &types.ContractDeploymentResult{
		Message: fmt.Sprintf("Deployed ERC-1155 contract to: %s", tokenContractAddress),
		DeployedContract: &types.DeployedContract{
			Name:     "erc1155",
			Location: map[string]string{"address": tokenContractAddress},
		},
	}

	return result, nil
}
