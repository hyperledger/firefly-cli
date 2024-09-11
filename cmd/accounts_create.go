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

// accountsCreateCmd represents the "accounts create" command
var accountsCreateCmd = &cobra.Command{
	Use:               "create <stack_name>",
	Short:             "Create a new account in the FireFly stack",
	Long:              `Create a new account in the FireFly stack`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: listStacks,
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
		stackManager := stacks.NewStackManager(cmd.Context())
		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		account, err := stackManager.CreateAccount(args[1:])
		if err != nil {
			return fmt.Errorf("%s. usage: %s accounts create <stack_name> <org_name> <account_name>", err.Error(), ExecutableName)
		}
		fmt.Print(account)
		fmt.Print("\n")
		return nil
	},
}

func init() {
	accountsCmd.AddCommand(accountsCreateCmd)
}
