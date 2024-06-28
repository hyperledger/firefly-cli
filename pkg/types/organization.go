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

type Organization struct {
	ID                          string       `json:"id,omitempty"`
	Index                       *int         `json:"index,omitempty"`
	Account                     interface{}  `json:"account,omitempty"`
	ExposedFireflyPort          int          `json:"exposedFireflyPort,omitempty"`
	ExposedFireflyAdminSPIPort  int          `json:"exposedFireflyAdminPort,omitempty"` // stack.json still contains the word "Admin" (rather than SPI) for migration
	ExposedFireflyMetricsPort   int          `json:"exposedFireflyMetricsPort,omitempty"`
	ExposedConnectorPort        int          `json:"exposedConnectorPort,omitempty"`
	ExposedConnectorMetricsPort int          `json:"exposedConnectorMetricsPort,omitempty"`
	ExposedDatabasePort         int          `json:"exposedPostgresPort,omitempty"`
	ExposedDataexchangePort     int          `json:"exposedDataexchangePort,omitempty"`
	ExposedIPFSApiPort          int          `json:"exposedIPFSApiPort,omitempty"`
	ExposedIPFSGWPort           int          `json:"exposedIPFSGWPort,omitempty"`
	ExposedUIPort               int          `json:"exposedUiPort,omitempty"`
	ExposedSandboxPort          int          `json:"exposedSandboxPort,omitempty"`
	ExposedTokensPorts          []int        `json:"exposedTokensPorts,omitempty"`
	ExposePtmTpPort             int          `json:"exposePtmTpPort,omitempty"`
	External                    bool         `json:"external,omitempty"`
	OrgName                     string       `json:"orgName,omitempty"`
	NodeName                    string       `json:"nodeName,omitempty"`
	Namespaces                  []*Namespace `json:"namespaces"`
}
