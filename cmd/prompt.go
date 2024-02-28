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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

func prompt(promptText string, validate func(string) error) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(promptText)
		if str, err := reader.ReadString('\n'); err != nil {
			return "", err
		} else {
			str = strings.TrimSpace(str)
			if err := validate(str); err != nil {
				printError(err)
			} else {
				return str, nil
			}
		}
	}
}

func confirm(promptText string) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N] ", promptText)
		if str, err := reader.ReadString('\n'); err != nil {
			return err
		} else {
			str = strings.ToLower(strings.TrimSpace(str))
			if str == "y" || str == "yes" {
				return nil
			} else {
				return fmt.Errorf("confirmation declined with response: '%s'", str)
			}
		}
	}
}

func selectMenu(promptText string, options []string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n")
		for i, option := range options {
			fmt.Printf("  %v) %s\n", i+1, option)
		}
		fmt.Printf("\n%s: ", promptText)
		if str, err := reader.ReadString('\n'); err != nil {
			return "", err
		} else {
			str = strings.TrimSpace(str)
			index, err := strconv.Atoi(str)
			if err != nil {
				printError(fmt.Errorf("'%s' is not a valid option", str))
				continue
			}
			if index < 1 || index > len(options) {
				printError(fmt.Errorf("'%s' is not a valid option", str))
				continue
			}
			return options[index-1], nil
		}
	}
}

func printError(err error) {
	if fancyFeatures {
		fmt.Printf("\u001b[31mError: %s\u001b[0m\n", err.Error())
	} else {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

// listStacks aids in completion, to provide completion to command for stack name.
func listStacks(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	allStacks, err := stacks.ListStacks()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return allStacks, cobra.ShellCompDirectiveNoSpace
}
