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

package quorum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type QuorumClient struct {
	rpcURL string
}

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JSONRPCError `json:"error,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewQuorumClient(rpcURL string) *QuorumClient {
	return &QuorumClient{
		rpcURL: rpcURL,
	}
}

func (g *QuorumClient) UnlockAccount(address string, password string) error {
	requestBody, err := json.Marshal(&JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      0,
		Method:  "personal_unlockAccount",
		Params:  []interface{}{address, password, 0},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", g.rpcURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var rpcResponse *JSONRPCResponse
	err = json.Unmarshal(responseBody, &rpcResponse)
	if err != nil {
		return err
	}
	if rpcResponse.Error != nil {
		return fmt.Errorf("%s", rpcResponse.Error.Message)
	}
	return nil
}
