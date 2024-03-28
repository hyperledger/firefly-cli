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
	"time"

	"github.com/briandowns/spinner"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

var forceUpgrade bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <stack_name> <version>",
	Short: "Upgrade a stack to different version",
	Long: `Upgrade a stack by pulling updated images.
	This operation will stop the stack if running.
	If certain containers were pinned to a specific image at init,
	this command will have no effect on those containers.`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: listStacks,
	RunE: func(cmd *cobra.Command, args []string) error {
		var spin *spinner.Spinner
		if fancyFeatures && !verbose {
			spin = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
			logger = log.NewSpinnerLogger(spin)
		}
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		dockerVersion, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, dockerVersion)

		stackManager := stacks.NewStackManager(ctx)
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]
		if len(args) <= 1 {
			return fmt.Errorf("no version specified")
		}
		version := args[1]

		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		fmt.Printf("upgrading stack '%s'... ", stackName)
		if err := stackManager.UpgradeStack(version, forceUpgrade); err != nil {
			return err
		}
		fmt.Printf("\n\nYour stack has been upgraded to %s\n\nTo start your upgraded stack run:\n\n%s start %s\n\n", version, rootCmd.Use, stackName)
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVarP(&forceUpgrade, "force", "f", false, "Force upgrade even between unsupported versions. May result in a broken environment. Use with caution.")
	rootCmd.AddCommand(upgradeCmd)
}
