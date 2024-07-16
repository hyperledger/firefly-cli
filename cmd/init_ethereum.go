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
	"github.com/hyperledger/firefly-common/pkg/fftypes"
)

var initEthereumCmd = &cobra.Command{
	Use:   "ethereum [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack using an Ethereum blockchain",
	Long:  `Create a new FireFly local dev stack using an Ethereum blockchain`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)
		version, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		// Needs this context for cleaning up as part of Remove Stack
		// If an error occurs as part of init
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)

		stackManager := stacks.NewStackManager(ctx)
		if err := initCommon(args); err != nil {
			return err
		}
		if err := stackManager.InitStack(&initOptions); err != nil {
			verr := stackManager.RemoveStack()

			// log the remove error if present
			if verr != nil {
				l := log.LoggerFromContext(ctx)
				l.Info(fmt.Sprintf("Error whilst removing the stack: %s", verr.Error()))
			}
			// return the init error to not hide the issue
			return err
		}
		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n", initOptions.StackName, rootCmd.Use, initOptions.StackName)
		fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(stackManager.Stack.StackDir, "docker-compose.yml"))
		return nil
	},
}

func init() {
	initEthereumCmd.Flags().IntVar(&initOptions.BlockPeriod, "block-period", -1, "Block period in seconds. Default is variable based on selected blockchain provider.")
	initEthereumCmd.Flags().StringVar(&initOptions.ContractAddress, "contract-address", "", "Do not automatically deploy a contract, instead use a pre-configured address")
	initEthereumCmd.Flags().StringVar(&initOptions.RemoteNodeURL, "remote-node-url", "", "For cases where the node is pre-existing and running remotely")
	initEthereumCmd.Flags().Int64Var(&initOptions.ChainID, "chain-id", 2021, "The chain ID - also used as the network ID")
	initEthereumCmd.Flags().StringVarP(&initOptions.BlockchainConnector, "blockchain-connector", "c", "evmconnect", "Blockchain connector to use. Options are: [evmconnect ethconnect]")
	initEthereumCmd.Flags().StringVarP(&initOptions.BlockchainNodeProvider, "blockchain-node", "n", "geth", fmt.Sprintf("Blockchain node type to use. Options are: %v", fftypes.FFEnumValues(types.BlockchainNodeProvider)))

	initCmd.AddCommand(initEthereumCmd)
}
