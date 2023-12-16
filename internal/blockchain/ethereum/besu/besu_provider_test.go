package besu

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/stretchr/testify/assert"
)

func TestNewBesuProvider(t *testing.T) {
	var ctx context.Context
	testCases := []struct {
		Name  string
		Ctx   context.Context
		Stack *types.Stack
	}{
		{
			Name: "testcase1",
			Ctx:  ctx,
			Stack: &types.Stack{
				Name:                   "TestBesuProviderWithEthconnect",
				Members:                []*types.Organization{{OrgName: "Org1"}},
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "EthConnect"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "EthConnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "EthConnect"),
			},
		},
		{
			Name: "TestBesuProviderWithEvmconnect",
			Stack: &types.Stack{
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "Geth"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "EvmConnect"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "Geth"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			besuProvider := NewBesuProvider(tc.Ctx, tc.Stack)
			assert.NotNil(t, besuProvider)
		})
	}
}

