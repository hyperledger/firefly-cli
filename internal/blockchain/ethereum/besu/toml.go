// Copyright © 2022 Kaleido, Inc.
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

package besu

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/pkg/types"
)

func WriteTomlKeyFile(directory string, member *types.Member) error {
	toml := fmt.Sprintf(`
[metadata]
createdAt = 2019-11-05T08:15:30-05:00
description = "File based configuration"

[signing]
type = "file-based-signer"
key-file = "/data/keystore/%s.key"
password-file = "/data/password"
`, member.Address[2:])
	filename := filepath.Join(directory, member.ID, fmt.Sprintf("%s.toml", member.Address[2:]))
	return ioutil.WriteFile(filename, []byte(toml), 0755)
}
