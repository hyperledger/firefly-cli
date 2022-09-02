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

package stacks

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSwarmKey() string {
	key := make([]byte, 32)
	rand.Read(key)
	hexKey := hex.EncodeToString(key)
	return "/key/swarm/psk/1.0.0/\n/base16/\n" + hexKey
}
