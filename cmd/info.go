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

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [a stack name]...",
	Short: "Get info about a stack",
	Long: `Get info about a stack such as each container name
	and image version. If non is given, it run the "info" command for all stack on the local machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = context.WithValue(ctx, docker.CtxIsLogCmdKey{}, true)
		ctx = log.WithLogger(ctx, logger)

		version, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)

		allStacks, err := stacks.ListStacks()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			namedStacks := make([]string, 0, len(args))
			for _, stackName := range args {
				if contains(allStacks, strings.TrimSpace(stackName)) {
					namedStacks = append(namedStacks, stackName)
				} else {
					fmt.Printf("stack name - %s, is not present on your local machine. Run `ff ls` to see all available stacks.\n", stackName)
				}
			}

			allStacks = namedStacks // replace only the user specified stacks in the slice instead.
		}

		stackManager := stacks.NewStackManager(ctx)
		for _, stackName := range allStacks {
			if err := stackManager.LoadStack(stackName); err != nil {
				return err
			}

			if err := stackManager.PrintStacksInfo(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
