// Copyright © 2021 Kaleido, Inc.
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
	"io/ioutil"
	"path"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"gopkg.in/yaml.v2"
)

type LogConfig struct {
	Level string `yaml:"level,omitempty"`
}

type HttpServerConfig struct {
	Port      int    `yaml:"port,omitempty"`
	Address   string `yaml:"address,omitempty"`
	PublicURL string `yaml:"publicURL,omitempty"`
}

type AdminServerConfig struct {
	HttpServerConfig `yaml:",inline"`
	Enabled          bool `yaml:"enabled,omitempty"`
	PreInit          bool `yaml:"preinit,omitempty"`
}

type MetricsServerConfig struct {
	HttpServerConfig `yaml:",inline"`
	Enabled          bool   `yaml:"enabled,omitempty"`
	Path             string `yaml:"path,omitempty"`
}

type BasicAuth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type HttpEndpointConfig struct {
	URL  string    `yaml:"url,omitempty"`
	Auth BasicAuth `yaml:"auth,omitempty"`
}

type UIConfig struct {
	Path string `yaml:"path,omitempty"`
}

type NodeConfig struct {
	Name string `yaml:"name,omitempty"`
}

type OrgConfig struct {
	Name     string `yaml:"name,omitempty"`
	Identity string `yaml:"identity,omitempty"`
}

type EthconnectConfig struct {
	URL      string     `yaml:"url,omitempty"`
	Instance string     `yaml:"instance,omitempty"`
	Topic    string     `yaml:"topic,omitempty"`
	Auth     *BasicAuth `yaml:"auth,omitempty"`
}

type FabconnectConfig struct {
	URL       string `yaml:"url,omitempty"`
	Channel   string `yaml:"channel,omitempty"`
	Chaincode string `yaml:"chaincode,omitempty"`
	Topic     string `yaml:"topic,omitempty"`
	Signer    string `yaml:"signer,omitempty"`
}

type EthereumConfig struct {
	Ethconnect *EthconnectConfig `yaml:"ethconnect,omitempty"`
}

type FabricConfig struct {
	Fabconnect *FabconnectConfig `yaml:"fabconnect,omitempty"`
}

type BlockchainConfig struct {
	Type     string          `yaml:"type,omitempty"`
	Ethereum *EthereumConfig `yaml:"ethereum,omitempty"`
	Fabric   *FabricConfig   `yaml:"fabric,omitempty"`
}

type DataExchangeConfig struct {
	Type  string              `yaml:"type,omitempty"`
	HTTPS *HttpEndpointConfig `yaml:"https,omitempty"`
}

type CommonDBConfig struct {
	URL        string            `yaml:"url,omitempty"`
	Migrations *MigrationsConfig `yaml:"migrations,omitempty"`
}

type MigrationsConfig struct {
	Auto      bool   `yaml:"auto,omitempty"`
	Directory string `yaml:"directory,omitempty"`
}

type DatabaseConfig struct {
	Type       string          `yaml:"type,omitempty"`
	PostgreSQL *CommonDBConfig `yaml:"postgres,omitempty"`
	SQLite3    *CommonDBConfig `yaml:"sqlite3,omitempty"`
}

type PublicStorageConfig struct {
	Type string             `yaml:"type,omitempty"`
	IPFS *FireflyIPFSConfig `yaml:"ipfs,omitempty"`
}

type FireflyIPFSConfig struct {
	API     *HttpEndpointConfig `yaml:"api,omitempty"`
	Gateway *HttpEndpointConfig `yaml:"gateway,omitempty"`
}

type TokenConnector struct {
	Plugin string `yaml:"plugin,omitempty"`
	Name   string `yaml:"name,omitempty"`
	URL    string `yaml:"url,omitempty"`
}

type TokensConfig []*TokenConnector

type FireflyConfig struct {
	Log          *LogConfig           `yaml:"log,omitempty"`
	Debug        *HttpServerConfig    `yaml:"debug,omitempty"`
	HTTP         *HttpServerConfig    `yaml:"http,omitempty"`
	Admin        *AdminServerConfig   `yaml:"admin,omitempty"`
	Metrics      *MetricsServerConfig `yaml:"metrics,omitempty"`
	UI           *UIConfig            `yaml:"ui,omitempty"`
	Node         *NodeConfig          `yaml:"node,omitempty"`
	Org          *OrgConfig           `yaml:"org,omitempty"`
	Blockchain   *BlockchainConfig    `yaml:"blockchain,omitempty"`
	Database     *DatabaseConfig      `yaml:"database,omitempty"`
	P2PFS        *PublicStorageConfig `yaml:"publicstorage,omitempty"`
	DataExchange *DataExchangeConfig  `yaml:"dataexchange,omitempty"`
	Tokens       *TokensConfig        `yaml:"tokens,omitempty"`
}

func NewFireflyConfig(stack *types.Stack, member *types.Member) *FireflyConfig {
	memberConfig := &FireflyConfig{
		Log: &LogConfig{
			Level: "debug",
		},
		Debug: &HttpServerConfig{
			Port: 6060,
		},
		HTTP: &HttpServerConfig{
			Port:      member.ExposedFireflyPort,
			Address:   "0.0.0.0",
			PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyPort),
		},
		Admin: &AdminServerConfig{
			HttpServerConfig: HttpServerConfig{
				Port:      member.ExposedFireflyAdminPort,
				Address:   "0.0.0.0",
				PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyAdminPort),
			},
			Enabled: true,
			PreInit: true,
		},
		UI: &UIConfig{
			Path: "./frontend",
		},
		Node: &NodeConfig{
			Name: member.NodeName,
		},
		P2PFS: &PublicStorageConfig{
			Type: "ipfs",
			IPFS: &FireflyIPFSConfig{
				API: &HttpEndpointConfig{
					URL: getIPFSAPIURL(member),
				},
				Gateway: &HttpEndpointConfig{
					URL: getIPFSGatewayURL(member),
				},
			},
		},
		DataExchange: &DataExchangeConfig{
			HTTPS: &HttpEndpointConfig{
				URL: getDataExchangeURL(member),
			},
		},
	}

	if stack.PrometheusEnabled {
		memberConfig.Metrics = &MetricsServerConfig{
			HttpServerConfig: HttpServerConfig{
				Port:      member.ExposedFireflyMetricsPort,
				Address:   "0.0.0.0",
				PublicURL: fmt.Sprintf("http://127.0.0.1:%d", member.ExposedFireflyMetricsPort),
			},
			Enabled: true,
			Path:    "/metrics",
		}
	} else {
		memberConfig.Metrics = &MetricsServerConfig{
			Enabled: false,
		}
	}

	switch stack.Database {
	case "postgres":
		memberConfig.Database = &DatabaseConfig{
			Type: "postgres",
			PostgreSQL: &CommonDBConfig{
				URL: getPostgresURL(member),
				Migrations: &MigrationsConfig{
					Auto: true,
				},
			},
		}
	case "sqlite3":
		memberConfig.Database = &DatabaseConfig{
			Type: stack.Database,
			SQLite3: &CommonDBConfig{
				URL: getSQLitePath(member, stack.Name),
				Migrations: &MigrationsConfig{
					Auto: true,
				},
			},
		}
	}
	return memberConfig
}

func getIPFSAPIURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ipfs_%s:5001", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedIPFSApiPort)
	}
}

func getIPFSGatewayURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://ipfs_%s:8080", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedIPFSGWPort)
	}
}

func getPostgresURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("postgres://postgres:f1refly@postgres_%s:5432?sslmode=disable", member.ID)
	} else {
		return fmt.Sprintf("postgres://postgres:f1refly@127.0.0.1:%v?sslmode=disable", member.ExposedPostgresPort)
	}
}

func getSQLitePath(member *types.Member, stackName string) string {
	if !member.External {
		return "/etc/firefly/db?_busy_timeout=5000"
	} else {
		return path.Join(constants.StacksDir, stackName, "data", member.ID+".db")
	}
}

func getDataExchangeURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://dataexchange_%s:3000", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedDataexchangePort)
	}
}

func ReadFireflyConfig(filePath string) (*FireflyConfig, error) {
	if bytes, err := ioutil.ReadFile(filePath); err != nil {
		return nil, err
	} else {
		var config *FireflyConfig
		err := yaml.Unmarshal(bytes, &config)
		return config, err
	}
}

func WriteFireflyConfig(config *FireflyConfig, filePath string) error {
	if bytes, err := yaml.Marshal(config); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filePath, bytes, 0755)
	}
}
