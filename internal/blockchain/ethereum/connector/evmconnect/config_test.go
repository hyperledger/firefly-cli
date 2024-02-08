package evmconnect

import (
	"path/filepath"
	"testing"
)

func TestWriteConfig(t *testing.T) {
	dir := "testdata"
	configFilename := dir + filepath.Join("config.yaml")
	extraEvmConfigPath := dir + filepath.Join("/conflate/extra.yaml")
	p := Config{}
	t.Run("TestWriteConfig", func(t *testing.T) {
		err := p.WriteConfig(configFilename, extraEvmConfigPath)
		if err != nil {
			t.Logf("unable to write to config files: %v", err)
		}
	})
}
