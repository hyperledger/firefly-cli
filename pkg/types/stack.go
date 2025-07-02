// Copyright © 2025 Kaleido, Inc.
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
	"github.com/hyperledger/firefly-common/pkg/fftypes"
)

type Stack struct {
	Name                      string                 `json:"name,omitempty"`
	Members                   []*Organization        `json:"members,omitempty"`
	SwarmKey                  string                 `json:"swarmKey,omitempty"`
	ExposedBlockchainPort     int                    `json:"exposedBlockchainPort,omitempty"`
	ExposedPtmPort            int                    `json:"exposedPtmPort,omitempty"`
	Database                  fftypes.FFEnum         `json:"database"`
	BlockchainProvider        fftypes.FFEnum         `json:"blockchainProvider"`
	BlockchainConnector       fftypes.FFEnum         `json:"blockchainConnector"`
	BlockchainNodeProvider    fftypes.FFEnum         `json:"blockchainNodeProvider"`
	PrivateTransactionManager fftypes.FFEnum         `json:"privateTransactionManager"`
	Consensus                 fftypes.FFEnum         `json:"consensus"`
	TokenProviders            []fftypes.FFEnum       `json:"tokenProviders"`
	VersionManifest           *VersionManifest       `json:"versionManifest,omitempty"`
	PrometheusEnabled         bool                   `json:"prometheusEnabled,omitempty"`
	SandboxEnabled            bool                   `json:"sandboxEnabled,omitempty"`
	MultipartyEnabled         bool                   `json:"multiparty"`
	ExposedPrometheusPort     int                    `json:"exposedPrometheusPort,omitempty"`
	ContractAddress           string                 `json:"contractAddress,omitempty"`
	ChainIDPtr                *int64                 `json:"chainID,omitempty"`
	Network                   string                 `json:"network,omitempty"`
	Socket                    string                 `json:"socket,omitempty"`
	BlockfrostKey             string                 `json:"blockfrostKey,omitempty"`
	BlockfrostBaseURL         string                 `json:"blockfrostBaseURL,omitempty"`
	RemoteNodeURL             string                 `json:"remoteNodeURL,omitempty"`
	DisableTokenFactories     bool                   `json:"disableTokenFactories,omitempty"`
	RequestTimeout            int                    `json:"requestTimeout,omitempty"`
	IPFSMode                  fftypes.FFEnum         `json:"ipfsMode"`
	RemoteFabricNetwork       bool                   `json:"remoteFabricNetwork,omitempty"`
	ChannelName               string                 `json:"channelName,omitempty"`
	ChaincodeName             string                 `json:"chaincodeName,omitempty"`
	CustomPinSupport          bool                   `json:"customPinSupport,omitempty"`
	RemoteNodeDeploy          bool                   `json:"remoteNodeDeploy,omitempty"`
	EnvironmentVars           map[string]interface{} `json:"environmentVars"`
	InitDir                   string                 `json:"-"`
	RuntimeDir                string                 `json:"-"`
	StackDir                  string                 `json:"-"`
	State                     *StackState            `json:"-"`
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
		switch {
		case os.IsNotExist(err):
			return false, nil
		case err != nil:
			return false, err
		default:
			return true, nil
		}
	} else {
		runtimeDir := filepath.Join(stackDir, "runtime")
		_, err := os.Stat(runtimeDir)
		switch {
		case os.IsNotExist(err):
			return false, nil
		case err != nil:
			return false, err
		default:
			return true, nil
		}
	}
}

func (s *Stack) IsOldFileStructure() (bool, error) {
	stackDir := filepath.Join(constants.StacksDir, s.Name)
	_, err := os.Stat(filepath.Join(stackDir, "init"))
	switch {
	case os.IsNotExist(err):
		return true, nil
	case err != nil:
		return false, err
	default:
		return false, nil
	}
}

func (s *Stack) ConcatenateWithProvidedEnvironmentVars(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}
	for k, v := range s.EnvironmentVars {
		result[k] = v // Overwrites existing keys from previous map
	}
	return result
}
