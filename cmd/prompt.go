// Copyright © 2022 Kaleido, Inc.
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
	"strings"
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
				if fancyFeatures {
					fmt.Printf("\u001b[31mError: %s\u001b[0m\n", err.Error())
				} else {
					fmt.Printf("Error: %s\n", err.Error())
				}
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
