package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func prompt(promptText string, validate func(string) error) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(promptText)
		if str, err := reader.ReadString('\n'); err != nil {
			return "", err
		} else {
			str = strings.TrimSpace(str)
			if err := validate(str); err != nil {
				if fancyFeatures {
					fmt.Printf("\u001b[31mError: %s\u001b[0m\n", err.Error())
				} else {
					fmt.Printf("Error: %s\n", err.Error())
				}
			} else {
				return str, nil
			}
		}
	}
}

func confirm(promptText string) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N] ", promptText)
		if str, err := reader.ReadString('\n'); err != nil {
			return err
		} else {
			str = strings.ToLower(strings.TrimSpace(str))
			if str == "y" || str == "yes" {
				return nil
			} else {
				return fmt.Errorf("confirmation declined with response: '%s'", str)
			}
		}
	}
}
