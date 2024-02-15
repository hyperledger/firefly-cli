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

func TestDeployEthereumCmd(t *testing.T) {
	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	createcmd := accountsCreateCmd
	createcmd.SetArgs([]string{"create", "stack-2"})
	err := createcmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unable to create stack : %v", err)
	}
	currDir := t.TempDir()
	contractFile := filepath.Join(currDir + "eth_deploy.json")

	ethPackage, err := utils.ReadFileToString(contractFile)
	if err != nil {
		t.Fatalf("Failed to read expected response file: %v", err)
	}
	Args := []string{"deploy", "ethereum", "stack-2", ethPackage, "param1", "param2"}
	ethDeployCmd := deployEthereumCmd
	ethDeployCmd.SetArgs(Args)
	ethDeployCmd.ExecuteContext(ctx)

	Outputwriter, outputBuffer := utils.CaptureOutput()
	defer func() {
		os.Stdout = Outputwriter
	}()
	ethDeployCmd.SetOutput(outputBuffer)

	actualResponse := outputBuffer.String()
	assert.NotNil(t, actualResponse)
}
