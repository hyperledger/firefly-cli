package ethsigner

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()

	testfiles := []struct {
		Name     string
		FileName string
	}{
		{Name: "test-1", FileName: filepath.Join(dir + "test1.yaml")},
		{Name: "test-2", FileName: filepath.Join(dir + "test2.yaml")},
		{Name: "test-3", FileName: filepath.Join(dir + "test3.yaml")},
		{Name: "test-4", FileName: filepath.Join(dir + "test4.yaml")},
		{Name: "test-5", FileName: filepath.Join(dir + "test5.yaml")},
	}
	for _, tc := range testfiles {
		t.Run(tc.Name, func(t *testing.T) {
			c := &Config{}
			err := c.WriteConfig(tc.FileName)
			if err != nil {
				t.Logf("unable to write config files :%s", err)
			}
		})
	}
}

func TestGenerateSignerConfig(t *testing.T) {
	chainID := int64(12345)
	rpcURL := "http://localhost:8545"
	expectedConfig := &Config{
		Server: ServerConfig{
			Port:    8545,
			Address: "0.0.0.0",
		},
		Backend: BackendConfig{
			URL:     rpcURL,
			ChainID: &chainID,
		},
		FileWallet: FileWalletConfig{
			Path: "/data/keystore",
			Filenames: &FileWalletFilenamesConfig{
				PrimaryExt: ".toml",
			},
			Metadata: &FileWalletMetadataConfig{
				KeyFileProperty:      `{{ index .signing "key-file" }}`,
				PasswordFileProperty: `{{ index .signing "password-file" }}`,
			},
		},
		Log: LogConfig{
			Level: "debug",
		},
	}
	config := GenerateSignerConfig(chainID, rpcURL)
	assert.NotNil(t, config.Backend)
	assert.NotNil(t, config.Server)
	assert.NotNil(t, config.FileWallet)
	assert.NotNil(t, config.Log)

	assert.Equal(t, expectedConfig, config, "Generated config should match the expected config")
}
