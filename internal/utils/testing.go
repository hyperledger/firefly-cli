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

package utils

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

type TestHelper struct {
	FabricURL     string
	EthConnectURL string
	EvmConnectURL string
}

var (
	FabricEndpoint     = "http://localhost:7054"
	EthConnectEndpoint = "http://localhost:8080"
	EvmConnectEndpoint = "http://localhost:5008"
)

func StartMockServer(t *testing.T) {
	httpmock.Activate()
}

// mockprotocol endpoints for testing
func NewTestEndPoint(t *testing.T) *TestHelper {
	return &TestHelper{
		FabricURL:     FabricEndpoint,
		EthConnectURL: EthConnectEndpoint,
		EvmConnectURL: EvmConnectEndpoint,
	}
}

func StopMockServer(_ *testing.T) {
	httpmock.DeactivateAndReset()
}
