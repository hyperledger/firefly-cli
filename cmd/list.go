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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hyperledger/firefly-cli/internal/stacks"
)

var listCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list stacks",
	Long:    `List stacks`,
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if stacks, err := stacks.ListStacks(); err != nil {
			return err
		} else {
			fmt.Print("FireFly Stacks:\n\n")
			for _, s := range stacks {
				fmt.Println(s)
			}
			fmt.Print("\n")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCommand)
}
