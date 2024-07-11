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

package evmconnect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/connector"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Log           *types.LogConfig           `yaml:"log,omitempty"`
	Connector     *ConnectorConfig           `yaml:"connector,omitempty"`
	Metrics       *types.MetricsServerConfig `yaml:"metrics,omitempty"`
	Persistence   *PersistenceConfig         `yaml:"persistence,omitempty"`
	FFCore        *FFCoreConfig              `yaml:"ffcore,omitempty"`
	Confirmations *ConfirmationsConfig       `yaml:"confirmations,omitempty"`
	PolicyEngine  *PolicyEngineSimpleConfig  `yaml:"policyengine.simple,omitempty"`
	API           *APIConfig                 `yaml:"api,omitempty"`
}

type APIConfig struct {
	Port      int    `yaml:"port,omitempty"`
	Address   string `yaml:"address,omitempty"`
	PublicURL string `yaml:"publicURL,omitempty"`
}

type ConnectorConfig struct {
	URL string `yaml:"url,omitempty"`
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
	Required *int `yaml:"required,omitempty"`
}

type PolicyEngineSimpleConfig struct {
	FixedGasPrice *int             `yaml:"fixedGasPrice,omitempty"`
	GasOracle     *GasOracleConfig `yaml:"gasOracle,omitempty"`
}

type GasOracleConfig struct {
	Mode string `yaml:"mode,omitempty"`
}

func (e *Config) WriteConfig(filename string, extraEvmconnectConfigPath string) error {
	configYamlBytes, _ := yaml.Marshal(e)
	basedir := filepath.Dir(filename)
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filename, configYamlBytes, 0755); err != nil {
		return err
	}
	if extraEvmconnectConfigPath != "" {
		c, err := conflate.FromFiles(filename, extraEvmconnectConfigPath)
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

func (e *Evmconnect) GenerateConfig(stack *types.Stack, org *types.Organization, blockchainServiceName string) connector.Config {
	confirmations := new(int)
	*confirmations = 0
	fixedGasPrice := new(int)
	*fixedGasPrice = 0
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

	return &Config{
		Log: &types.LogConfig{
			Level: "debug",
		},
		API: &APIConfig{
			Port:      e.Port(),
			Address:   "0.0.0.0",
			PublicURL: fmt.Sprintf("http://127.0.0.1:%v", org.ExposedConnectorPort),
		},
		Connector: &ConnectorConfig{
			URL: fmt.Sprintf("http://%s:8545", blockchainServiceName),
		},
		Persistence: &PersistenceConfig{
			LevelDB: &LevelDBConfig{
				Path: "/evmconnect/data/leveldb",
			},
		},
		FFCore: &FFCoreConfig{
			URL:        getCoreURL(org),
			Namespaces: []string{"default"},
		},
		Metrics: metrics,
		Confirmations: &ConfirmationsConfig{
			Required: confirmations,
		},
		PolicyEngine: &PolicyEngineSimpleConfig{
			FixedGasPrice: fixedGasPrice,
			GasOracle: &GasOracleConfig{
				Mode: "fixed",
			},
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
