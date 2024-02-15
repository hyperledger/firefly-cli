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

var resetCmd = &cobra.Command{
	Use:               "reset <stack_name>",
	Short:             "Clear all data in a stack",
	ValidArgsFunction: listStacks,
	Long: `Clear all data in a stack

This command clears all data in a stack, but leaves the stack configuration.
This is useful for testing when you want to start with a clean slate
but don't want to actually recreate the resources in the stack itself.
Note: this will also stop the stack if it is running.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		version, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)

		stackManager := stacks.NewStackManager(ctx)
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]

		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}

		if stackManager.IsOldFileStructure {
			return fmt.Errorf("the FireFly stack '%s' was created with an older version of the CLI and resetting the stack is not supported. If you want to start fresh, please remove and recreate the stack", stackName)
		}

		if !force {
			fmt.Println("WARNING: This will completely remove all transactions and data from your FireFly stack. Are you sure you want to do that?")
			if err := confirm(fmt.Sprintf("reset all data in FireFly stack '%s'", stackName)); err != nil {
				cancel()
			}
		}

		fmt.Printf("resetting FireFly stack '%s'... ", stackName)
		if err := stackManager.StopStack(); err != nil {
			return err
		}
		if err := stackManager.ResetStack(); err != nil {
			return err
		}
		fmt.Printf("done\n\nYour stack has been reset. To start your stack run:\n\n%s start %s\n\n", rootCmd.Use, stackName)

		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&force, "force", "f", false, "Reset the stack without prompting for confirmation")
	rootCmd.AddCommand(resetCmd)
}
