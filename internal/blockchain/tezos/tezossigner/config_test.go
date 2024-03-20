package tezossigner

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
	accountAddresses := []string{"0xAddress21", "0xAddress91", "0xAddress10", "0xAddress17", "0xAddress12"}

	config := GenerateSignerConfig(accountAddresses)

	expectedConfig := &Config{
		Server: ServerConfig{
			Address:        ":6732",
			UtilityAddress: ":9583",
		},
		Vaults: VaultsConfig{
			LocalSecret: LocalSecretConfig{
				Driver: "file",
				File: FileConfig{
					SecretPath: "/etc/secret.json",
				},
			},
		},
		Tezos: map[string]AccountConfig{
			"0xAddress21": {
				LogPayloads: true,
				Allow: AllowedTransactionsConfig{
					Generic: []string{
						"transaction",
						"endorsement",
						"reveal",
						"origination",
					},
				},
			},
			"0xAddress91": {
				LogPayloads: true,
				Allow: AllowedTransactionsConfig{
					Generic: []string{
						"transaction",
						"endorsement",
						"reveal",
						"origination",
					},
				},
			},
			"0xAddress10": {
				LogPayloads: true,
				Allow: AllowedTransactionsConfig{
					Generic: []string{
						"transaction",
						"endorsement",
						"reveal",
						"origination",
					},
				},
			},
			"0xAddress17": {
				LogPayloads: true,
				Allow: AllowedTransactionsConfig{
					Generic: []string{
						"transaction",
						"endorsement",
						"reveal",
						"origination",
					},
				},
			},
			"0xAddress12": {
				LogPayloads: true,
				Allow: AllowedTransactionsConfig{
					Generic: []string{
						"transaction",
						"endorsement",
						"reveal",
						"origination",
					},
				},
			},
		},
	}
	assert.Equal(t, expectedConfig.Server, config.Server, "Server configuration should match")
	assert.Equal(t, expectedConfig.Vaults, config.Vaults, "Vaults configuration should match")
	assert.Equal(t, expectedConfig.Tezos, config.Tezos, "Tezos configuration should match")
}
