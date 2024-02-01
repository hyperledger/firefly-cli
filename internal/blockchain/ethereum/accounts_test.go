package ethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletFile(t *testing.T) {
	dir := "walletdata"

	prefix := "WalletPair"
	outputDirectory := dir + "wallet.json"
	password := "26371628355334###"
	t.Run("TestCreateWalletFile", func(t *testing.T) {
		keypair, filename, err := CreateWalletFile(outputDirectory, prefix, password)
		if err != nil {
			t.Logf("unable to create wallet file %v: ", err)
		}
		assert.NotNil(t, keypair)
		assert.NotNil(t, filename)
	})
}
