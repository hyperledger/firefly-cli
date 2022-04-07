/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// accountsListCmd represents the "accounts create" command
var accountsListCmd = &cobra.Command{
	Use:     "list <stack_name>",
	Short:   "List the accounts in the FireFly stack",
	Long:    `List the accounts in the FireFly stack`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"ls"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDockerConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]
		stackManager := stacks.NewStackManager(logger)
		if err := stackManager.LoadStack(stackName, verbose); err != nil {
			return err
		}
		accounts, err := json.MarshalIndent(stackManager.Stack.State.Accounts, "", "  ")
		if err != nil {
			return err
		}
		fmt.Print(string(accounts))
		return nil
	},
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
}
