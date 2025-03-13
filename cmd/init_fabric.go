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
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var initFabricCmd = &cobra.Command{
	Use:   "fabric [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack using a Fabric network",
	Long:  `Create a new FireFly local dev stack using a Fabric network`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		version, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)

		stackManager := stacks.NewStackManager(ctx)
		initOptions.BlockchainProvider = types.BlockchainProviderFabric.String()
		initOptions.TokenProviders = []string{}
		if err := validateFabricFlags(); err != nil {
			return err
		}
		if err := initCommon(args); err != nil {
			return err
		}
		if err := stackManager.InitStack(&initOptions); err != nil {
			if err := stackManager.RemoveStack(); err != nil {
				return err
			}
			return err
		}
		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n", initOptions.StackName, rootCmd.Use, initOptions.StackName)
		fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(stackManager.Stack.StackDir, "docker-compose.yml"))
		return nil
	},
}

func validateFabricFlags() error {
	if len(initOptions.CCPYAMLPaths) != 0 || len(initOptions.MSPPaths) != 0 {
		if len(initOptions.CCPYAMLPaths) != len(initOptions.MSPPaths) {
			return fmt.Errorf("you must provide ccp and msp flags for each organization")
		}
		if initOptions.ChannelName == "" || initOptions.ChaincodeName == "" {
			return fmt.Errorf("channel and chaincode flags must be set when using an external fabric network")
		}
		initOptions.MemberCount = len(initOptions.CCPYAMLPaths)
	}
	return nil
}

func init() {
	initFabricCmd.Flags().StringArrayVar(&initOptions.CCPYAMLPaths, "ccp", nil, "Path to the ccp.yaml file for an org in your Fabric network")
	initFabricCmd.Flags().StringArrayVar(&initOptions.MSPPaths, "msp", nil, "Path to the MSP directory for an org in your Fabric network")
	initFabricCmd.Flags().StringVar(&initOptions.ChannelName, "channel", "", "The name of the Fabric channel on which the FireFly chaincode has been deployed")
	initFabricCmd.Flags().StringVar(&initOptions.ChaincodeName, "chaincode", "", "The name given to the FireFly chaincode when it was deployed")
	initFabricCmd.Flags().BoolVar(&initOptions.CustomPinSupport, "custom-pin-support", false, "Configure the blockchain listener to listen for BatchPin events from any chaincode on the channel")
	initCmd.AddCommand(initFabricCmd)
}
