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

package cmd

import (
	"context"
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// deployEthereumCmd represents the "deploy ethereum" command
var deployEthereumCmd = &cobra.Command{
	Use:               "ethereum <stack_name> <contract_json_file> [constructor_param1 [constructor_param2 ...]]",
	Short:             "Deploy a compiled solidity contract",
	ValidArgsFunction: listStacks,
	Long: `Deploy a solidity contract compiled with solc to the blockchain used by a FireFly stack. If the
contract has a constructor that takes arguments specify them as arguments to the command after the filename.

To compile a .sol file to a .json file run:

solc --combined-json abi,bin contract.sol > contract.json
`,
	Args: cobra.MinimumNArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		version, err := docker.CheckDockerConfig()
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)
		cmd.SetContext(ctx)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]
		filename := args[1]
		stackManager := stacks.NewStackManager(cmd.Context())
		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		contractNames, err := stackManager.GetContracts(filename, args[2:])
		if err != nil {
			return err
		}
		if len(contractNames) < 1 {
			return fmt.Errorf("no contracts found in file: '%s'", filename)
		}
		selectedContractName := contractNames[0]
		if len(contractNames) > 1 {
			selectedContractName, err = selectMenu("select the contract to deploy", contractNames)
			fmt.Print("\n")
			if err != nil {
				return err
			}
		}
		location, err := stackManager.DeployContract(filename, selectedContractName, 0, args[2:])
		if err != nil {
			return fmt.Errorf("%s. usage: %s deploy <stack_name> <filename> <channel> <chaincode> <version>", err.Error(), ExecutableName)
		}
		fmt.Print(location)
		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployEthereumCmd)
}
