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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func GetManifestForReleaseChannel(releaseChannel types.ReleaseChannelSelection) (*types.VersionManifest, error) {
	dockerTag := types.ReleaseChannelSelectionStrings[releaseChannel]
	if releaseChannel == types.Stable {
		dockerTag = "latest"
	}
	gitCommit, err := docker.GetImageLabel(fmt.Sprintf("%s:%s", constants.FireFlyCoreImageName, dockerTag), "commit")
	if err != nil {
		return nil, err
	}
	return GetReleaseManifest(gitCommit)
}

func GetReleaseManifest(version string) (*types.VersionManifest, error) {
	manifest := &types.VersionManifest{}
	if err := request("GET", fmt.Sprintf("https://raw.githubusercontent.com/hyperledger/firefly/%s/manifest.json", version), nil, &manifest); err != nil {
		return nil, err
	}
	if manifest.FireFly == nil {
		// Fill in the FireFly version number
		manifest.FireFly = &types.ManifestEntry{
			Image: "ghcr.io/hyperledger/firefly",
			Tag:   version,
		}
	}
	return manifest, nil
}

func ReadManifestFile(p string) (*types.VersionManifest, error) {
	d, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var manifest *types.VersionManifest
	err = json.Unmarshal(d, &manifest)

	// If core is not specified in the manifest, use a locally built image called "firefly"
	if manifest.FireFly == nil {
		manifest.FireFly = &types.ManifestEntry{
			Image: "hyperledger/firefly",
			Local: true,
		}
	}
	return manifest, err
}
