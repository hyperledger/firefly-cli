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

package stacks

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/miracl/conflate"
	"gopkg.in/yaml.v2"
)

type FFTMLogConfig struct {
	Level string `yaml:"level,omitempty"`
}

type FFTMAPIConfig struct {
	Address string `yaml:"address,omitempty"`
	Port    int64  `yaml:"port,omitempty"`
}

type FFTMConnectorConfig struct {
	URL     string `yaml:"url,omitempty"`
	Variant string `yaml:"variant,omitempty"`
}

type FFTMFFCoreConfig struct {
	URL string `yaml:"url,omitempty"`
}

type FFTMManagerConfig struct {
	Name string `yaml:"name,omitempty"`
}

type FFTMPolicyEngineConfig struct {
	Name   string                       `yaml:"name,omitempty"`
	Simple FFTMSimplePolicyEngineConfig `yaml:"simple,omitempty"`
}

type FFTMSimplePolicyEngineGasOracleConfig struct {
	Mode string `yaml:"mode,omitempty"`
}

type FFTMSimplePolicyEngineConfig struct {
	FixedGasPrice string                                `yaml:"fixedGasPrice,omitempty"`
	GasOracle     FFTMSimplePolicyEngineGasOracleConfig `yaml:"gasOracle,omitempty"`
}

type FFTMConfirmationsConfig struct {
	Required int64 `yaml:"required,omitempty"`
}

type FFTMConfig struct {
	Log           FFTMLogConfig           `yaml:"log"`
	API           FFTMAPIConfig           `yaml:"api"`
	Connector     FFTMConnectorConfig     `yaml:"connector"`
	FFCore        FFTMFFCoreConfig        `yaml:"ffcore"`
	Manager       FFTMManagerConfig       `yaml:"manager"`
	PolicyEngine  FFTMPolicyEngineConfig  `yaml:"policyengine"`
	Confirmations FFTMConfirmationsConfig `yaml:"confirmations"`
}

func NewFFTMConfig(stack *types.Stack, member *types.Organization) *FFTMConfig {
	return &FFTMConfig{
		Log: FFTMLogConfig{
			Level: "debug",
		},
		Manager: FFTMManagerConfig{
			Name: fmt.Sprintf("fftm_%s", member.ID),
		},
		API: FFTMAPIConfig{
			Address: "0.0.0.0",
			Port:    5008,
		},
		FFCore: FFTMFFCoreConfig{
			URL: fmt.Sprintf("http://firefly_core_%s:%d", member.ID, member.ExposedFireflyAdminSPIPort),
		},
		Connector: FFTMConnectorConfig{
			URL:     fmt.Sprintf("http://ethconnect_%s:8080", member.ID),
			Variant: "evm",
		},
		PolicyEngine: FFTMPolicyEngineConfig{
			Name: "simple",
			Simple: FFTMSimplePolicyEngineConfig{
				GasOracle: FFTMSimplePolicyEngineGasOracleConfig{
					Mode: "connector",
				},
			},
		},
		Confirmations: FFTMConfirmationsConfig{
			Required: 3,
		},
	}
}

func WriteFFTMConfig(config *FFTMConfig, filePath string, extraCoreConfigPath string) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, bytes, 0755); err != nil {
		return err
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
		if err := ioutil.WriteFile(filePath, bytes, 0755); err != nil {
			return err
		}
	}
	return nil
}
