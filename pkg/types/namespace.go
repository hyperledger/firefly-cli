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

package types

type Namespace struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Plugins     []string          `yaml:"plugins"`
	Multiparty  *MultipartyConfig `yaml:"multiparty,omitempty"`
	DefaultKey  interface{}       `yaml:"defaultKey"`
}

type Plugins struct {
	Database      []*DatabaseConfig      `yaml:"database,omitempty"`
	Blockchain    []*BlockchainConfig    `yaml:"blockchain,omitempty"`
	SharedStorage []*SharedStorageConfig `yaml:"sharedstorage,omitempty"`
	DataExchange  []*DataExchangeConfig  `yaml:"dataexchange,omitempty"`
	Tokens        []*TokensConfig        `yaml:"tokens,omitempty"`
}

type MultipartyConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Org      *OrgConfig        `yaml:"org"`
	Node     *NodeConfig       `yaml:"node"`
	Contract []*ContractConfig `yaml:"contract"`
}

type ContractConfig struct {
	Location   interface{} `yaml:"location"`
	FirstEvent string      `yaml:"firstEvent,omitempty"`
	Options    interface{} `yaml:"options"`
}

type MultipartyOrgConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Key         string `yaml:"key"`
}
