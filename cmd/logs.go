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

var follow bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:               "logs <stack_name>",
	Short:             "View log output from a stack",
	ValidArgsFunction: listStacks,
	Long: `View log output from a stack.

The most recent logs can be viewed, or you can follow the
output with the -f flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = context.WithValue(ctx, docker.CtxIsLogCmdKey{}, true)
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

		stackHasRunBefore, err := stackManager.Stack.HasRunBefore()
		if err != nil {
			return err
		}

		if stackHasRunBefore {
			fmt.Println("getting logs... ")
			commandLine := []string{}
			if fancyFeatures {
				commandLine = append(commandLine, "--ansi", "always")
			}
			commandLine = append(commandLine, "-p", stackName, "logs")
			if follow {
				commandLine = append(commandLine, "-f")
			}
			if err := docker.RunDockerComposeCommand(ctx, stackManager.Stack.RuntimeDir, commandLine...); err != nil {
				return err
			}
		} else {
			fmt.Println("no logs found - stack has not been started")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
}
