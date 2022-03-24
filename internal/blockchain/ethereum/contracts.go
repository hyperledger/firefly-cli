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

package ethereum

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
)

type CompiledContracts struct {
	Contracts map[string]*CompiledContract `json:"contracts"`
}

type CompiledContract struct {
	ABI      interface{} `json:"abi"`
	Bytecode string      `json:"bin"`
}

type truffleCompiledContract struct {
	ABI      interface{} `json:"abi"`
	Bytecode string      `json:"bytecode"`
}

func ReadTruffleCompiledContract(filePath string) (*CompiledContract, error) {
	d, _ := ioutil.ReadFile(filePath)
	var truffleCompiledContract *truffleCompiledContract
	err := json.Unmarshal(d, &truffleCompiledContract)
	if err != nil {
		return nil, err
	}
	contract := &CompiledContract{
		ABI:      truffleCompiledContract.ABI,
		Bytecode: truffleCompiledContract.Bytecode,
	}
	return contract, nil
}

func ReadCombinedABIJSON(filePath string) (*CompiledContracts, error) {
	d, _ := ioutil.ReadFile(filePath)
	var contracts *CompiledContracts
	err := json.Unmarshal(d, &contracts)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func ExtractContracts(stackName string, containerName string, dirName string, verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, stackName)
	if err := docker.RunDockerCommand(workingDir, verbose, verbose, "cp", containerName+":"+dirName, workingDir); err != nil {
		return err
	}
	return nil
}
