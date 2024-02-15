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
	"errors"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/spf13/cobra"
)

var pullOptions types.PullOptions

var pullCmd = &cobra.Command{
	Use:               "pull <stack_name>",
	Short:             "Pull a stack",
	ValidArgsFunction: listStacks,
	Long: `Pull a stack

Pull the images for a stack .
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var spin *spinner.Spinner
		if fancyFeatures && !verbose {
			spin = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
			logger = log.NewSpinnerLogger(spin)
		}
		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		version, err := docker.CheckDockerConfig()
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, docker.CtxComposeVersionKey{}, version)

		stackManager := stacks.NewStackManager(ctx)
		if len(args) == 0 {
			return errors.New("no stack specified")
		}
		stackName := args[0]

		if err := stackManager.LoadStack(stackName); err != nil {
			return err
		}
		if spin != nil {
			spin.Start()
		}
		if err := stackManager.PullStack(&pullOptions); err != nil {
			return err
		}
		if spin != nil {
			spin.Stop()
		}
		return nil
	},
}

func init() {
	pullCmd.Flags().IntVarP(&pullOptions.Retries, "retries", "r", 0, "Retry attempts to perform on image pull failure")

	rootCmd.AddCommand(pullCmd)
}
