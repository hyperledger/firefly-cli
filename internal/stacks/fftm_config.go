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
	Address string `json:"address"`
	Port    int64  `json:"port"`
}

type FFTMConnectorConfig struct {
	URL     string `json:"url"`
	Variant string `json:"variant,omitempty"`
}

type FFTMFFCoreConfig struct {
	URL string `json:"url"`
}

type FFTMManagerConfig struct {
	Name string `json:"name"`
}

type FFTMPolicyEngineConfig struct {
	Name   string                       `json:"name"`
	Simple FFTMSimplePolicyEngineConfig `json:"simple"`
}

type FFTMSimplePolicyEngineConfig struct {
	FixedGasPrice string
}

type FFTMConfirmationsConfig struct {
	Required int64 `json:"required"`
}

type FFTMConfig struct {
	Log           FFTMLogConfig           `json:"log"`
	API           FFTMAPIConfig           `json:"api"`
	Connector     FFTMConnectorConfig     `json:"connector"`
	FFCore        FFTMFFCoreConfig        `json:"ffcore"`
	Manager       FFTMManagerConfig       `json:"manager"`
	PolicyEngine  FFTMPolicyEngineConfig  `json:"policyengine"`
	Confirmations FFTMConfirmationsConfig `json:"confirmations"`
}

func NewFFTMConfig(stack *types.Stack, member *types.Member) *FFTMConfig {
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
			URL: fmt.Sprintf("http://firefly_core_%s:%d", member.ID, member.ExposedFireflyAdminPort),
		},
		Connector: FFTMConnectorConfig{
			URL:     fmt.Sprintf("http://ethconnect_%s:8080", member.ID),
			Variant: "evm",
		},
		PolicyEngine: FFTMPolicyEngineConfig{
			Name: "simple",
			Simple: FFTMSimplePolicyEngineConfig{
				FixedGasPrice: "2000000000",
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
