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
	Use:   "init [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack",
	Long:  `Create a new FireFly local dev stack`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("initializing new FireFly stack...")

		validateName := func(stackName string) error {
			if exists, err := stacks.CheckExists(stackName); exists {
				return fmt.Errorf("stack '%s' already exists", stackName)
			} else {
				return err
			}
		}

		var stackName string
		if len(args) > 0 {
			stackName = args[0]
			err := validateName(stackName)
			if err != nil {
				return err
			}
		} else {
			defaultStackName, _ := sententia.Make("{{ adjective }}-{{ nouns }}")
			prompt := promptui.Prompt{
				Label:    "stack name",
				Default:  defaultStackName,
				Validate: validateName,
			}
			stackName, _ = prompt.Run()
		}

		validateCount := func(input string) error {
			if i, err := strconv.Atoi(input); err != nil {
				return errors.New("invalid number")
			} else if i <= 0 {
				return errors.New("number of members must be greater than zero")
			}
			return nil
		}

		var memberCountInput string
		if len(args) > 1 {
			memberCountInput = args[1]
			if err := validateCount(memberCountInput); err != nil {
				return err
			}
		} else {
			prompt := promptui.Prompt{
				Label:    "number of members",
				Default:  "2",
				Validate: validateCount,
			}
			memberCountInput, _ = prompt.Run()
		}
		memberCount, _ := strconv.Atoi(memberCountInput)

		stacks.InitStack(stackName, memberCount)

		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n\n", stackName, rootCmd.Use, stackName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
