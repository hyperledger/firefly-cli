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
	"bufio"
	"fmt"
	"os/exec"
	"path"

	"github.com/kaleido-io/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View log output from a stack",
	Long: `View log output from a stack.

The most recent logs can be viewed, or you can follow the
output with the -f flag.`,
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

		fmt.Println("Getting logs...")
		follow, _ := cmd.Flags().GetBool("follow")
		stdoutChan := make(chan string)
		go runScript(stackName, follow, stdoutChan)
		for s := range stdoutChan {
			fmt.Print(s)
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}

func runScript(stackName string, follow bool, stdoutChan chan string) {
	stackDir := path.Join(stacks.StacksDir, stackName)
	cmd := exec.Command("docker", "compose", "--ansi", "always", "logs")
	if follow {
		cmd.Args = append(cmd.Args, "-f")
	}
	cmd.Dir = stackDir
	// stdin := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	buf := bufio.NewReader(stdout)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			close(stdoutChan)
			break
		} else {
			stdoutChan <- line
		}

	}
}
