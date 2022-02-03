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
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func DeployContracts(s *types.Stack, log log.Logger, verbose bool, tokenIndex int) error {

	// Currently the act of creating and deploying a suitable ERC20 or ERC721 compliant
	// contract, or contract factory, is an exercise left to the user.
	//
	// For users simply experimenting with how tokens work, the ERC1155 standard is recommended
	// as a flexbile and fully formed sample implementation of fungible and non-fungible tokens
	// with a set of features you would expect.
	//
	// For users looking to take the next step and create a "proper" coin or NFT collection,
	// you really can't bypass the step of investigating the right OpenZeppelin (or other)
	// base class and determining the tokenomics (around supply / minting / burning / governance)
	// on top of that base class using the examples and standards out there.

	return nil
}
