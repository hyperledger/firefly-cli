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

package ethconnect

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethtypes"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type Ethconnect struct {
	ctx context.Context
}

type PublishAbiResponseBody struct {
	ID string `json:"id,omitempty"`
}

type DeployContractResponseBody struct {
	ContractAddress string `json:"contractAddress,omitempty"`
}

type RegisterResponseBody struct {
	Created      string `json:"created,omitempty"`
	Address      string `json:"string,omitempty"`
	Path         string `json:"path,omitempty"`
	ABI          string `json:"ABI,omitempty"`
	OpenAPI      string `json:"openapi,omitempty"`
	RegisteredAs string `json:"registeredAs,omitempty"`
}

type EthconnectMessageRequest struct {
	Headers  EthconnectMessageHeaders `json:"headers,omitempty"`
	To       string                   `json:"to"`
	From     string                   `json:"from,omitempty"`
	ABI      interface{}              `json:"abi,omitempty"`
	Bytecode string                   `json:"compiled"`
	Params   []interface{}            `json:"params"`
}

type EthconnectMessageHeaders struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type EthconnectMessageResponse struct {
	Sent bool   `json:"sent,omitempty"`
	ID   string `json:"id,omitempty"`
}

type EthconnectReply struct {
	ID              string                  `json:"_id,omitempty"`
	Headers         *EthconnectReplyHeaders `json:"headers,omitempty"`
	ContractAddress string                  `json:"contractAddress,omitempty"`
	ErrorCode       string                  `json:"errorCode,omitempty"`
	ErrorMessage    string                  `json:"errorMessage,omitempty"`
}

type EthconnectReplyHeaders struct {
	ID            string  `json:"id,omitempty"`
	RequestID     string  `json:"requestId,omitempty"`
	RequestOffset string  `json:"requestOffset,omitempty"`
	TimeElapsed   float64 `json:"timeElapsed,omitempty"`
	TimeReceived  string  `json:"timeReceived,omitempty"`
	Type          string  `json:"type,omitempty"`
}

func NewEthconnect(ctx context.Context) *Ethconnect {
	return &Ethconnect{
		ctx: ctx,
	}
}

func (e *Ethconnect) Name() string {
	return "ethconnect"
}

func (e *Ethconnect) Port() int {
	return 8080
}

func (e *Ethconnect) DeployContract(contract *ethtypes.CompiledContract, contractName string, member *types.Organization, extraArgs []string) (*types.ContractDeploymentResult, error) {
	ethconnectURL := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	address := member.Account.(*ethereum.Account).Address
	hexBytecode, err := hex.DecodeString(strings.TrimPrefix(contract.Bytecode, "0x"))
	if err != nil {
		return nil, err
	}
	base64Bytecode := base64.StdEncoding.EncodeToString(hexBytecode)

	params := make([]interface{}, len(extraArgs))
	for i, arg := range extraArgs {
		params[i] = arg
	}

	requestBody := &EthconnectMessageRequest{
		Headers: EthconnectMessageHeaders{
			Type: "DeployContract",
		},
		From:     address,
		ABI:      contract.ABI,
		Bytecode: base64Bytecode,
		Params:   params,
	}

	ethconnectResponse := &EthconnectMessageResponse{}
	if err := core.RequestWithRetry(e.ctx, "POST", ethconnectURL, requestBody, ethconnectResponse); err != nil {
		return nil, err
	}
	reply, err := getReply(e.ctx, ethconnectURL, ethconnectResponse.ID)
	if err != nil {
		return nil, err
	}
	if reply.Headers.Type != "TransactionSuccess" {
		return nil, fmt.Errorf("%s", reply.ErrorMessage)
	}

	result := &types.ContractDeploymentResult{
		DeployedContract: &types.DeployedContract{
			Name:     contractName,
			Location: map[string]string{"address": reply.ContractAddress},
		},
	}
	return result, nil
}

func getReply(ctx context.Context, ethconnectURL, id string) (*EthconnectReply, error) {
	u, err := url.Parse(ethconnectURL)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("replies", id))
	if err != nil {
		return nil, err
	}
	requestURL := u.String()

	reply := &EthconnectReply{}
	err = core.RequestWithRetry(ctx, "GET", requestURL, nil, reply)
	return reply, err
}

func (e *Ethconnect) FirstTimeSetup(stack *types.Stack) error {
	for _, member := range stack.Members {
		if err := docker.MkdirInVolume(e.ctx, fmt.Sprintf("%s_ethconnect_data_%s", stack.Name, member.ID), "/abis"); err != nil {
			return err
		}
		if err := docker.MkdirInVolume(e.ctx, fmt.Sprintf("%s_ethconnect_data_%s", stack.Name, member.ID), "/events"); err != nil {
			return err
		}
	}
	return nil
}
