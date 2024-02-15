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
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:               "remove <stack_name>",
	Aliases:           []string{"rm"},
	Short:             "Completely remove a stack",
	ValidArgsFunction: listStacks,
	Long: `Completely remove a stack

This command will completely delete a stack, including all of its data
and configuration.`,
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

		if !force {
			fmt.Println("WARNING: This will completely remove your stack and all of its data. Are you sure this is what you want to do?")
			if err := confirm(fmt.Sprintf("completely delete FireFly stack '%s'", stackName)); err != nil {
				cancel()
			}
		}

		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		fmt.Printf("deleting FireFly stack '%s'... ", stackName)
		if err := stackManager.StopStack(); err != nil {
			return err
		}
		if err := stackManager.RemoveStack(); err != nil {
			return err
		}
		os.RemoveAll(filepath.Join(constants.StacksDir, stackName))
		fmt.Println("done")
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the stack without prompting for confirmation")
	rootCmd.AddCommand(removeCmd)
}
