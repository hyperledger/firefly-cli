// Copyright © 2024 Kaleido, Inc.
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

package tezosconnect

import (
	"fmt"
	"os"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/blockchain/tezos/connector"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log           *types.LogConfig           `yaml:"log,omitempty"`
	Connector     *ConnectorConfig           `yaml:"connector,omitempty"`
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
}

type BlockchainConfig struct {
	Network   string `yaml:"network,omitempty"`
	RPC       string `yaml:"rpc,omitempty"`
	Signatory string `yaml:"signatory,omitempty"`
}

type PersistenceConfig struct {
	LevelDB *LevelDBConfig `yaml:"leveldb,omitempty"`
}

type LevelDBConfig struct {
	Path string `yaml:"path,omitempty"`
}

type FFCoreConfig struct {
	URL        string   `yaml:"url,omitempty"`
	Namespaces []string `yaml:"namespaces,omitempty"`
}

type ConfirmationsConfig struct {
	Required              *int `yaml:"required,omitempty"`
	FetchReceiptUponEntry bool `yaml:"fetchReceiptUponEntry,omitempty"`
}

func (c *Config) WriteConfig(filename string, extraTezosconnectConfigPath string) error {
	configYamlBytes, _ := yaml.Marshal(c)
	if err := os.WriteFile(filename, configYamlBytes, 0755); err != nil {
		return err
	}
	if extraTezosconnectConfigPath != "" {
		c, err := conflate.FromFiles(filename, extraTezosconnectConfigPath)
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

func (t *Tezosconnect) GenerateConfig(stack *types.Stack, org *types.Organization, signerHostname, rpcURL string) connector.Config {
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

	network := "mainnet"
	if strings.Contains(rpcURL, "ghost") {
		network = "ghostnet"
	}

	return &Config{
		Log: &types.LogConfig{
			Level: "debug",
		},
		API: &APIConfig{
			Port:      t.Port(),
			Address:   "0.0.0.0",
			PublicURL: fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort),
		},
		Connector: &ConnectorConfig{
			Blockchain: &BlockchainConfig{
				Network:   network,
				RPC:       rpcURL,
				Signatory: fmt.Sprintf("http://%s:6732", signerHostname),
			},
		},
		Persistence: &PersistenceConfig{
			LevelDB: &LevelDBConfig{
				Path: "/tezosconnect/db/leveldb",
			},
		},
		FFCore: &FFCoreConfig{
			URL:        getCoreURL(org),
			Namespaces: []string{"default"},
		},
		Metrics: metrics,
		Confirmations: &ConfirmationsConfig{
			Required:              confirmations,
			FetchReceiptUponEntry: true,
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
