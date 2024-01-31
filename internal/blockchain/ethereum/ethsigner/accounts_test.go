package ethsigner

import (
	"context"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestWriteTomlKeyFile(t *testing.T) {
	t.Run("TestwriteTomlKeyFile", func(t *testing.T) {
		directory := "testdata"
		FilePath := directory + "/wallet.toml"

		p := &EthSignerProvider{}

		File, err := p.writeTomlKeyFile(FilePath)
		if err != nil {
			t.Fatalf("unable to write file: %v", err)
		}
		assert.NotNil(t, File)
	})

}

func TestCopyTomlFileToVolume(t *testing.T) {
	t.Run("TestCopyTomltoVolume", func(t *testing.T) {
		ctx := log.WithLogger(context.Background(), &log.StdoutLogger{})

		directory := "testdata"
		tomlPath := directory + "/copy.toml"
		VolumeName := "ethsigner"

		p := &EthSignerProvider{}
		err := p.copyTomlFileToVolume(ctx, tomlPath, VolumeName)
		if err != nil {
			t.Fatalf("unable to copy file: %v", err)
		}
	})

}