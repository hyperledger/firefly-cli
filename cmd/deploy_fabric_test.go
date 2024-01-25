package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Not able to get current working directory")
	}
	currDir := filepath.Dir(filename)
	chainCodefile := filepath.Join(currDir, "testdata", "fabric_deploy.json")
	ChaincodePackage, err := utils.ReadFileToString(chainCodefile)
	if err != nil {
		t.Fatalf("Failed to read expected response file: %v", err)
	}
	Args := []string{"fabric", ChaincodePackage, "firefly", "fabric-user-1", "1.0"}

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
