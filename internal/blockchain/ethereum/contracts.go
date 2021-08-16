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
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum/ethconnect"
	"github.com/hyperledger-labs/firefly-cli/internal/constants"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/internal/log"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

func DeployContracts(s *types.Stack, log log.Logger, verbose bool) error {
	var coreContainer string
	for _, member := range s.Members {
		if !member.External {
			coreContainer = fmt.Sprintf("%s_firefly_core_%s_1", s.Name, member.ID)
			break
		}
	}
	if coreContainer == "" {
		return errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	log.Info("extracting smart contracts")

	if err := extractContracts(s.Name, coreContainer, "/firefly/contracts", verbose); err != nil {
		return err
	}

	fireflyContract, err := readCompiledContract(filepath.Join(constants.StacksDir, s.Name, "contracts", "Firefly.json"))
	if err != nil {
		return err
	}

	var fireflyContractAddress string
	for _, member := range s.Members {
		if fireflyContractAddress == "" {
			// TODO: version the registered name
			log.Info(fmt.Sprintf("deploying firefly contract on '%s'", member.ID))
			fireflyContractAddress, err = deployContract(member, fireflyContract, "firefly", map[string]string{})
			if err != nil {
				return err
			}
		} else {
			log.Info(fmt.Sprintf("registering firefly contract on '%s'", member.ID))
			err = registerContract(member, fireflyContract, fireflyContractAddress, "firefly", map[string]string{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readCompiledContract(filePath string) (*types.Contract, error) {
	d, _ := ioutil.ReadFile(filePath)
	var contract *types.Contract
	err := json.Unmarshal(d, &contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func extractContracts(stackName string, containerName string, dirName string, verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, stackName)
	if err := docker.RunDockerCommand(workingDir, verbose, verbose, "cp", containerName+":"+dirName, workingDir); err != nil {
		return err
	}
	return nil
}

func deployContract(member *types.Member, contract *types.Contract, name string, args map[string]string) (string, error) {
	ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	abiResponse, err := ethconnect.PublishABI(ethconnectUrl, contract)
	if err != nil {
		return "", err
	}
	deployResponse, err := ethconnect.DeployContract(ethconnectUrl, abiResponse.ID, member.Address, args, name)
	if err != nil {
		return "", err
	}
	return deployResponse.ContractAddress, nil
}

func registerContract(member *types.Member, contract *types.Contract, contractAddress string, name string, args map[string]string) error {
	ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
	abiResponse, err := ethconnect.PublishABI(ethconnectUrl, contract)
	if err != nil {
		return err
	}
	_, err = ethconnect.RegisterContract(ethconnectUrl, abiResponse.ID, contractAddress, member.Address, name, args)
	if err != nil {
		return err
	}
	return nil
}
