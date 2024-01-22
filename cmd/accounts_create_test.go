package cmd

import (
	"os"
	"path/filepath"
	"runtime"
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
			Args:             []string{"ff", "create", "stack-1"},
			ExpectedResponse: "account1.output",
		},
		{
			Name:             "testcase-2",
			Args:             []string{"ff", "create", "stack-2"},
			ExpectedResponse: "account1.output",
		},
		{
			Name:             "testcase-3",
			Args:             []string{"ff", "create", "stack-3"},
			ExpectedResponse: "account1.output",
		},
		{
			Name:             "testcase-4",
			Args:             []string{"ff", "create", "stack-4"},
			ExpectedResponse: "account1.output",
		},
		{
			Name:             "testcase-5",
			Args:             []string{"ff", "create", "stack-5"},
			ExpectedResponse: "account1.output",
		},
		{
			Name:             "testcase-6",
			Args:             []string{"ff", "create", "stack-6"},
			ExpectedResponse: "account1.output",
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
			// Load the expected response from file
			// get current directory
			_, filename, _, ok := runtime.Caller(0)
			if !ok {
				t.Fatal("Not able to get current working directory")
			}
			currDir := filepath.Dir(filename)
			expectedResponseFile := filepath.Join(currDir, "testdata", tc.ExpectedResponse)
			expectedResponse, err := utils.ReadFileToString(expectedResponseFile)
			if err != nil {
				t.Fatalf("Failed to read expected response file: %v", err)
			}

			// Compare actual and expected responses
			assert.Equal(t, expectedResponse, actualResponse, "Responses do not match")

			assert.NotNil(t, actualResponse)
		})
	}
}
