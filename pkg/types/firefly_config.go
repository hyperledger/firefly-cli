// Copyright © 2025 Kaleido, Inc.
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

package types

type LogConfig struct {
	Level string `yaml:"level,omitempty"`
}

type HTTPServerConfig struct {
	Port      int    `yaml:"port,omitempty"`
	Address   string `yaml:"address,omitempty"`
	PublicURL string `yaml:"publicURL,omitempty"`
}

type AdminServerConfig struct {
	HTTPServerConfig `yaml:",inline"`
	Enabled          bool `yaml:"enabled,omitempty"`
	PreInit          bool `yaml:"preinit,omitempty"`
}

type SPIServerConfig struct {
	HTTPServerConfig `yaml:",inline"`
	Enabled          bool `yaml:"enabled,omitempty"`
}

type MetricsServerConfig struct {
	HTTPServerConfig `yaml:",inline"`
	Enabled          bool   `yaml:"enabled,omitempty"`
	Path             string `yaml:"path,omitempty"`
}

type BasicAuth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type HTTPEndpointConfig struct {
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
	Name string `yaml:"name,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

type CardanoconnectConfig struct {
	URL   string `yaml:"url,omitempty"`
	Topic string `yaml:"topic,omitempty"`
}

type EthconnectConfig struct {
	URL   string     `yaml:"url,omitempty"`
	Topic string     `yaml:"topic,omitempty"`
	Auth  *BasicAuth `yaml:"auth,omitempty"`
}

type TezosconnectConfig struct {
	URL   string     `yaml:"url,omitempty"`
	Topic string     `yaml:"topic,omitempty"`
	Auth  *BasicAuth `yaml:"auth,omitempty"`
}

type FabconnectConfig struct {
	URL       string `yaml:"url,omitempty"`
	Channel   string `yaml:"channel,omitempty"`
	Chaincode string `yaml:"chaincode,omitempty"`
	Topic     string `yaml:"topic,omitempty"`
	Signer    string `yaml:"signer,omitempty"`
}

type CardanoConfig struct {
	Cardanoconnect *CardanoconnectConfig `yaml:"cardanoconnect,omitempty"`
}

type EthereumConfig struct {
	Ethconnect *EthconnectConfig `yaml:"ethconnect,omitempty"`
}

type TezosConfig struct {
	Tezosconnect *TezosconnectConfig `yaml:"tezosconnect,omitempty"`
}

type FabricConfig struct {
	Fabconnect *FabconnectConfig `yaml:"fabconnect,omitempty"`
}

type BlockchainConfig struct {
	Name     string          `yaml:"name,omitempty"`
	Type     string          `yaml:"type,omitempty"`
	Cardano  *CardanoConfig  `yaml:"cardano,omitempty"`
	Ethereum *EthereumConfig `yaml:"ethereum,omitempty"`
	Tezos    *TezosConfig    `yaml:"tezos,omitempty"`
	Fabric   *FabricConfig   `yaml:"fabric,omitempty"`
}

type DataExchangeConfig struct {
	Name string              `yaml:"name,omitempty"`
	Type string              `yaml:"type,omitempty"`
	FFDX *HTTPEndpointConfig `yaml:"ffdx,omitempty"`
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
	Name       string          `yaml:"name,omitempty"`
	Type       string          `yaml:"type,omitempty"`
	PostgreSQL *CommonDBConfig `yaml:"postgres,omitempty"`
	SQLite3    *CommonDBConfig `yaml:"sqlite3,omitempty"`
}

type SharedStorageConfig struct {
	Name string             `yaml:"name,omitempty"`
	Type string             `yaml:"type,omitempty"`
	IPFS *FireflyIPFSConfig `yaml:"ipfs,omitempty"`
}

type FireflyIPFSConfig struct {
	API     *HTTPEndpointConfig `yaml:"api,omitempty"`
	Gateway *HTTPEndpointConfig `yaml:"gateway,omitempty"`
}

type TokensConfig struct {
	Type     string          `yaml:"type,omitempty"`
	Name     string          `yaml:"name,omitempty"`
	FFTokens *FFTokensConfig `yaml:"fftokens,omitempty"`
}

type FFTokensConfig struct {
	URL string `yaml:"url,omitempty"`
}

type DBEventsConfig struct {
	BufferSize int `yaml:"bufferSize,omitempty"`
}

type EventConfig struct {
	DBEvents *DBEventsConfig `yaml:"dbevents,omitempty"`
}

type NamespacesConfig struct {
	Default    string       `json:"default"`
	Predefined []*Namespace `json:"predefined"`
}

type FireflyConfig struct {
	Log        *LogConfig           `yaml:"log,omitempty"`
	Debug      *HTTPServerConfig    `yaml:"debug,omitempty"`
	HTTP       *HTTPServerConfig    `yaml:"http,omitempty"`
	Admin      *AdminServerConfig   `yaml:"admin,omitempty"` // V1.0 admin API
	SPI        *SPIServerConfig     `yaml:"spi,omitempty"`   // V1.1 and later SPI
	Metrics    *MetricsServerConfig `yaml:"metrics,omitempty"`
	UI         *UIConfig            `yaml:"ui,omitempty"`
	Event      *EventConfig         `yaml:"event,omitempty"`
	Plugins    *Plugins             `yaml:"plugins"`
	Namespaces *NamespacesConfig    `yaml:"namespaces"`
}
