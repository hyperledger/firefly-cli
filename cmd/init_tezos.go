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

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var initTezosCmd = &cobra.Command{
	Use:   "tezos [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack using an Tezos blockchain",
	Long:  `Create a new FireFly local dev stack using an Tezos blockchain`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)
		stackManager := stacks.NewStackManager(ctx)
		initOptions.BlockchainProvider = types.BlockchainProviderTezos.String()
		initOptions.BlockchainConnector = types.BlockchainConnectorTezosconnect.String()
		initOptions.BlockchainNodeProvider = types.BlockchainNodeProviderRemoteRPC.String()
		// By default we turn off multiparty mode while it's not supported yet
		initOptions.MultipartyEnabled = false
		initOptions.TokenProviders = []string{}
		if err := validateTezosFlags(); err != nil {
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

func validateTezosFlags() error {
	if initOptions.RemoteNodeURL == "" {
		return fmt.Errorf("you must provide 'remote-node-url' flag as local node mode is not supported")
	}
	return nil
}

func init() {
	initTezosCmd.Flags().IntVar(&initOptions.BlockPeriod, "block-period", -1, "Block period in seconds. Default is variable based on selected blockchain provider.")
	initTezosCmd.Flags().StringVar(&initOptions.ContractAddress, "contract-address", "", "Do not automatically deploy a contract, instead use a pre-configured address")
	initTezosCmd.Flags().StringVar(&initOptions.RemoteNodeURL, "remote-node-url", "", "For cases where the node is pre-existing and running remotely")

	initCmd.AddCommand(initTezosCmd)
}
