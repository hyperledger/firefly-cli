package tessera

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateTesseraKeys(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name         string
		Stack        *types.Stack
		TesseraImage string
		KeysPrefix   string
		KeysName     string
	}{
		{
			Name: "testcase1",
			Stack: &types.Stack{
				Name:    "Org-1_quorum",
				InitDir: t.TempDir(),
			},
			TesseraImage: "quorumengineering/tessera:24.4",
			KeysPrefix:   "",
			KeysName:     "tm",
		},
		{
			Name: "testcase2",
			Stack: &types.Stack{
				Name:    "Org-1_quorum",
				InitDir: t.TempDir(),
			},
			TesseraImage: "quorumengineering/tessera:24.4",
			KeysPrefix:   "xyz",
			KeysName:     "tm",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			privateKey, publicKey, tesseraKeysPath, err := CreateTesseraKeys(ctx, tc.TesseraImage, filepath.Join(tc.Stack.InitDir, "tessera", "tessera_0", "keystore"), tc.KeysPrefix, tc.KeysName)
			if err != nil {
				t.Log("unable to create tessera keys", err)
			}
			//validate properties of tessera keys
			assert.NotEmpty(t, privateKey)
			assert.NotEmpty(t, publicKey)
			assert.NotEmpty(t, tesseraKeysPath)

			expectedOutputName := tc.KeysName
			if tc.KeysPrefix != "" {
				expectedOutputName = fmt.Sprintf("%s_%s", tc.KeysPrefix, expectedOutputName)
			}
			assert.Equal(t, tesseraKeysPath, filepath.Join(tc.Stack.InitDir, "tessera", "tessera_0", "keystore", expectedOutputName), "invalid output path")

			assert.Nil(t, err)
		})
	}
}

func TestCreateTesseraEntrypoint(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name        string
		Stack       *types.Stack
		StackName   string
		MemberCount int
	}{
		{
			Name: "testcase1",
			Stack: &types.Stack{
				Name:    "Org-1_quorum",
				InitDir: t.TempDir(),
			},
			StackName:   "org1",
			MemberCount: 4,
		},
		{
			Name: "testcase2",
			Stack: &types.Stack{
				Name:    "Org-2_quorum",
				InitDir: t.TempDir(),
			},
			StackName:   "org2",
			MemberCount: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := CreateTesseraEntrypoint(ctx, tc.Stack.InitDir, tc.StackName, tc.MemberCount)
			if err != nil {
				t.Log("unable to create tessera docker entrypoint", err)
			}
			path := filepath.Join(tc.Stack.InitDir, "docker-entrypoint.sh")
			_, err = os.Stat(path)
			assert.NoError(t, err, "docker entrypoint file not created")

			b, err := os.ReadFile(path)
			assert.NoError(t, err, "unable to read docker entrypoint file")
			for i := 0; i < tc.MemberCount; i++ {
				strings.Contains(string(b), fmt.Sprintf("member%dtessera", i))
			}
		})
	}
}
