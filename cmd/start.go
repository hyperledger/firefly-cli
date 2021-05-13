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
	"io/ioutil"
	"os/exec"
	"path"
	"time"

	"github.com/briandowns/spinner"
	"github.com/kaleido-io/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a stack",
	Long: `Start a stack

This command will start a stack and run it in the background.
`,
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

		stackDir := path.Join(stacks.StacksDir, stackName)
		dockerCmd := exec.Command("docker", "compose", "up", "-d")
		dockerCmd.Dir = stackDir
		fmt.Printf("Starting FireFly stack '%s'... ", stackName)

		files, _ := ioutil.ReadDir(path.Join(stackDir, "postgres"))
		if len(files) == 0 {
			fmt.Println("\nThis will take a few seconds longer since this is the first time you're running this stack...")
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()
		err := dockerCmd.Run()
		s.Stop()
		if err != nil {
			fmt.Printf("command finished with error: %v", err)
		} else {
			// TODO: Print some useful information about URL and ports to use the stack
			fmt.Println("done!")
			fmt.Printf("To see logs for your stack run:\n\nfirefly-cli logs %s\n\n", stackName)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
