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

package fabconnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type CreateIdentityRequest struct {
	Name string
	Type string
}

type CreateIdentityResponse struct {
	Name   string
	Secret string
}

type EnrollIdentityRequest struct {
	Secret string
}

type EnrollIdentityResponse struct {
	Name    string
	Success bool
}

func CreateIdentity(fabconnectURL string, signer string) (*CreateIdentityResponse, error) {
	u, err := url.Parse(fabconnectURL)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("identities"))
	if err != nil {
		return nil, err
	}
	requestURL := u.String()
	requestBody, err := json.Marshal(&CreateIdentityRequest{Name: signer, Type: "client"})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var createIdentityResponseBody *CreateIdentityResponse
	if err := json.Unmarshal(responseBody, &createIdentityResponseBody); err != nil {
		return nil, err
	}
	return createIdentityResponseBody, nil
}

func EnrollIdentity(fabconnectURL, signer, secret string) (*EnrollIdentityResponse, error) {
	u, err := url.Parse(fabconnectURL)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("identities", signer, "enroll"))
	if err != nil {
		return nil, err
	}
	requestURL := u.String()
	requestBody, err := json.Marshal(&EnrollIdentityRequest{Secret: secret})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s [%d] %s", req.URL, resp.StatusCode, responseBody)
	}
	var enrollIdentityResponse *EnrollIdentityResponse
	if err := json.Unmarshal(responseBody, &enrollIdentityResponse); err != nil {
		return nil, err
	}
	return enrollIdentityResponse, nil
}
