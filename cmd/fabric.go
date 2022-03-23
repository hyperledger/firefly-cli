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

// fabricCmd represents the fabric command
var fabricCmd = &cobra.Command{
	Use:   "fabric <stack_name> <chaincode_package> <channel> <chaincodeName> <version>",
	Short: "Deploy fabric chaincode",
	Long:  `Deploy a packaged chaincode to the Fabric network used by a FireFly stack`,
	Args:  cobra.ExactArgs(5),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDockerConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]
		filename := args[1]
		channel := args[2]
		chaincode := args[3]
		version := args[4]
		stackManager := stacks.NewStackManager(logger)
		if err := stackManager.LoadStack(stackName, verbose); err != nil {
			return err
		}
		fmt.Printf("deploying %s... ", filename)
		contractAddress, err := stackManager.DeployContract(filename, filename, 0, channel, chaincode, version)
		if err != nil {
			return err
		}
		fmt.Print("done\n\n")
		fmt.Printf("deployed to: %s\n", contractAddress)
		fmt.Print("\n")
		return nil
	},
}

func init() {
	deployCmd.AddCommand(fabricCmd)
}
