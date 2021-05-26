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

var follow bool
var ansi string

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <stack_name>",
	Short: "View log output from a stack",
	Long: `View log output from a stack.

The most recent logs can be viewed, or you can follow the
output with the -f flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]

		if exists, err := stacks.CheckExists(stackName); !exists {
			return fmt.Errorf("stack '%s' does not exist", stackName)
		} else if err != nil {
			return err
		}

		fmt.Println("Getting logs...")

		stdoutChan := make(chan string)
		go runScript(stackName, follow, ansi, stdoutChan)
		for s := range stdoutChan {
			fmt.Print(s)
		}
		return nil
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

	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	logsCmd.Flags().StringVarP(&ansi, "ansi", "a", "always", "control when to print ANSI control characters (\"never\"|\"always\"|\"auto\") (default \"always\")")
}

func runScript(stackName string, follow bool, ansi string, stdoutChan chan string) {
	stackDir := path.Join(stacks.StacksDir, stackName)
	cmd := exec.Command("docker", "compose", "--ansi", ansi, "logs")
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
