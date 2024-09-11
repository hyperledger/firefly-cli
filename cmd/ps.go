// Copyright © 2024 Giwa Oluwatobi <giwaoluwatobi@gmial.com>
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

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps [a stack name]...",
	Short: "Returns information on running stacks",
	Long: `ps returns currently running stacks on your local machine.
	
	It also takes a continuous list of whitespace optional argument - stack name.`,
	Aliases: []string{"process"},
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

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
					fmt.Printf("stack name - %s, is not present on your local machine. Run `%s ls` to see all available stacks.\n", stackName, ExecutableName)
				}
			}

			allStacks = namedStacks // replace only the user specified stacks in the slice instead.
		}

		stackManager := stacks.NewStackManager(ctx)
		for _, stackName := range allStacks {
			if err := stackManager.LoadStack(stackName); err != nil {
				return err
			}

			if err := stackManager.IsRunning(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}

// contains can be removed if the go mod version is bumped to version Go 1.18
// and replaced with slices.Contains().
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
