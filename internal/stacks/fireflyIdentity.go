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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
)

func (s *Stack) httpJSONWithRetry(method, url string, body, result interface{}) (err error) {
	retries := 30
	for {
		if err := s.httpJSON(method, url, body, result); err != nil {
			if retries > 0 {
				retries--
				time.Sleep(1 * time.Second)
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}

func (s *Stack) httpJSON(method, url string, body, result interface{}) (err error) {
	if body == nil {
		body = make(map[string]interface{})
	}
	if result == nil {
		result = make(map[string]interface{})
	}

	var requestBody []byte
	if method != http.MethodGet && method != http.MethodDelete {
		requestBody, err = json.Marshal(&body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s returned %d: %s", url, resp.StatusCode, b)
	}

	if resp.StatusCode == 204 {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(&result)
}

func (s *Stack) registerFireflyIdentities(spin *spinner.Spinner, verbose bool) error {
	for _, member := range s.Members {
		orgName := fmt.Sprintf("org_%s", member.ID)
		nodeName := fmt.Sprintf("node_%s", member.ID)
		ffURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1", member.ExposedFireflyPort)
		updateStatus(fmt.Sprintf("registering %s and %s", orgName, nodeName), spin)

		registerOrgURL := fmt.Sprintf("%s/network/register/node/organization", ffURL)
		err := s.httpJSONWithRetry(http.MethodPost, registerOrgURL, nil, nil)
		if err != nil {
			return err
		}

		foundOrg := false
		for !foundOrg {
			type establishedOrg struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			orgURL := fmt.Sprintf("%s/network/organizations", ffURL)
			var orgs []establishedOrg
			err := s.httpJSONWithRetry(http.MethodGet, orgURL, nil, &orgs)
			if err != nil {
				return nil
			}
			for _, o := range orgs {
				foundOrg = foundOrg || o.Name == orgName
			}
			if !foundOrg {
				time.Sleep(1 * time.Second)
			}
		}

		registerNodeURL := fmt.Sprintf("%s/network/register/node", ffURL)
		err = s.httpJSONWithRetry(http.MethodPost, registerNodeURL, nil, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}
