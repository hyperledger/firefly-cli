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
	"github.com/hyperledger/firefly-common/pkg/fftypes"
)

func GetManifestForReleaseChannel(releaseChannel fftypes.FFEnum) (*types.VersionManifest, error) {
	dockerTag := releaseChannel.String()
	if releaseChannel == types.ReleaseChannelStable {
		dockerTag = "latest"
	}

	imageName := fmt.Sprintf("%s:%s", constants.FireFlyCoreImageName, dockerTag)

	gitCommit, err := docker.GetImageLabel(imageName, "commit")
	if err != nil {
		return nil, err
	}

	imageDigest, err := docker.GetImageDigest(imageName)
	if err != nil {
		return nil, err
	}

	manifest, err := GetReleaseManifest(gitCommit)
	if err != nil {
		return nil, err
	}

	if manifest.FireFly == nil {
		// Fill in the FireFly version number
		manifest.FireFly = &types.ManifestEntry{
			Image: "ghcr.io/hyperledger/firefly",
			SHA:   imageDigest[7:],
		}
	}
	return manifest, nil
}

func GetReleaseManifest(version string) (*types.VersionManifest, error) {
	manifest := &types.VersionManifest{}
	if err := request("GET", fmt.Sprintf("https://raw.githubusercontent.com/hyperledger/firefly/%s/manifest.json", version), nil, &manifest); err != nil {
		return nil, err
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
