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

package core

import (
	"fmt"
	"os"
	"path"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"
	"gopkg.in/yaml.v2"
)

func NewFireflyConfig(stack *types.Stack, member *types.Organization) *types.FireflyConfig {
	// TODO: If we move to support multiple namespaces at the same time, we will need to
	// change the Name field of some of these plugins

	spiHTTPConfig := types.HTTPServerConfig{
		Port:      member.ExposedFireflyAdminSPIPort,
		Address:   "0.0.0.0",
		PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyAdminSPIPort),
	}
	memberConfig := &types.FireflyConfig{
		Log: &types.LogConfig{
			Level: "debug",
		},
		Debug: &types.HTTPServerConfig{
			Port: 6060,
		},
		HTTP: &types.HTTPServerConfig{
			Port:      member.ExposedFireflyPort,
			Address:   "0.0.0.0",
			PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyPort),
		},
		Admin: &types.AdminServerConfig{
			HTTPServerConfig: spiHTTPConfig,
			Enabled:          true,
		},
		SPI: &types.SPIServerConfig{
			HTTPServerConfig: spiHTTPConfig,
			Enabled:          true,
		},
		UI: &types.UIConfig{
			Path: "./frontend",
		},
		Event: &types.EventConfig{
			DBEvents: &types.DBEventsConfig{
				BufferSize: 10000,
			},
		},
		Plugins: &types.Plugins{},
	}

	memberConfig.Plugins.SharedStorage = []*types.SharedStorageConfig{
		{
			Type: "ipfs",
			Name: "sharedstorage0",
			IPFS: &types.FireflyIPFSConfig{
				API: &types.HTTPEndpointConfig{
					URL: getIPFSAPIURL(member),
				},
				Gateway: &types.HTTPEndpointConfig{
					URL: getIPFSGatewayURL(member),
				},
			},
		},
	}

	memberConfig.Plugins.DataExchange = []*types.DataExchangeConfig{
		{
			Type: "ffdx",
			Name: "dataexchange0",
			FFDX: &types.HTTPEndpointConfig{
				URL: getDataExchangeURL(member),
			},
		},
	}

	if stack.PrometheusEnabled {
		memberConfig.Metrics = &types.MetricsServerConfig{
			HTTPServerConfig: types.HTTPServerConfig{
				Port:      member.ExposedFireflyMetricsPort,
				Address:   "0.0.0.0",
				PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyMetricsPort),
			},
			Enabled: true,
			Path:    "/metrics",
		}
	} else {
		memberConfig.Metrics = &types.MetricsServerConfig{
			Enabled: false,
		}
	}

	var databaseConfig *types.DatabaseConfig
	switch stack.Database {
	case types.DatabaseSelectionPostgres:
		databaseConfig = &types.DatabaseConfig{
			Name: "database0",
			Type: "postgres",
			PostgreSQL: &types.CommonDBConfig{
				URL: getPostgresURL(member),
				Migrations: &types.MigrationsConfig{
					Auto: true,
				},
			},
		}
	case types.DatabaseSelectionSQLite:
		databaseConfig = &types.DatabaseConfig{
			Name: "database0",
			Type: stack.Database.String(),
			SQLite3: &types.CommonDBConfig{
				URL: getSQLitePath(member, stack.RuntimeDir),
				Migrations: &types.MigrationsConfig{
					Auto: true,
				},
			},
		}
	}
	memberConfig.Plugins.Database = []*types.DatabaseConfig{databaseConfig}
	return memberConfig
}

func getIPFSAPIURL(member *types.Organization) string {
	if !member.External {
		return fmt.Sprintf("http://ipfs_%s:5001", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedIPFSApiPort)
	}
}

func getIPFSGatewayURL(member *types.Organization) string {
	if !member.External {
		return fmt.Sprintf("http://ipfs_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedIPFSGWPort)
	}
}

func getPostgresURL(member *types.Organization) string {
	if !member.External {
		return fmt.Sprintf("postgres://postgres:f1refly@postgres_%s:5432?sslmode=disable", member.ID)
	} else {
		return fmt.Sprintf("postgres://postgres:f1refly@127.0.0.1:%v?sslmode=disable", member.ExposedDatabasePort)
	}
}

func getSQLitePath(member *types.Organization, runtimeDir string) string {
	if !member.External {
		return "/etc/firefly/data/db/sqlite.db?_busy_timeout=5000"
	} else {
		return path.Join(runtimeDir, member.ID+".db")
	}
}

func getDataExchangeURL(member *types.Organization) string {
	if !member.External {
		return fmt.Sprintf("http://dataexchange_%s:3000", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedDataexchangePort)
	}
}

func ReadFireflyConfig(filePath string) (*types.FireflyConfig, error) {
	if bytes, err := os.ReadFile(filePath); err != nil {
		return nil, err
	} else {
		var config *types.FireflyConfig
		err := yaml.Unmarshal(bytes, &config)
		return config, err
	}
}

func WriteFireflyConfig(config *types.FireflyConfig, filePath, extraCoreConfigPath string) error {
	if bytes, err := yaml.Marshal(config); err != nil {
		return err
	} else {
		if err := os.WriteFile(filePath, bytes, 0755); err != nil {
			return err
		}
	}
	if extraCoreConfigPath != "" {
		c, err := conflate.FromFiles(filePath, extraCoreConfigPath)
		if err != nil {
			return err
		}
		bytes, err := c.MarshalYAML()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filePath, bytes, 0755); err != nil {
			return err
		}
	}
	return nil
}
