// Copyright © 2025 IOG Singapore and SundaeSwap, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cardanosigner

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	API        APIConfig        `yaml:"api"`
	FileWallet FileWalletConfig `yaml:"fileWallet"`
}

type APIConfig struct {
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
		API: APIConfig{
			Address: "0.0.0.0",
			Port:    8555,
		},
		FileWallet: FileWalletConfig{
			Path: "/data/wallet",
		},
	}
	return config
}
