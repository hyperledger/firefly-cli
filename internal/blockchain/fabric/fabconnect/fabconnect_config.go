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

package fabconnect

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FabconnectConfig struct {
	MaxInFlight     int
	MaxTXWaitTime   int
	SendConcurrency int
	Receipts        *Receipts
	Events          *Events
	HTTP            *HTTP
	RPC             *RPC
}

type Receipts struct {
	MaxDocs           int
	QueryLimit        int
	RetryInitialDelay int
	RetryTimeout      int
	LevelDB           *LevelDB
}

type LevelDB struct {
	Path string
}

type Events struct {
	WebhooksAllowPrivateIPs bool `yaml:"webhooksAllowPrivateIPs,omitempty"`
	LevelDB                 *LevelDB
}

type HTTP struct {
	Port int
}

type RPC struct {
	ConfigPath string
}

func WriteFabconnectConfig(filePath string) error {
	fabconnectConfig := &FabconnectConfig{
		MaxInFlight:     10,
		MaxTXWaitTime:   60,
		SendConcurrency: 25,
		Receipts: &Receipts{
			MaxDocs:           1000,
			QueryLimit:        100,
			RetryInitialDelay: 5,
			RetryTimeout:      30,
			LevelDB: &LevelDB{
				Path: "/fabconnect/receipts",
			},
		},
		Events: &Events{
			WebhooksAllowPrivateIPs: true,
			LevelDB: &LevelDB{
				Path: "/fabconnect/events",
			},
		},
		HTTP: &HTTP{
			Port: 3000,
		},
		RPC: &RPC{
			ConfigPath: "/fabconnect/ccp.yaml",
		},
	}

	fabconnectConfigBytes, _ := yaml.Marshal(fabconnectConfig)
	return os.WriteFile(filePath, fabconnectConfigBytes, 0755)
}
