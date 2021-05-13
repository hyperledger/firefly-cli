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
	"os/exec"
	"path"

	"github.com/kaleido-io/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a stack",
	Long:  `Stop a stack`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No stack specified!")
			return
		}
		stackName := args[0]

		if !stacks.CheckExists(stackName) {
			fmt.Printf("Stack '%s' does not exist!", stackName)
			return
		}

		dockerCmd := exec.Command("docker", "compose", "stop")
		dockerCmd.Dir = path.Join(stacks.StacksDir, stackName)
		fmt.Printf("Stopping FireFly stack '%s'... ", stackName)
		err := dockerCmd.Run()
		if err != nil {
			fmt.Printf("command finished with error: %v", err)
		} else {
			// TODO: Print some useful information about URL and ports to use the stack
			fmt.Println("done!")
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
