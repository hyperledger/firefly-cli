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
	"fmt"

	"github.com/hyperledger-labs/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop <stack_name>",
	Short: "Stop a stack",
	Long:  `Stop a stack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]
		if exists, err := stacks.CheckExists(stackName); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("stack '%s' does not exist", stackName)
		}

		if stack, err := stacks.LoadStack(stackName); err != nil {
			return err
		} else {
			fmt.Printf("stopping stack '%s'... ", stackName)
			if err := stack.StopStack(verbose); err != nil {
				return err
			}
			fmt.Print("done\n")
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
