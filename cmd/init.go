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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/castillobgr/sententia"
	"github.com/hyperledger-labs/firefly-cli/internal/stacks"
)

var initOptions stacks.InitOptions

var initCmd = &cobra.Command{
	Use:   "init [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack",
	Long:  `Create a new FireFly local dev stack`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("initializing new FireFly stack...")

		defaultStackName, _ := sententia.Make("{{ adjective }}-{{ nouns }}")
		stackName := defaultStackName

		if len(args) > 0 {
			stackName = args[0]
			err := validateName(stackName)
			if err != nil {
				return err
			}
		} else {
			stackName, _ = prompt("stack name: ", validateName)
			fmt.Println("You selected " + stackName)
		}

		var memberCountInput string
		if len(args) > 1 {
			memberCountInput = args[1]
			if err := validateCount(memberCountInput); err != nil {
				return err
			}
		} else {
			memberCountInput, _ = prompt("number of members: ", validateCount)
		}
		memberCount, _ := strconv.Atoi(memberCountInput)

		if err := stacks.InitStack(stackName, memberCount, &initOptions); err != nil {
			return err
		}

		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n", stackName, rootCmd.Use, stackName)
		fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(stacks.StacksDir, stackName, "docker-compose.yml"))
		return nil
	},
}

func validateName(stackName string) error {
	if strings.TrimSpace(stackName) == "" {
		return errors.New("stack name must not be empty")
	}
	if exists, err := stacks.CheckExists(stackName); exists {
		return fmt.Errorf("stack '%s' already exists", stackName)
	} else {
		return err
	}
}

func validateCount(input string) error {
	if i, err := strconv.Atoi(input); err != nil {
		return errors.New("invalid number")
	} else if i <= 0 {
		return errors.New("number of members must be greater than zero")
	}
	return nil
}

func init() {
	initCmd.Flags().IntVarP(&initOptions.FireFlyBasePort, "firefly-base-port", "p", 5000, "Mapped port base of FireFly core API (1 added for each member)")
	initCmd.Flags().IntVarP(&initOptions.ServicesBasePort, "services-base-port", "s", 5100, "Mapped port base of services (100 added for each member)")

	rootCmd.AddCommand(initCmd)
}
