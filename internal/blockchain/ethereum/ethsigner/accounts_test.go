package ethsigner

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteTomlKeyFile(t *testing.T) {
	t.Run("TestwriteTomlKeyFile", func(t *testing.T) {
		directory := t.TempDir()
		FilePath := filepath.Join(directory + "/wallet.toml")

		p := &EthSignerProvider{}

		File, err := p.writeTomlKeyFile(FilePath)
		if err != nil {
			t.Fatalf("unable to write file: %v", err)
		}
		assert.NotNil(t, File)
	})

}
