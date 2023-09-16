/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate markdown docs",
	Long: `Generate markdown docs for the entire command tree.
			
	The command takes an optional argument specifying directory to put the
	generated documentation, default is "{cwd}/docs/command_docs/"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string

		if len(args) == 0 {
			currentWoringDir, err := os.Getwd()
			if err != nil {
				return err
			}
			path = fmt.Sprintf("%s/docs/command_docs", currentWoringDir)
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		} else {
			path = args[0]
		}

		err := doc.GenMarkdownTree(rootCmd, path)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// docsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// docsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
