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

// deployFabricCmd represents the "deploy fabric" command
var deployFabricCmd = &cobra.Command{
	Use:               "fabric <stack_name> <chaincode_package> <channel> <chaincodeName> <version>",
	Short:             "Deploy fabric chaincode",
	ValidArgsFunction: listStacks,
	Long:              `Deploy a packaged chaincode to the Fabric network used by a FireFly stack`,
	Args:              cobra.ExactArgs(5),
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
		contractAddress, err := stackManager.DeployContract(filename, filename, 0, args[2:])
		if err != nil {
			return fmt.Errorf("%s. usage: %s deploy <stack_name> <filename> <channel> <chaincode> <version>", err.Error(), ExecutableName)
		}
		fmt.Print(contractAddress)
		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployFabricCmd)
}
