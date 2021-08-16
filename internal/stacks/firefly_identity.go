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
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger-labs/firefly-cli/internal/core"
)

func (s *StackManager) registerFireflyIdentities(verbose bool) error {
	emptyObject := make(map[string]interface{})

	for _, member := range s.Stack.Members {
		orgName := fmt.Sprintf("org_%s", member.ID)
		nodeName := fmt.Sprintf("node_%s", member.ID)
		ffURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1", member.ExposedFireflyPort)
		s.Log.Info(fmt.Sprintf("registering %s and %s", orgName, nodeName))

		registerOrgURL := fmt.Sprintf("%s/network/register/node/organization", ffURL)
		err := core.RequestWithRetry(http.MethodPost, registerOrgURL, emptyObject, nil)
		if err != nil {
			return err
		}

		foundOrg := false
		retries := 60
		for !foundOrg {
			type establishedOrg struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			orgURL := fmt.Sprintf("%s/network/organizations", ffURL)
			var orgs []establishedOrg
			err := core.RequestWithRetry(http.MethodGet, orgURL, nil, &orgs)
			if err != nil {
				return nil
			}
			for _, o := range orgs {
				foundOrg = foundOrg || o.Name == orgName
			}
			if !foundOrg && retries > 0 {
				time.Sleep(1 * time.Second)
				retries--
			} else if !foundOrg && retries == 0 {
				return fmt.Errorf("timeout error waiting to register %s and %s", orgName, nodeName)
			}
		}

		registerNodeURL := fmt.Sprintf("%s/network/register/node", ffURL)
		err = core.RequestWithRetry(http.MethodPost, registerNodeURL, emptyObject, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}
