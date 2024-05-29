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

package cardanosigner

import (
	"context"

	"github.com/blinklabs-io/bursa"
	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type CardanoSignerProvider struct {
	ctx   context.Context
	stack *types.Stack
}

func NewCardanoSignerProvider(ctx context.Context, stack *types.Stack) *CardanoSignerProvider {
	return &CardanoSignerProvider{
		ctx:   ctx,
		stack: stack,
	}
}

func (p *CardanoSignerProvider) CreateAccount(args []string) (interface{}, error) {
	var network string
	if len(args) >= 1 {
		network = args[0]
	} else {
		network = "mainnet"
	}

	mnemonic, err := bursa.NewMnemonic()
	if err != nil {
		return nil, err
	}
	wallet, err := bursa.NewWallet(mnemonic, network, 0, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	// TODO: probably persist this information
	return &cardano.Account{
		Address:    wallet.PaymentAddress,
		PrivateKey: wallet.PaymentSKey.CborHex,
	}, nil
}
