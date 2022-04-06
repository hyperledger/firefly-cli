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

// ethereumCmd represents the ethereum command
var ethereumCmd = &cobra.Command{
	Use:   "ethereum <stack_name> <contract_json_file>",
	Short: "Deploy a compiled solidity contract",
	Long: `Deploy a solidity contract compiled with solc to the blockchain used by a FireFly stack

To compile a .sol file to a .json file run:

solc --combined-json abi,bin contract.sol > contract.json
`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDockerConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]
		filename := args[1]
		stackManager := stacks.NewStackManager(logger)
		if err := stackManager.LoadStack(stackName, verbose); err != nil {
			return err
		}
		contractNames, err := stackManager.GetContracts(filename, args[2:])
		if err != nil {
			return err
		}
		if len(contractNames) < 1 {
			return fmt.Errorf("no contracts found in file: '%s'", filename)
		}
		selectedContractName := contractNames[0]
		if len(contractNames) > 1 {
			selectedContractName, err = selectMenu("select the contract to deploy", contractNames)
			fmt.Print("\n")
			if err != nil {
				return err
			}
		}
		fmt.Printf("deploying %s... ", selectedContractName)
		contractAddress, err := stackManager.DeployContract(filename, selectedContractName, 0, args[2:])
		if err != nil {
			return err
		}

		fmt.Print("done\n\n")
		fmt.Print(contractAddress)
		fmt.Print("\n")
		return nil
	},
}

func init() {
	deployCmd.AddCommand(ethereumCmd)
}
