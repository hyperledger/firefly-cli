package quorum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestCreateQuorumEntrypoint(t *testing.T) {
	ctx := log.WithVerbosity(log.WithLogger(context.Background(), &log.StdoutLogger{}), false)
	testCases := []struct {
		Name                      string
		Stack                     *types.Stack
		Consensus                 string
		StackName                 string
		MemberIndex               int
		ChainID                   int
		BlockPeriodInSeconds      int
		PrivateTransactionManager fftypes.FFEnum
	}{
		{
			Name: "testcase1",
			Stack: &types.Stack{
				Name:    "Org-1_quorum",
				InitDir: t.TempDir(),
			},
			Consensus:                 "ibft",
			StackName:                 "org1",
			MemberIndex:               0,
			ChainID:                   1337,
			BlockPeriodInSeconds:      -1,
			PrivateTransactionManager: types.PrivateTransactionManagerTessera,
		},
		{
			Name: "testcase2_with_tessera_disabled",
			Stack: &types.Stack{
				Name:    "Org-2_quorum",
				InitDir: t.TempDir(),
			},
			Consensus:                 "clique",
			StackName:                 "org2",
			MemberIndex:               1,
			ChainID:                   1337,
			BlockPeriodInSeconds:      3,
			PrivateTransactionManager: types.PrivateTransactionManagerNone,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := CreateQuorumEntrypoint(ctx, tc.Stack.InitDir, tc.Consensus, tc.StackName, tc.MemberIndex, tc.ChainID, tc.BlockPeriodInSeconds, tc.PrivateTransactionManager)
			if err != nil {
				t.Log("unable to create quorum docker entrypoint", err)
			}
			path := filepath.Join(tc.Stack.InitDir, "docker-entrypoint.sh")
			_, err = os.Stat(path)
			assert.NoError(t, err, "docker entrypoint file not created")

			b, err := os.ReadFile(path)
			assert.NoError(t, err, "unable to read docker entrypoint file")
			output := string(b)
			strings.Contains(output, fmt.Sprintf("member%dtessera", tc.MemberIndex))
			strings.Contains(output, fmt.Sprintf("GOQUORUM_CONS_ALGO=%s", tc.Consensus))
		})
	}
}
