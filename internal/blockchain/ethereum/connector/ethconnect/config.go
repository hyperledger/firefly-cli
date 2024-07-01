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

package ethconnect

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
	Rest *Rest `yaml:"rest,omitempty"`
}

type Rest struct {
	RestGateway *RestGateway `yaml:"rest-gateway,omitempty"`
}

type RestGateway struct {
	RPC           *RPC     `yaml:"rpc,omitempty"`
	OpenAPI       *OpenAPI `yaml:"openapi,omitempty"`
	HTTP          *HTTP    `yaml:"http,omitempty"`
	MaxTXWaitTime int      `yaml:"maxTXWaitTime,omitempty"`
	MaxInFlight   int      `yaml:"maxInFlight,omitempty"`
}

type RPC struct {
	URL string `yaml:"url,omitempty"`
}

type OpenAPI struct {
	EventPollingIntervalSec int    `yaml:"eventPollingIntervalSec,omitempty"`
	StoragePath             string `yaml:"storagePath,omitempty"`
	EventsDB                string `yaml:"eventsDB,omitempty"`
}

type HTTP struct {
	Port int `yaml:"port,omitempty"`
}

func (e *Config) WriteConfig(filename string, extraConnectorConfigPath string) error {
	configYamlBytes, _ := yaml.Marshal(e)
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

func (e *Ethconnect) GenerateConfig(stack *types.Stack, member *types.Organization, blockchainServiceName string) connector.Config {
	return &Config{
		Rest: &Rest{
			RestGateway: &RestGateway{
				MaxTXWaitTime: 60,
				MaxInFlight:   10,
				RPC:           &RPC{URL: fmt.Sprintf("http://%s:8545", blockchainServiceName)},
				OpenAPI: &OpenAPI{
					EventPollingIntervalSec: 1,
					StoragePath:             "./data/abis",
					EventsDB:                "./data/events",
				},
				HTTP: &HTTP{
					Port: 8080,
				},
			},
		},
	}
}
