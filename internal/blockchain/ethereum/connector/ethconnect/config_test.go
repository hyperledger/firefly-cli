package ethconnect

import (
	"path/filepath"
	"testing"
)

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()
	// Construct the paths for config files within the temporary directory
	configFilename := filepath.Join(dir, "config.yaml")
	extraEvmConfigPath := filepath.Join(dir, "conflate", "extra.yaml")
	p := Config{}
	t.Run("TestWriteConfig", func(t *testing.T) {
		err := p.WriteConfig(configFilename, extraEvmConfigPath)
		if err != nil {
			t.Logf("unable to write to config files: %v", err)
		}
	})
}
