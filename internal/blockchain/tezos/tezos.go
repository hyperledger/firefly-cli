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

package tezos

import tz "blockwatch.cc/tzgo/tezos"

type Account struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func GenerateAddressAndPrivateKey() (address string, privateKey string, err error) {
	prk, err := tz.GenerateKey(tz.KeyTypeEd25519)
	if err != nil {
		return "", "", err
	}

	return prk.Address().String(), prk.String(), nil
}
