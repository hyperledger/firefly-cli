package cmd

import (
	"os"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccountCmd(t *testing.T) {

	testcases := []struct {
		Name             string
		Args             []string
		ExpectedResponse string
	}{
		{
			Name:             "testcase1",
			Args:             []string{"create", "stack-1"},
			ExpectedResponse: "",
		},
		{
			Name:             "testcase-2",
			Args:             []string{"create", "stack-2"},
			ExpectedResponse: "",
		},
		{
			Name:             "testcase-3",
			Args:             []string{"create", "stack-3"},
			ExpectedResponse: "",
		},
		{
			Name:             "testcase-4",
			Args:             []string{"create", "stack-4"},
			ExpectedResponse: "",
		},
		{
			Name:             "testcase-5",
			Args:             []string{"create", "stack-5"},
			ExpectedResponse: "",
		},
		{
			Name:             "testcase-6",
			Args:             []string{"create", "stack-6"},
			ExpectedResponse: "",
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
			cmd.SetOut(outputBuffer)

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Get the actual response
			actualResponse := outputBuffer.String()

			// Compare actual and expected responses
			assert.Equal(t, tc.ExpectedResponse, actualResponse, "Responses do not match")

			assert.NotNil(t, actualResponse)
		})
	}
}
