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
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func contractName(tokenIndex int) string {
	return fmt.Sprintf("erc20erc721_TokenFactory_%d", tokenIndex)
}

func DeployContracts(ctx context.Context, s *types.Stack, blockchainProvider blockchain.IBlockchainProvider, tokenIndex int) (*types.ContractDeploymentResult, error) {
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

	if err := ethereum.ExtractContracts(ctx, containerName, "/home/node/contracts", s.RuntimeDir); err != nil {
		return nil, err
	}

	return blockchainProvider.DeployContract(filepath.Join(s.RuntimeDir, "contracts", "TokenFactory.json"), "TokenFactory", s.Members[0], nil)
}
