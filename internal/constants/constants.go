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

package constants

import (
	"os"
	"path/filepath"
)

var StacksDir = checkHome()
var FireFlyCoreImageName = "ghcr.io/hyperledger/firefly"
var IPFSImageName = "ipfs/go-ipfs:v0.10.0"
var PostgresImageName = "postgres"
var PrometheusImageName = "prom/prometheus"
var SandboxImageName = "ghcr.io/hyperledger/firefly-sandbox:latest"

func checkHome() string {
	var homeDir, _ = os.UserHomeDir()
	var StacksDir = filepath.Join(homeDir, ".firefly", "stacks")
	var fireflyhome, present = os.LookupEnv("FIREFLY_HOME")
	if present {
		StacksDir = filepath.Join(fireflyhome, "stacks")
	}
	return StacksDir
}
