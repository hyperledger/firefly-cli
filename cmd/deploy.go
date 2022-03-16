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
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy <stack_name> <filename> [extra_args]",
	Short: "Deploy a compiled smart contract to the blockchain used by a FireFly stack",
	Long:  `Deploy a compiled smart contract to the blockchain used by a FireFly stack`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := docker.CheckDockerConfig(); err != nil {
			return err
		}

		stackManager := stacks.NewStackManager(logger)
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]
		if len(args) < 2 {
			return fmt.Errorf("no filename specified")
		}
		filename := args[1]

		if exists, err := stacks.CheckExists(stackName); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("stack '%s' does not exist", stackName)
		}

		if err := stackManager.LoadStack(stackName, verbose); err != nil {
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

		fmt.Printf("deploying %s... ", selectedContractName)
		contractAddress, err := stackManager.DeployContract(filename, selectedContractName, 0, args[2:])
		if err != nil {
			return err
		}

		fmt.Print("done\n\n")
		fmt.Printf("contract address: %s\n", contractAddress)
		fmt.Print("\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
