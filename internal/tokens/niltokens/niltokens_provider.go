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

package niltokens

import (
	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type NilTokensProvider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *NilTokensProvider) DeploySmartContracts() error {
	return nil
}

func (p *NilTokensProvider) FirstTimeSetup() error {
	return nil
}

func (p *NilTokensProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	return nil
}

func (p *NilTokensProvider) GetFireflyConfig(m *types.Member) *core.TokensConfig {
	return nil
}
