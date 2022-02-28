// Copyright © 2021 Kaleido, Inc.
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
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/stacks"
)

var initOptions stacks.InitOptions
var databaseSelection string
var blockchainProviderInput string
var tokenProvidersSelection []string
var promptNames bool

var ffNameValidator = regexp.MustCompile(`^[0-9a-zA-Z]([0-9a-zA-Z._-]{0,62}[0-9a-zA-Z])?$`)

var initCmd = &cobra.Command{
	Use:   "init [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack",
	Long:  `Create a new FireFly local dev stack`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var stackName string
		stackManager := stacks.NewStackManager(logger)

		if err := validateDatabaseProvider(databaseSelection); err != nil {
			return err
		}
		if err := validateBlockchainProvider(blockchainProviderInput); err != nil {
			return err
		}
		if err := validateTokensProvider(tokenProvidersSelection); err != nil {
			return err
		}

		fmt.Println("initializing new FireFly stack...")

		if len(args) > 0 {
			stackName = args[0]
			err := validateStackName(stackName)
			if err != nil {
				return err
			}
		} else {
			stackName, _ = prompt("stack name: ", validateStackName)
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

		initOptions.OrgNames = make([]string, 0, memberCount)
		initOptions.NodeNames = make([]string, 0, memberCount)
		if promptNames {
			for i := 0; i < memberCount; i++ {
				name, _ := prompt(fmt.Sprintf("name for org %d: ", i), validateFFName)
				initOptions.OrgNames = append(initOptions.OrgNames, name)
				name, _ = prompt(fmt.Sprintf("name for node %d: ", i), validateFFName)
				initOptions.NodeNames = append(initOptions.NodeNames, name)
			}
		} else {
			for i := 0; i < memberCount; i++ {
				initOptions.OrgNames = append(initOptions.OrgNames, fmt.Sprintf("org_%d", i))
				initOptions.NodeNames = append(initOptions.NodeNames, fmt.Sprintf("node_%d", i))
			}
		}

		initOptions.Verbose = verbose
		initOptions.BlockchainProvider, _ = stacks.BlockchainProviderFromString(blockchainProviderInput)
		initOptions.DatabaseSelection, _ = stacks.DatabaseSelectionFromString(databaseSelection)
		initOptions.TokenProviders, _ = stacks.TokenProvidersFromStrings(tokenProvidersSelection)

		if err := stackManager.InitStack(stackName, memberCount, &initOptions); err != nil {
			return err
		}

		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n", stackName, rootCmd.Use, stackName)
		fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(constants.StacksDir, stackName, "docker-compose.yml"))

		switch runtime.GOOS {
		case "windows", "darwin":
			if memberCount > 2 {
				fmt.Printf("\u001b[33mNOTE: If your Docker Desktop configuration is set to the default memory configuration (2 GB), you may need to increase this value. It is recommended to allocate 1 GB per member of your FireFly network.\u001b[0m\n\n")
			}
		}
		return nil
	},
}

func validateStackName(stackName string) error {
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
	} else if initOptions.ExternalProcesses >= i {
		return errors.New("number of external processes should not be equal to or greater than the number of members in the network - at least one FireFly core container must exist to be able to extrat and deploy smart contracts")
	}
	return nil
}

func validateFFName(input string) error {
	if !ffNameValidator.MatchString(input) {
		return fmt.Errorf("name must be 1-64 characters, including alphanumerics (a-zA-Z0-9), dot (.), dash (-) and underscore (_), and must start/end in an alphanumeric")
	}
	return nil
}

func validateDatabaseProvider(input string) error {
	_, err := stacks.DatabaseSelectionFromString(input)
	if err != nil {
		return err
	}
	return nil
}

func validateBlockchainProvider(input string) error {
	blockchainSelection, err := stacks.BlockchainProviderFromString(input)
	if err != nil {
		return err
	}

	if blockchainSelection == stacks.Corda {
		return errors.New("support for corda is coming soon")
	}

	// TODO: When we get tokens on Fabric this should change
	if blockchainSelection == stacks.HyperledgerFabric {
		tokenProvidersSelection = []string{}
	}

	return nil
}

func validateTokensProvider(input []string) error {
	_, err := stacks.TokenProvidersFromStrings(input)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	initCmd.Flags().IntVarP(&initOptions.FireFlyBasePort, "firefly-base-port", "p", 5000, "Mapped port base of FireFly core API (1 added for each member)")
	initCmd.Flags().IntVarP(&initOptions.ServicesBasePort, "services-base-port", "s", 5100, "Mapped port base of services (100 added for each member)")
	initCmd.Flags().StringVarP(&databaseSelection, "database", "d", "sqlite3", fmt.Sprintf("Database type to use. Options are: %v", stacks.DBSelectionStrings))
	initCmd.Flags().StringVarP(&blockchainProviderInput, "blockchain-provider", "b", "geth", fmt.Sprintf("Blockchain provider to use. Options are: %v", stacks.BlockchainProviderStrings))
	initCmd.Flags().StringArrayVarP(&tokenProvidersSelection, "token-providers", "t", []string{"erc1155"}, fmt.Sprintf("Token providers to use. Options are: %v", stacks.ValidTokenProviders))
	initCmd.Flags().IntVarP(&initOptions.ExternalProcesses, "external", "e", 0, "Manage a number of FireFly core processes outside of the docker-compose stack - useful for development and debugging")
	initCmd.Flags().StringVarP(&initOptions.FireFlyVersion, "release", "r", "latest", "Select the FireFly release version to use")
	initCmd.Flags().StringVarP(&initOptions.ManifestPath, "manifest", "m", "", "Path to a manifest.json file containing the versions of each FireFly microservice to use. Overrides the --release flag.")
	initCmd.Flags().BoolVar(&promptNames, "prompt-names", false, "Prompt for org and node names instead of using the defaults")
	initCmd.Flags().BoolVar(&initOptions.PrometheusEnabled, "prometheus-enabled", false, "Enables Prometheus metrics exposition and aggregation to a shared Prometheus server")
	initCmd.Flags().IntVar(&initOptions.PrometheusPort, "prometheus-port", 9090, "Port for the shared Prometheus server")

	rootCmd.AddCommand(initCmd)
}
