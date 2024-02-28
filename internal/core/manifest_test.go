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

package core

import (
	"testing"

	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGetFireFlyManifest(t *testing.T) {
	manifest, err := GetManifestForRelease("main")
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.NotNil(t, manifest.Ethconnect)
	assert.NotNil(t, manifest.Fabconnect)
	assert.NotNil(t, manifest.DataExchange)
	assert.NotNil(t, manifest.TokensERC1155)
	assert.NotNil(t, manifest.TokensERC20ERC721)
}

func TestGetLatestReleaseManifest(t *testing.T) {
	manifest, err := GetManifestForChannel(types.ReleaseChannelStable)
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.NotNil(t, manifest.FireFly)
	assert.NotNil(t, manifest.Ethconnect)
	assert.NotNil(t, manifest.Fabconnect)
	assert.NotNil(t, manifest.DataExchange)
	assert.NotNil(t, manifest.TokensERC1155)
	assert.NotNil(t, manifest.TokensERC20ERC721)
}

func TestIsSupportedVersionUpgrade(t *testing.T) {
	assert.NoError(t, ValidateVersionUpgrade("v1.2.1", "v1.2.2"))
	assert.NoError(t, ValidateVersionUpgrade("v1.2.0", "v1.2.2"))
	assert.NoError(t, ValidateVersionUpgrade("1.2.1", "v1.2.2"))
	assert.NoError(t, ValidateVersionUpgrade("v1.2.1", "1.2.2"))

	assert.Error(t, ValidateVersionUpgrade("v1.2.2", "v1.3.0"))
	assert.Error(t, ValidateVersionUpgrade("latest", "v1.3.0"))
	assert.Error(t, ValidateVersionUpgrade("v1.2.2", "latest"))
}
