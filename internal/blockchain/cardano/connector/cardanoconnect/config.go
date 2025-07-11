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

package cardanoconnect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/cardano/connector"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log           *types.LogConfig           `yaml:"log,omitempty"`
	Connector     *ConnectorConfig           `yaml:"connector,omitempty"`
	Contracts     *ContractsConfig           `yaml:"contracts,omitempty"`
	Metrics       *types.MetricsServerConfig `yaml:"metrics,omitempty"`
	Persistence   *PersistenceConfig         `yaml:"persistence,omitempty"`
	FFCore        *FFCoreConfig              `yaml:"ffcore,omitempty"`
	Confirmations *ConfirmationsConfig       `yaml:"confirmations,omitempty"`
	API           *APIConfig                 `yaml:"api,omitempty"`
}

type APIConfig struct {
	Port      int    `yaml:"port,omitempty"`
	Address   string `yaml:"address,omitempty"`
	PublicURL string `yaml:"publicURL,omitempty"`
}

type ConnectorConfig struct {
	Blockchain *BlockchainConfig `yaml:"blockchain,omitempty"`
	SignerURL  string            `yaml:"signerUrl,omitempty"`
}

type BlockchainConfig struct {
	BlockfrostKey     string `yaml:"blockfrostKey,omitempty"`
	BlockfrostBaseURL string `yaml:"blockfrostBaseUrl,omitempty"`
	Socket            string `yaml:"socket,omitempty"`
	Network           string `yaml:"network,omitempty"`
}

type ContractsConfig struct {
	ComponentsPath string `yaml:"componentsPath"`
	StoresPath     string `yaml:"storesPath"`
}

type PersistenceConfig struct {
	Type string `yaml:"type,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type FFCoreConfig struct {
	URL        string   `yaml:"url,omitempty"`
	Namespaces []string `yaml:"namespaces,omitempty"`
}

type ConfirmationsConfig struct {
	Required *int `yaml:"required,omitempty"`
}

func (c *Config) WriteConfig(filename string, extraConnectorConfigPath string) error {
	configYamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	basedir := filepath.Dir(filename)
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filename, configYamlBytes, 0755); err != nil {
		return err
	}
	if extraConnectorConfigPath != "" {
		c, err := conflate.FromFiles(filename, extraConnectorConfigPath)
		if err != nil {
			return err
		}
		bytes, err := c.MarshalYAML()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, bytes, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cardanoconnect) GenerateConfig(stack *types.Stack, org *types.Organization) connector.Config {
	confirmations := new(int)
	*confirmations = 0
	var metrics *types.MetricsServerConfig

	if stack.PrometheusEnabled {
		metrics = &types.MetricsServerConfig{
			HTTPServerConfig: types.HTTPServerConfig{
				Port:      org.ExposedConnectorMetricsPort,
				Address:   "0.0.0.0",
				PublicURL: fmt.Sprintf("http://127.0.0.1:%d", org.ExposedConnectorMetricsPort),
			},
			Enabled: true,
			Path:    "/metrics",
		}
	} else {
		metrics = nil
	}

	socket := ""
	if stack.Socket != "" {
		socket = "/ipc/socket"
	}

	return &Config{
		Log: &types.LogConfig{
			Level: "info",
		},
		API: &APIConfig{
			Port:      c.Port(),
			Address:   "0.0.0.0",
			PublicURL: fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort),
		},
		Connector: &ConnectorConfig{
			Blockchain: &BlockchainConfig{
				BlockfrostKey:     stack.BlockfrostKey,
				BlockfrostBaseURL: stack.BlockfrostBaseURL,
				Network:           stack.Network,
				Socket:            socket,
			},
			SignerURL: fmt.Sprintf("http://%s_cardanosigner:8555", stack.Name),
		},
		Contracts: &ContractsConfig{
			ComponentsPath: "/cardanoconnect/contracts/components",
			StoresPath:     "/cardanoconnect/contracts/stores",
		},
		Persistence: &PersistenceConfig{
			Type: "sqlite",
			Path: "/cardanoconnect/sqlite/db.sqlite3",
		},
		FFCore: &FFCoreConfig{
			URL:        getCoreURL(org),
			Namespaces: []string{"default"},
		},
		Metrics: metrics,
		Confirmations: &ConfirmationsConfig{
			Required: confirmations,
		},
	}
}

func getCoreURL(org *types.Organization) string {
	host := fmt.Sprintf("firefly_core_%v", org.ID)
	if org.External {
		host = "host.docker.internal"
	}
	return fmt.Sprintf("http://%s:%v", host, org.ExposedFireflyPort)
}
