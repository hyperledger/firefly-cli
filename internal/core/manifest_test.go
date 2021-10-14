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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLatestFireFlyRelease(T *testing.T) {
	gitHubRelease, err := getLatestFireFlyRelease()
	assert.NoError(T, err)
	assert.NotNil(T, gitHubRelease)
	assert.NotEmpty(T, gitHubRelease.TagName)
}

func TestGetFireFlyManifest(T *testing.T) {
	manifest, err := GetReleaseManifest("main")
	assert.NoError(T, err)
	assert.NotNil(T, manifest)
	assert.NotNil(T, manifest.FireFly)
	assert.NotNil(T, manifest.Ethconnect)
	assert.NotNil(T, manifest.Fabconnect)
	assert.NotNil(T, manifest.DataExchange)
	assert.NotNil(T, manifest.Tokens)
}

func TestGetLatestReleaseManifest(T *testing.T) {
	manifest, err := GetLatestReleaseManifest()
	assert.NoError(T, err)
	assert.NotNil(T, manifest)
	assert.NotNil(T, manifest.FireFly)
	assert.NotNil(T, manifest.Ethconnect)
	assert.NotNil(T, manifest.Fabconnect)
	assert.NotNil(T, manifest.DataExchange)
	assert.NotNil(T, manifest.Tokens)
}
