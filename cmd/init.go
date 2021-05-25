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
	"errors"
	"fmt"
	"strconv"

	"github.com/nguyer/promptui"
	"github.com/spf13/cobra"

	"github.com/castillobgr/sententia"
	"github.com/kaleido-io/firefly-cli/internal/stacks"
)

var initCmd = &cobra.Command{
	Use:   "init [stack_name]",
	Short: "Create a new FireFly local dev stack",
	Long:  `Create a new FireFly local dev stack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Initializing new FireFly stack...")

		var stackName string
		defaultStackName, _ := sententia.Make("{{ adjective }}-{{ nouns }}")

		if len(args) == 0 {
			prompt := promptui.Prompt{
				Label:   "Stack name",
				Default: defaultStackName,
				Validate: func(stackName string) error {
					if exists, err := stacks.CheckExists(stackName); exists {
						return fmt.Errorf("stack '%s' already exists", stackName)
					} else if err != nil {
						return err
					}
					return nil
				},
			}
			stackName, _ = prompt.Run()
		} else {
			stackName = args[0]
			if exists, err := stacks.CheckExists(stackName); exists {
				return fmt.Errorf("stack '%s' already exists", stackName)
			} else if err != nil {
				return err
			}
		}

		validate := func(input string) error {
			i, err := strconv.Atoi(input)
			if err != nil {
				return errors.New("invalid number")
			}
			if i <= 0 {
				return errors.New("number of members must be greater than zero")
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "Number of members",
			Default:  "2",
			Validate: validate,
		}
		memberCountInput, _ := prompt.Run()
		memberCount, _ := strconv.Atoi(memberCountInput)

		stacks.InitStack(stackName, memberCount)

		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\nfirefly-cli start %s\n\n", stackName, stackName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
