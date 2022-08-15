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

import (
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/constants"
)

type Stack struct {
	Name                   string           `json:"name,omitempty"`
	Members                []*Organization  `json:"members,omitempty"`
	SwarmKey               string           `json:"swarmKey,omitempty"`
	ExposedBlockchainPort  int              `json:"exposedBlockchainPort,omitempty"`
	Database               string           `json:"database"`
	BlockchainProvider     string           `json:"blockchainProvider"`
	BlockchainConnector    string           `json:"blockchainConnector"`
	BlockchainNodeProvider string           `json:"blockchainNodeProvider"`
	TokenProviders         TokenProviders   `json:"tokenProviders"`
	VersionManifest        *VersionManifest `json:"versionManifest,omitempty"`
	PrometheusEnabled      bool             `json:"prometheusEnabled,omitempty"`
	SandboxEnabled         bool             `json:"sandboxEnabled,omitempty"`
	MultipartyEnabled      bool             `json:"multiparty"`
	ExposedPrometheusPort  int              `json:"exposedPrometheusPort,omitempty"`
	ContractAddress        string           `json:"contractAddress,omitempty"`
	ChainIDPtr             *int64           `json:"chainID,omitempty"`
	RemoteNodeURL          string           `json:"remoteNodeURL,omitempty"`
	DisableTokenFactories  bool             `json:"disableTokenFactories,omitempty"`
	RequestTimeout         int              `json:"requestTimeout,omitempty"`
	InitDir                string           `json:"-"`
	RuntimeDir             string           `json:"-"`
	StackDir               string           `json:"-"`
	State                  *StackState      `json:"-"`
}

func (s *Stack) ChainID() int64 {
	if s.ChainIDPtr == nil {
		return 2021 // the original default, before it could be customized
	}
	return *s.ChainIDPtr
}

func (s *Stack) HasRunBefore() (bool, error) {
	stackDir := filepath.Join(constants.StacksDir, s.Name)
	isOldFileStructure, err := s.IsOldFileStructure()
	if err != nil {
		return false, err
	}
	if isOldFileStructure {
		dataDir := filepath.Join(stackDir, "data")
		_, err := os.Stat(dataDir)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		runtimeDir := filepath.Join(stackDir, "runtime")
		_, err := os.Stat(runtimeDir)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

func (s *Stack) IsOldFileStructure() (bool, error) {
	stackDir := filepath.Join(constants.StacksDir, s.Name)
	_, err := os.Stat(filepath.Join(stackDir, "init"))
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}
}
