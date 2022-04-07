/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/stacks"
	"github.com/spf13/cobra"
)

// accountsCreateCmd represents the "accounts create" command
var accountsCreateCmd = &cobra.Command{
	Use:   "create <stack_name>",
	Short: "Create a new account in the FireFly stack",
	Long:  `Create a new account in the FireFly stack`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDockerConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]
		stackManager := stacks.NewStackManager(logger)
		if err := stackManager.LoadStack(stackName, verbose); err != nil {
			return err
		}
		account, err := stackManager.CreateAccount()
		if err != nil {
			return err
		}
		fmt.Print(account)
		fmt.Print("\n")
		return nil
	},
}

func init() {
	accountsCmd.AddCommand(accountsCreateCmd)
}
