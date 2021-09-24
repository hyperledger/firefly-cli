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

	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <stack_name>",
	Short: "Get info about a stack",
	Long: `Get info about a stack such as each container name
	and image version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stackManager := stacks.NewStackManager(logger)
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]
		if exists, err := stacks.CheckExists(stackName); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("stack '%s' does not exist", stackName)
		}

		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		if err := stackManager.PrintStackInfo(verbose); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
