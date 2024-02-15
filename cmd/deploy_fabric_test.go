package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestDeployFabricCmd(t *testing.T) {
	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	createCmd := accountsCreateCmd
	createCmd.SetArgs([]string{"create", "stack-1"})
	err := createCmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unable to execute command :%v", err)
	}
	currDir := t.TempDir()
	chainCodefile := filepath.Join(currDir + "fabric_deploy.json")
	Args := []string{"fabric", "stack-1", chainCodefile, "firefly", "fabric-user-1", "1.0"}

	t.Run("Test Deploy Cmd", func(t *testing.T) {
		DeployFabric := deployFabricCmd
		DeployFabric.SetArgs(Args)
		DeployFabric.ExecuteContext(ctx)

		originalOutput, outBuffer := utils.CaptureOutput()
		defer func() {
			os.Stdout = originalOutput
		}()
		DeployFabric.SetOutput(outBuffer)

		actualResponse := outBuffer.String()

		assert.NotNil(t, actualResponse)
	})
}
