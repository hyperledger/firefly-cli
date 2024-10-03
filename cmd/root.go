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
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/hyperledger/firefly-cli/internal/log"
)

var cfgFile string
var ansi string
var fancyFeatures bool
var verbose bool
var force bool
var logger log.Logger = &log.StdoutLogger{
	LogLevel: log.Debug,
}

// name of the executable, this is for the help messages
var ExecutableName string = os.Args[0]

func GetFireflyASCIIArt() string {
	s := ""
	s += "\u001b[33m    _______           ________     \u001b[0m\n"   // yellow
	s += "\u001b[33m   / ____(_)_______  / ____/ /_  __\u001b[0m\n"   // yellow
	s += "\u001b[31m  / /_  / / ___/ _ \\/ /_  / / / / /\u001b[0m\n"  // red
	s += "\u001b[31m / __/ / / /  /  __/ __/ / / /_/ / \u001b[0m\n"   // red
	s += "\u001b[35m/_/   /_/_/   \\___/_/   /_/\\__, /  \u001b[0m\n" // magenta
	s += "\u001b[35m                          /____/   \u001b[0m\n"   // magenta

	return s
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   ExecutableName,
	Short: "FireFly CLI is a developer tool used to manage local development stacks",
	Long: GetFireflyASCIIArt() + `
FireFly CLI is a developer tool used to manage local development stacks
	
This tool automates creation of stacks with many infrastructure components which
would otherwise be a time consuming manual task. It also wraps docker compose
commands to manage the lifecycle of stacks.

To get started run: ` + ExecutableName + ` init
Optional: Set FIREFLY_HOME env variable for FireFly stack configuration path.
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if ansi == "always" {
			fancyFeatures = true
		} else if ansi == "auto" && (isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
			fancyFeatures = true
		} else {
			fancyFeatures = false
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().StringVarP(&ansi, "ansi", "", "auto", "control when to print ANSI control characters (\"never\"|\"always\"|\"auto\")")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose log output")
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
}

func cancel() {
	fmt.Println("canceled")
	os.Exit(1)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".firefly-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".firefly-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		var e viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			fmt.Println(err.Error())
		}
	}

}
