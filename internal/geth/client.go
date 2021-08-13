package geth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GethClient struct {
	rpcUrl string
}

type RpcRequest struct {
	JsonRPC string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

func NewGethClient(rpcUrl string) *GethClient {
	return &GethClient{
		rpcUrl: rpcUrl,
	}
}

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

func (g *GethClient) UnlockAccount(address string, password string) error {
	requestBody, err := json.Marshal(&RpcRequest{
		JsonRPC: "2.0",
		ID:      0,
		Method:  "personal_unlockAccount",
		Params:  []string{address, password},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", g.rpcUrl, bytes.NewBuffer(requestBody))
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
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%d %s", resp.StatusCode, responseBody)
	}
	return nil
}
