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

package ethconnect

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

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

type CompiledContracts struct {
	Contracts map[string]*CompiledContract `json:"contracts"`
}

type CompiledContract struct {
	ABI      interface{} `json:"abi"`
	Bytecode string      `json:"bin"`
}

type EthconnectMessageRequest struct {
	Headers  EthconnectMessageHeaders `json:"headers,omitempty"`
	To       string                   `json:"to"`
	From     string                   `json:"from,omitempty"`
	ABI      interface{}              `json:"abi,omitempty"`
	Bytecode string                   `json:"compiled"`
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

func publishABI(ethconnectUrl string, contract *types.Contract) (*PublishAbiResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse("abis")
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	abi, err := json.Marshal(contract.ABI)
	if err != nil {
		return nil, err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormField("abi")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, bytes.NewReader(abi)); err != nil {
		return nil, err
	}
	fw, err = writer.CreateFormField("bytecode")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, strings.NewReader(contract.Bytecode)); err != nil {
		return nil, err
	}
	writer.Close()
	req, err := http.NewRequest("POST", requestUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var publishAbiResponse *PublishAbiResponseBody
	json.Unmarshal(responseBody, &publishAbiResponse)
	return publishAbiResponse, nil
}

func deployContract(ethconnectUrl string, abiId string, fromAddress string, params map[string]string, registeredName string) (*DeployContractResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("abis", abiId))
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-firefly-from", fromAddress)
	req.Header.Set("x-firefly-sync", "true")
	if registeredName != "" {
		req.Header.Set("x-firefly-register", registeredName)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var deployContractResponse *DeployContractResponseBody
	json.Unmarshal(responseBody, &deployContractResponse)
	return deployContractResponse, nil
}

func registerContract(ethconnectUrl string, abiId string, contractAddress string, fromAddress string, registeredName string, params map[string]string) (*RegisterResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("abis", abiId, contractAddress))
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-firefly-sync", "true")
	req.Header.Set("x-firefly-register", registeredName)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var registerResponseBody *RegisterResponseBody
	json.Unmarshal(responseBody, &registerResponseBody)
	return registerResponseBody, nil
}

// Deprecated
func DeployContract(member *types.Member, contract *types.Contract, name string, args map[string]string) (string, error) {
	ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	abiResponse, err := publishABI(ethconnectUrl, contract)
	if err != nil {
		return "", err
	}
	deployResponse, err := deployContract(ethconnectUrl, abiResponse.ID, member.Address, args, name)
	if err != nil {
		return "", err
	}
	return deployResponse.ContractAddress, nil
}

// Deprecated
func RegisterContract(member *types.Member, contract *types.Contract, contractAddress string, name string, args map[string]string) error {
	ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedConnectorPort)
	abiResponse, err := publishABI(ethconnectUrl, contract)
	if err != nil {
		return err
	}
	_, err = registerContract(ethconnectUrl, abiResponse.ID, contractAddress, member.Address, name, args)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated
func DeployContracts(s *types.Stack, log log.Logger, verbose bool) error {
	var containerName string
	for _, member := range s.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_firefly_core_%s", s.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	log.Info("extracting smart contracts")

	if err := ethereum.ExtractContracts(s.Name, containerName, "/firefly/contracts", verbose); err != nil {
		return err
	}

	fireflyContract, err := ethereum.ReadCompiledContract(filepath.Join(constants.StacksDir, s.Name, "contracts", "Firefly.json"))
	if err != nil {
		return err
	}

	for _, member := range s.Members {
		if s.ContractAddress == "" {
			log.Info(fmt.Sprintf("deploying firefly contract on '%s'", member.ID))
			s.ContractAddress, err = DeployContract(member, fireflyContract, "firefly", map[string]string{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DeployCustomContract(ethconnectUrl, fromAddress, filename, contractName string) (string, error) {
	contracts, err := ethereum.ReadCombinedABIJSON(filename)
	if err != nil {
		return "", nil
	}

	contract := contracts.Contracts[contractName]

	hexBytecode, err := hex.DecodeString(contract.Bytecode)
	if err != nil {
		return "", err
	}
	base64Bytecode := base64.StdEncoding.EncodeToString(hexBytecode)

	requestBody := &EthconnectMessageRequest{
		Headers: EthconnectMessageHeaders{
			Type: "DeployContract",
		},
		From:     fromAddress,
		ABI:      contract.ABI,
		Bytecode: base64Bytecode,
	}

	ethconnectResponse := &EthconnectMessageResponse{}
	if err := core.RequestWithRetry("POST", ethconnectUrl, requestBody, ethconnectResponse, false); err != nil {
		return "", err
	}
	reply, err := getReply(ethconnectUrl, ethconnectResponse.ID)
	if err != nil {
		return "", err
	}
	if reply.Headers.Type != "TransactionSuccess" {
		return "", fmt.Errorf(reply.ErrorMessage)
	}

	return reply.ContractAddress, nil
}

func getReply(ethconnectUrl, id string) (*EthconnectReply, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("replies", id))
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()

	reply := &EthconnectReply{}
	err = core.RequestWithRetry("GET", requestUrl, nil, reply, false)
	return reply, err
}
