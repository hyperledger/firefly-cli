package cmd

import (
	"os"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccountCmd(t *testing.T) {

	testcases := []struct {
		Name string
		Args []string
	}{
		{
			Name: "testcase1",
			Args: []string{"ff", "create", "stack-1"},
		},
		{
			Name: "testcase-2",
			Args: []string{"ff", "create", "stack-2"},
		},
		{
			Name: "testcase-3",
			Args: []string{"ff", "create", "stack-3"},
		},
		{
			Name: "testcase-4",
			Args: []string{"ff", "create", "stack-4"},
		},
		{
			Name: "testcase-5",
			Args: []string{"ff", "create", "stack-5"},
		},
		{
			Name: "testcase-6",
			Args: []string{"ff", "create", "stack-6"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {

			cmd := accountsCreateCmd
			cmd.SetArgs(tc.Args)

			// Capture the output
			originalOutput, outputBuffer := utils.CaptureOutput()
			defer func() {
				// Restore the original output after capturing
				os.Stdout = originalOutput
			}()
			cmd.SetOutput(outputBuffer)

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Get the actual response
			actualResponse := outputBuffer.String()

			assert.NotNil(t, actualResponse)
		})
	}
}
