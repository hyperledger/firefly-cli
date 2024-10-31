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

package types

import "fmt"

type GitHubRelease struct {
	TagName string `json:"tag_name,omitempty"`
}

type VersionManifest struct {
	FireFly           *ManifestEntry `json:"firefly,omitempty"`
	Cardanoconnect    *ManifestEntry `json:"cardanoconnect"`
	Cardanosigner     *ManifestEntry `json:"cardanoconnect"`
	Ethconnect        *ManifestEntry `json:"ethconnect"`
	Evmconnect        *ManifestEntry `json:"evmconnect"`
	Tezosconnect      *ManifestEntry `json:"tezosconnect"`
	Fabconnect        *ManifestEntry `json:"fabconnect"`
	DataExchange      *ManifestEntry `json:"dataexchange-https"`
	TokensERC1155     *ManifestEntry `json:"tokens-erc1155"`
	TokensERC20ERC721 *ManifestEntry `json:"tokens-erc20-erc721"`
	Signer            *ManifestEntry `json:"signer"`
}

func (m *VersionManifest) Entries() []*ManifestEntry {
	if m == nil {
		return []*ManifestEntry{}
	}
	return []*ManifestEntry{
		m.FireFly,
		m.Cardanoconnect,
		m.Ethconnect,
		m.Evmconnect,
		m.Tezosconnect,
		m.Fabconnect,
		m.DataExchange,
		m.TokensERC1155,
		m.TokensERC20ERC721,
		m.Signer,
	}
}

type ManifestEntry struct {
	Image string `json:"image,omitempty"`
	Local bool   `json:"local,omitempty"`
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
