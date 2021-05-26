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
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

func GetFireflyAsciiArt() string {
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
	Use:   "ff",
	Short: "FireFly CLI is a developer tool used to manage local development stacks",
	Long: GetFireflyAsciiArt() + `
FireFly CLI is a developer tool used to manage local development stacks
	
This tool automates creation of stacks with many infrastructure components which
would otherwise be a time consuming manual task. It also wraps docker compose
commands to manage the lifecycle of stacks.

To get started run: ff init
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
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
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
