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

package stacks

import (
	"fmt"
	"net/http"

	"github.com/hyperledger/firefly-cli/internal/core"
)

func (s *StackManager) registerFireflyIdentities() error {
	emptyObject := make(map[string]interface{})

	for _, member := range s.Stack.Members {
		ffURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1", member.ExposedFireflyPort)
		s.Log.Info(fmt.Sprintf("registering org and node for member %s", member.ID))

		registerOrgURL := fmt.Sprintf("%s/network/organizations/self?confirm=true", ffURL)
		err := core.RequestWithRetry(s.ctx, http.MethodPost, registerOrgURL, emptyObject, nil)
		if err != nil {
			return err
		}

		registerNodeURL := fmt.Sprintf("%s/network/nodes/self?confirm=true", ffURL)
		err = core.RequestWithRetry(s.ctx, http.MethodPost, registerNodeURL, emptyObject, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}
