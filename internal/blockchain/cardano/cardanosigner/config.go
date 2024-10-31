package cardanosigner

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Api        ApiConfig        `yaml:"api"`
	FileWallet FileWalletConfig `yaml:"fileWallet"`
}

type ApiConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port,omitempty"`
}

type FileWalletConfig struct {
	Path string `yaml:"path"`
}

func (c *Config) WriteConfig(filename string) error {
	configYamlBytes, _ := yaml.Marshal(c)
	return os.WriteFile(filename, configYamlBytes, 0755)
}

func GenerateSignerConfig() *Config {
	config := &Config{
		Api: ApiConfig{
			Address: "0.0.0.0",
			Port:    8555,
		},
		FileWallet: FileWalletConfig{
			Path: "/data/wallet",
		},
	}
	return config
}
