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

package fabric

import (
	"encoding/json"
	"io/ioutil"
)

type Template struct {
	Count int `yaml:Count,omitempty`
}

type Users struct {
	Count int `yaml:Count,omitempty`
}

type CA struct {
	Hostname           string `yaml:"Hostname,omitempty"`
	Country            string `yaml:"Country,omitempty"`
	Province           string `yaml:"Province,omitempty"`
	Locality           string `yaml:"Locality,omitempty"`
	OrganizationalUnit string `yaml:"OrganizationalUnit,omitempty"`
}

type Spec struct {
	Hostname string `yaml:"Hostname,omitempty"`
}

type Org struct {
	Name          string    `yaml:"Orderer,omitempty"`
	Domain        string    `yaml:"Domain,omitempty"`
	EnableNodeOUs bool      `yaml:"EnableNodeOUs,omitempty"`
	Specs         []*Spec   `yaml:"Specs,omitempty`
	CA            *CA       `yaml:"CA,omitempty"`
	Template      *Template `yaml:"Template,omitempty"`
	Users         *Users    `yaml:"Users,omitempty"`
}

type CryptogenConfig struct {
	OrdererOrgs []*Org `yaml:"OrdererOrgs,omitempty"`
	PeerOrgs    []*Org `yaml:"PeerOrgs,omitempty"`
}

func WriteCryptogenConfig(memberCount int, path string) error {
	cryptogenConfig := &CryptogenConfig{
		OrdererOrgs: []*Org{
			&Org{
				Name:          "Orderer",
				Domain:        "example.com",
				EnableNodeOUs: false,
			},
		},
		PeerOrgs: []*Org{
			&Org{
				Name:          "Org1",
				Domain:        "org1.example.com",
				EnableNodeOUs: false,
				CA: &CA{
					Hostname:           "ca",
					Country:            "US",
					Province:           "California",
					Locality:           "San Francisco",
					OrganizationalUnit: "Hyperledger Fabric",
				},
				Template: &Template{
					Count: 1,
				},
				Users: &Users{
					Count: memberCount,
				},
			},
		},
	}

	cryptogenConfigBytes, _ := json.MarshalIndent(cryptogenConfig, "", " ")
	return ioutil.WriteFile(path, cryptogenConfigBytes, 0755)
}
