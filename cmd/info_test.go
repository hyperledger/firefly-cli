package cmd

import (
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestInfoCmd(t *testing.T) {
	accNames := []string{"acc-1", "acc-2", "acc-3"}
	for _, stacks := range accNames {
		createAcc := accountsCreateCmd
		createAcc.SetArgs([]string{"create", stacks})
		err := createAcc.Execute()
		if err != nil {
			t.Fatalf("unable to execute command :%v", err)
		}
		args := []string{"info"}
		t.Run("Info Cmd Test", func(t *testing.T) {
			InfoCmd := infoCmd
			InfoCmd.SetArgs(args)

			_, outBuff := utils.CaptureOutput()
			InfoCmd.SetOut(outBuff)
			err = InfoCmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}
			actualResponse := outBuff.String()

			assert.NotNil(t, actualResponse)
		})
	}
}
