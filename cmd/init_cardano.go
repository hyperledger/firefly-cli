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

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/spf13/cobra"
)

var initCardanoCmd = &cobra.Command{
	Use:   "cardano [network] [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack using a Cardano blockchain",
	Long:  "Create a new FireFly local dev stack using a Cardano blockchain",
	Args:  cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)
		stackManager := stacks.NewStackManager(ctx)
		initOptions.BlockchainProvider = types.BlockchainProviderCardano.String()
		initOptions.BlockchainConnector = types.BlockchainConnectorCardanoConnect.String()
		initOptions.BlockchainNodeProvider = types.BlockchainNodeProviderRemoteRPC.String()
		initOptions.MultipartyEnabled = false
		if err := initCommon(args); err != nil {
			return err
		}
		if err := stackManager.InitStack(&initOptions); err != nil {
			if err := stackManager.RemoveStack(); err != nil {
				return err
			}
			return err
		}
		return nil
	},
}

func init() {
	initCmd.AddCommand(initCardanoCmd)
}
