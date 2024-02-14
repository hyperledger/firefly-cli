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
	"encoding/json"
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// accountsListCmd represents the "accounts list" command
var accountsListCmd = &cobra.Command{
	Use:               "list <stack_name>",
	Short:             "List the accounts in the FireFly stack",
	Long:              `List the accounts in the FireFly stack`,
	ValidArgsFunction: listStacks,
	Args:              cobra.ExactArgs(1),
	Aliases:           []string{"ls"},
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
		accounts, err := json.MarshalIndent(stackManager.Stack.State.Accounts, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(accounts))
		return nil
	},
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
}
