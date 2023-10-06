/*
Copyright © 2023 Giwa Oluwatobi <giwaoluwatobi@gmial.com>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps [a stack name]...",
	Short: "Returns information on running stacks",
	Long: `ps returns currently running stacks on your local machine.
	
	It also takes a continuous list of whitespace optional arguement - stack name. If non
	is given, it run the "ps" command for all stack on the local machine.`,
	Aliases: []string{"process"},
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := log.WithVerbosity(context.Background(), verbose)
		ctx = log.WithLogger(ctx, logger)

		allStacks, err := stacks.ListStacks()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			namedStacks := make([]string, 0, len(args))
			for _, stackName := range args {
				if contains(allStacks, strings.TrimSpace(stackName)) {
					namedStacks = append(namedStacks, stackName)
				} else {
					fmt.Printf("stack name - %s, is not present on your local machine. Run `ff ls` to see all available stacks.\n", stackName)
				}
			}

			allStacks = namedStacks // replace only the user specified stacks in the slice instead.
		}

		stackManager := stacks.NewStackManager(ctx)
		for _, stackName := range allStacks {
			if err := stackManager.LoadStack(stackName); err != nil {
				return err
			}

			if err := stackManager.IsRunning(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}

// contains can be removed if the go mod version is bumped to version Go 1.18
// and replaced with slices.Contains().
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
