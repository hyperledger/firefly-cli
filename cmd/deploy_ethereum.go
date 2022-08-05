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
	Use:   "ethereum <stack_name> <contract_json_file>",
	Short: "Deploy a compiled solidity contract",
	Long: `Deploy a solidity contract compiled with solc to the blockchain used by a FireFly stack

To compile a .sol file to a .json file run:

solc --combined-json abi,bin contract.sol > contract.json
`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDockerConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)
		stackName := args[0]
		filename := args[1]
		stackManager := stacks.NewStackManager(ctx)
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
			return err
		}
		fmt.Print(location)
		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployEthereumCmd)
}
