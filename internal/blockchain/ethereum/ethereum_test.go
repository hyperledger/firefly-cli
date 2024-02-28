package ethereum

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestReadFireFlyContract(t *testing.T) {
	ctx := log.WithLogger(context.Background(), &log.StdoutLogger{})
	dir := "testdata"

	Stack := &types.Stack{
		Name: "Firefly_Ethereum",
		Members: []*types.Organization{
			{
				ID: "user_1",
				Account: &Account{
					Address:    "0x1f2a000000000000000000000000000000000000",
					PrivateKey: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
				},
				External: true,
			},
			{
				ID: "user_2",
				Account: &Account{
					Address:    "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
					PrivateKey: "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
				},
				External: true,
			},
			{
				ID: "user_3",
				Account: &Account{
					Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
					PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				},
				External: true,
			},
		},
		RuntimeDir: filepath.Join(dir + "Firefly.json"),
	}
	t.Run("TestReadFileflyContract", func(t *testing.T) {
		Contracts, err := ReadFireFlyContract(ctx, Stack)
		if err != nil {
			t.Logf("unable to read eth firefly contract: %v", err)
		}
		assert.Nil(t, Contracts)
	})

}
