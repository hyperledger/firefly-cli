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

package stacks

import (
	"fmt"
)

type DataExchangeListenerConfig struct {
	Hostname string `json:"hostname,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Port     int    `json:"port,omitempty"`
}

type PeerConfig struct {
	ID       string `json:"id,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type DataExchangePeerConfig struct {
	API   *DataExchangeListenerConfig `json:"api,omitempty"`
	P2P   *DataExchangeListenerConfig `json:"p2p,omitempty"`
	Peers []*PeerConfig               `json:"peers"`
}

func (s *StackManager) GenerateDataExchangeHTTPSConfig(memberID string) *DataExchangePeerConfig {
	return &DataExchangePeerConfig{
		API: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3000,
		},
		P2P: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3001,
			Endpoint: fmt.Sprintf("https://dataexchange_%s:3001", memberID),
		},
		Peers: []*PeerConfig{},
	}
}
