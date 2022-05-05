// Copyright © 2021 Kaleido, Inc.
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

package erc20erc721

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func contractName(tokenIndex int) string {
	return fmt.Sprintf("erc20erc721_TokenFactory_%d", tokenIndex)
}

func DeployContracts(s *types.Stack, log log.Logger, verbose bool, tokenIndex int) (*types.ContractDeploymentResult, error) {
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
	log.Info("extracting smart contracts")

	if err := ethereum.ExtractContracts(containerName, "/home/node/contracts", s.RuntimeDir, verbose); err != nil {
		return nil, err
	}

	contractAddress, err := ethconnect.DeployCustomContract(s.Members[0], filepath.Join(s.RuntimeDir, "contracts", "TokenFactory.json"), "TokenFactory")
	if err != nil {
		return nil, err
	}

	result := &types.ContractDeploymentResult{
		Message: fmt.Sprintf("Deployed TokenFactory contract to: %s\nSource code for this contract can be found at %s", contractAddress, filepath.Join(s.RuntimeDir, "contracts", "source")),
		DeployedContract: &types.DeployedContract{
			Name:     contractName(tokenIndex),
			Location: map[string]string{"address": contractAddress},
		},
	}

	return result, nil
}
