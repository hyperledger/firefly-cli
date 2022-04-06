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
// See the License for the specific lan

package types

type DeployedContract struct {
	Name     string      `json:"name"`
	Location interface{} `json:"location"`
}

type Identity struct {
	PrivateKey string `json:"privateKey"`
	Address    string `json:"address"`
}

type StackState struct {
	DeployedContracts []*DeployedContract `json:"deployedContracts"`
	Identities        []*Identity         `json:"identities"`
}
