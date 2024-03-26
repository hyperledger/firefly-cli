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

import "fmt"

type GlobalConfig struct {
	ScrapeInterval string `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  string `yaml:"scrape_timeout,omitempty"`
}

type ScrapeConfig struct {
	JobName       string          `yaml:"job_name,omitempty"`
	MetricsPath   string          `yaml:"metrics_path,omitempty"`
	StaticConfigs []*StaticConfig `yaml:"static_configs,omitempty"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets,omitempty"`
}

type PrometheusConfig struct {
	Global        *GlobalConfig   `yaml:"global,omitempty"`
	ScrapeConfigs []*ScrapeConfig `yaml:"scrape_configs,omitempty"`
}

func (s *StackManager) GeneratePrometheusConfig() *PrometheusConfig {
	config := &PrometheusConfig{
		Global: &GlobalConfig{
			ScrapeInterval: "5s",
			ScrapeTimeout:  "5s",
		},
		ScrapeConfigs: []*ScrapeConfig{
			{
				JobName:     "fireflies",
				MetricsPath: "/metrics",
				StaticConfigs: []*StaticConfig{
					{
						Targets: []string{},
					},
				},
			},
		},
	}

	for i, member := range s.Stack.Members {
		config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets, fmt.Sprintf("firefly_core_%d:%d", i, member.ExposedFireflyMetricsPort))

		if s.blockchainProvider.GetConnectorName() == "evmconnect" {
			config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets, fmt.Sprintf("evmconnect_%d:%d", i, member.ExposedConnectorMetricsPort))
		}
	}

	return config
}
