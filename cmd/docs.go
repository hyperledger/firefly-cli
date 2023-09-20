/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate markdown documentation for all command",
	Long: `Generate markdown documentation for the entire command tree.
	
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
			if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) || err != nil {
				err = fmt.Errorf("path you supplied for documentation is invalid: %v", err)
				return err
			}
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
}
