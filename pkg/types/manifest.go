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

import "fmt"

type GitHubRelease struct {
	TagName string `json:"tag_name,omitempty"`
}

type VersionManifest struct {
	FireFly      *ManifestEntry `json:"firefly,omitempty"`
	Ethconnect   *ManifestEntry `json:"ethconnect"`
	Fabconnect   *ManifestEntry `json:"fabconnect"`
	DataExchange *ManifestEntry `json:"dataexchange-https"`
	Tokens       *ManifestEntry `json:"tokens-erc1155"`
}

type ManifestEntry struct {
	Image string `json:"image,omitempty"`
	Tag   string `json:"tag,omitempty"`
	SHA   string `json:"sha,omitempty"`
}

func (m *ManifestEntry) GetDockerImageString() string {
	if m.SHA != "" {
		return fmt.Sprintf("%s@sha256:%s", m.Image, m.SHA)
	} else if m.Tag != "" {
		return fmt.Sprintf("%s:%s", m.Image, m.Tag)
	}
	return m.Image
}
