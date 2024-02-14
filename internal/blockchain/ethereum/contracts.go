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

package ethereum

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/ethtypes"
	"github.com/hyperledger/firefly-cli/internal/docker"
)

type truffleCompiledContract struct {
	ABI          interface{} `json:"abi"`
	Bytecode     string      `json:"bytecode"`
	ContractName string      `json:"contractName"`
}

func ReadTruffleCompiledContract(filePath string) (*ethtypes.CompiledContracts, error) {
	d, _ := os.ReadFile(filePath)
	var truffleCompiledContract *truffleCompiledContract
	err := json.Unmarshal(d, &truffleCompiledContract)
	if err != nil {
		return nil, err
	}
	contract := &ethtypes.CompiledContract{
		ABI:      truffleCompiledContract.ABI,
		Bytecode: truffleCompiledContract.Bytecode,
	}
	contracts := &ethtypes.CompiledContracts{
		Contracts: map[string]*ethtypes.CompiledContract{
			truffleCompiledContract.ContractName: contract,
		},
	}
	return contracts, nil
}

func ReadSolcCompiledContract(filePath string) (*ethtypes.CompiledContracts, error) {
	d, _ := os.ReadFile(filePath)
	var contracts *ethtypes.CompiledContracts
	err := json.Unmarshal(d, &contracts)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func ReadContractJSON(filePath string) (*ethtypes.CompiledContracts, error) {
	contracts, err := ReadSolcCompiledContract(filePath)
	if err != nil {
		return nil, err
	}
	if len(contracts.Contracts) > 0 {
		return contracts, nil
	}
	return ReadTruffleCompiledContract(filePath)
}

func ExtractContracts(ctx context.Context, containerName, sourceDir, destinationDir string) error {
	if err := docker.RunDockerCommand(ctx, destinationDir, "cp", containerName+":"+sourceDir, destinationDir); err != nil {
		return err
	}
	return nil
}
