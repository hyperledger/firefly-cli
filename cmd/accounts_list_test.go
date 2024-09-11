package cmd

import (
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestAccountListCmd(t *testing.T) {
	testNames := []string{"stack-1", "stack-2", "stack-3", "stack-4", "stack-5"}
	for _, stackNames := range testNames {
		createCmd := accountsCreateCmd
		createCmd.SetArgs([]string{ExecutableName, "create", stackNames})
		err := createCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to create account for testing: %v", err)
		}
		Args := []string{"ls"}
		t.Run("Test-Account-List", func(t *testing.T) {
			cmd := accountsListCmd
			cmd.SetArgs(Args)

			_, outputBuffer := utils.CaptureOutput()
			cmd.SetOut(outputBuffer)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}
			actualResponse := outputBuffer.String()

			assert.NotNil(t, actualResponse)

		})
	}

}
