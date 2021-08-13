/*
Copyright © 2021 Kaleido, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/hyperledger-labs/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <stack_name>",
	Short: "Upgrade a stack",
	Long: `Upgrade a stack by pulling newer images.
	This operation will restart the stack if running.
	If certain containers were pinned to a specific image at init,
	this command will have no effect on those containers.`,
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
		fmt.Printf("upgrading stack '%s'... ", stackName)
		if err := stackManager.UpgradeStack(verbose); err != nil {
			return err
		}
		fmt.Printf("done\n\nYour stack has been upgraded. To start your upgraded stack run:\n\n%s start %s\n\n", rootCmd.Use, stackName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
