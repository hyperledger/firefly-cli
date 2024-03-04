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

package evmconnect

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethtypes"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type EvmconnectRequest struct {
	Headers    EvmconnectHeaders `json:"headers,omitempty"`
	To         string            `json:"to"`
	From       string            `json:"from,omitempty"`
	Definition interface{}       `json:"definition,omitempty"`
	Contract   string            `json:"contract,omitempty"`
	Params     []interface{}     `json:"params,omitempty"`
}

type EvmconnectHeaders struct {
	Type string `json:"type,omitempty"`
}

type EvmconnectTransactionResponse struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	Receipt *Receipt `json:"receipt"`
}

type Receipt struct {
	ExtraInfo *ExtraInfo `json:"extraInfo,omitempty"`
}

type ExtraInfo struct {
	ContractAddress string `json:"contractAddress,omitempty"`
}

type Evmconnect struct {
	ctx context.Context
}

func NewEvmconnect(ctx context.Context) *Evmconnect {
	return &Evmconnect{
		ctx: ctx,
	}
}

func (e *Evmconnect) Name() string {
	return "evmconnect"
}

func (e *Evmconnect) Port() int {
	return 5008
}

func (e *Evmconnect) DeployContract(contract *ethtypes.CompiledContract, contractName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	evmconnectURL := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	fromAddress := member.Account.(*ethereum.Account).Address

	params := make([]interface{}, len(extraArgs))
	for i, arg := range extraArgs {
		params[i] = arg
	}

	requestBody := &EvmconnectRequest{
		Headers: EvmconnectHeaders{
			Type: "DeployContract",
		},
		From:       fromAddress,
		Definition: contract.ABI,
		Contract:   contract.Bytecode,
		Params:     params,
	}

	txResponse := &EvmconnectTransactionResponse{}
	err := core.RequestWithRetry(e.ctx, "POST", evmconnectURL, requestBody, txResponse)
	if err != nil {
		return nil, err
	}

	txResponse, err = e.waitForTransactionSuccess(evmconnectURL, txResponse.ID)
	if err != nil ||
		txResponse == nil ||
		txResponse.Receipt == nil ||
		txResponse.Receipt.ExtraInfo == nil {
		return nil, err
	}

	result := &types.ContractDeploymentResult{
		DeployedContract: &types.DeployedContract{
			Name:     contractName,
			Location: map[string]string{"address": txResponse.Receipt.ExtraInfo.ContractAddress},
		},
	}
	return result, nil
}

func (e *Evmconnect) waitForTransactionSuccess(evmconnectURL, id string) (*EvmconnectTransactionResponse, error) {
	retries := 10
	for retries > 0 {
		tx, err := e.getTransactionStatus(evmconnectURL, id)
		if err != nil {
			return nil, err
		}
		if tx.Status == "Succeeded" {
			return tx, nil
		}
		time.Sleep(time.Millisecond * 3000)
	}
	return nil, nil
}

func (e *Evmconnect) getTransactionStatus(evmconnectURL, id string) (*EvmconnectTransactionResponse, error) {
	u, err := url.Parse(evmconnectURL)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("transactions", id))
	if err != nil {
		return nil, err
	}
	requestURL := u.String()

	reply := &EvmconnectTransactionResponse{}
	err = core.RequestWithRetry(e.ctx, "GET", requestURL, nil, reply)
	return reply, err
}

func (e *Evmconnect) FirstTimeSetup(stack *types.Stack) error {
	for _, member := range stack.Members {
		if err := docker.MkdirInVolume(e.ctx, fmt.Sprintf("%s_evmconnect_data_%s", stack.Name, member.ID), "/leveldb"); err != nil {
			return err
		}
	}
	return nil
}
