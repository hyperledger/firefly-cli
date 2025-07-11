// Copyright © 2025 IOG Singapore and SundaeSwap, Inc.
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

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/spf13/cobra"
)

var initCardanoCmd = &cobra.Command{
	Use:   "cardano [stack_name]",
	Short: "Create a new FireFly local dev stack using a Cardano blockchain",
	Long:  "Create a new FireFly local dev stack using a Cardano blockchain",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)
		stackManager := stacks.NewStackManager(ctx)
		initOptions.BlockchainProvider = types.BlockchainProviderCardano.String()
		initOptions.BlockchainConnector = types.BlockchainConnectorCardanoConnect.String()
		initOptions.BlockchainNodeProvider = types.BlockchainNodeProviderRemoteRPC.String()
		initOptions.TokenProviders = []string{}
		initOptions.MultipartyEnabled = false
		if len(args) == 1 {
			// stacks are enforced to have 1 member
			args = append(args, "1")
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

func init() {
	initCardanoCmd.Flags().StringVar(&initOptions.Network, "network", "mainnet", "The name of the network to connect to")
	initCardanoCmd.Flags().StringVar(&initOptions.Socket, "socket", "", "Socket to mount for the cardano node to connect to")
	initCardanoCmd.Flags().StringVar(&initOptions.BlockfrostKey, "blockfrost-key", "", "Blockfrost key")
	initCardanoCmd.Flags().StringVar(&initOptions.BlockfrostBaseURL, "blockfrost-base-url", "", "Blockfrost base URL (for run-your-own blockfrost setups)")

	initCmd.AddCommand(initCardanoCmd)
}
