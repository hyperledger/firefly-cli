/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"log"
	"os"
	"path"
	"strings"

	"github.com/hyperledger/ff/internal/stacks"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No stack specified!")
		}
		stackName := args[0]

		if !stacks.CheckExists(stackName) {
			log.Fatalf("Stack '%s' does not exist!", stackName)
		}

		prompt := promptui.Prompt{
			Label:     "Reset all data in FireFly stack " + stackName,
			IsConfirm: true,
		}

		fmt.Println("WARNING: This will completely remove your all transactions and data from your FireFly stack. Are you sure you want to do that?")
		result, err := prompt.Run()

		if err != nil || strings.ToLower(result) != "y" {
			fmt.Printf("Canceled.")
			return
		} else {
			fmt.Printf("Reseting FireFly stack '%s'... ", stackName)
			os.RemoveAll(path.Join(stacks.FireflyDir, stackName, "data"))
			fmt.Println("done!")
		}
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
