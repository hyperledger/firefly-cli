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

package types

type Stack struct {
	Name                  string           `json:"name,omitempty"`
	Members               []*Member        `json:"members,omitempty"`
	SwarmKey              string           `json:"swarmKey,omitempty"`
	ExposedBlockchainPort int              `json:"exposedGethPort,omitempty"`
	Database              string           `json:"database"`
	BlockchainProvider    string           `json:"blockchainProvider"`
	TokenProviders        TokenProviders   `json:"tokenProviders"`
	VersionManifest       *VersionManifest `json:"versionManifest,omitempty"`
	PrometheusEnabled     bool             `json:"prometheusEnabled,omitempty"`
	ExposedPrometheusPort int              `json:"exposedPrometheusPort,omitempty"`
	ContractAddress       string           `json:"contractAddress,omitempty"`
	InitDir               string           `json:-`
	RuntimeDir            string           `json:-`
	StackDir              string           `json:-`
}

type Member struct {
	ID                        string `json:"id,omitempty"`
	Index                     *int   `json:"index,omitempty"`
	Address                   string `json:"address,omitempty"`
	PrivateKey                string `json:"privateKey,omitempty"`
	ExposedFireflyPort        int    `json:"exposedFireflyPort,omitempty"`
	ExposedFireflyAdminPort   int    `json:"exposedFireflyAdminPort,omitempty"`
	ExposedFireflyMetricsPort int    `json:"exposedFireflyMetricsPort,omitempty"`
	ExposedConnectorPort      int    `json:"exposedConnectorPort,omitempty"`
	ExposedPostgresPort       int    `json:"exposedPostgresPort,omitempty"`
	ExposedDataexchangePort   int    `json:"exposedDataexchangePort,omitempty"`
	ExposedIPFSApiPort        int    `json:"exposedIPFSApiPort,omitempty"`
	ExposedIPFSGWPort         int    `json:"exposedIPFSGWPort,omitempty"`
	ExposedUIPort             int    `json:"exposedUiPort,omitempty"`
	ExposedTokensPorts        []int  `json:"exposedTokensPorts,omitempty"`
	External                  bool   `json:"external,omitempty"`
	OrgName                   string `json:"orgName,omitempty"`
	NodeName                  string `json:"nodeName,omitempty"`
}
