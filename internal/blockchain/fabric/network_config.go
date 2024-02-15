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

package fabric

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Registrar struct {
	EnrollID     string `yaml:"enrollId,omitempty"`
	EnrollSecret string `yaml:"enrollSecret,omitempty"`
}

type Path struct {
	Path string `yaml:"path,omitempty"`
}

type NetworkEntity struct {
	TLSCACerts *Path      `yaml:"tlsCACerts,omitempty"`
	URL        string     `yaml:"url,omitempty"`
	Registrar  *Registrar `yaml:"registrar,omitempty"`
}

type ChannelPeer struct {
	ChaincodeQuery bool `yaml:"chaincodeQuery,omitempty"`
	EndorsingPeer  bool `yaml:"endorsingPeer,omitempty"`
	EventSource    bool `yaml:"eventSource,omitempty"`
	LedgerQuery    bool `yaml:"ledgerQuery,omitempty"`
}

type Channel struct {
	Orderers []string                `yaml:"orderers,omitempty"`
	Peers    map[string]*ChannelPeer `yaml:"peers,omitempty"`
}

type Provider struct {
	Provider string `yaml:"provider,omitempty"`
}

type BCCSPSecurity struct {
	Default       *Provider `yaml:"default,omitempty"`
	Enabled       bool      `yaml:"enabled,omitempty"`
	HashAlgorithm string    `yaml:"hashAlgorithm,omitempty"`
	Level         int       `yaml:"level,omitempty"`
	SoftVerify    bool      `yaml:"softVerify,omitempty"`
}

type BCCSP struct {
	Security *BCCSPSecurity `yaml:"security,omitempty"`
}

type CredentialStore struct {
	CryptoStore *Path  `yaml:"cryptoStore,omitempty"`
	Path        string `yaml:"path,omitempty"`
}

type Logging struct {
	Level string `yaml:"level,omitempty"`
}

type TLSCertsClient struct {
	Cert *Path `yaml:"cert,omitempty"`
	Key  *Path `yaml:"key,omitempty"`
}

type TLSCerts struct {
	Client *TLSCertsClient `yaml:"client,omitempty"`
}

type Organization struct {
	CertificateAuthorities []string `yaml:"certificateAuthorities,omitempty"`
	CryptoPath             string   `yaml:"cryptoPath,omitempty"`
	MSPID                  string   `yaml:"mspid,omitempty"`
	Peers                  []string `yaml:"peers,omitempty"`
}

type Client struct {
	BCCSP           *BCCSP           `yaml:"BCCSP,omitempty"`
	CredentialStore *CredentialStore `yaml:"credentialStore"`
	CryptoConfig    *Path            `yaml:"cryptoconfig,omitempty"`
	Logging         *Logging         `yaml:"logging,omitempty"`
	Organization    string           `yaml:"organization,omitempty"`
	TLSCerts        *TLSCerts        `yaml:"tlsCerts,omitempty"`
}

type FabricNetworkConfig struct {
	CertificateAuthorities map[string]*NetworkEntity `yaml:"certificateAuthorities,omitempty"`
	Channels               map[string]*Channel       `yaml:"channels,omitempty"`
	Client                 *Client                   `yaml:"client,omitempty"`
	Organization           string                    `yaml:"organization,omitempty"`
	Orderers               map[string]*NetworkEntity `yaml:"orderers,omitempty"`
	Organizations          map[string]*Organization  `yaml:"organizations,omitempty"`
	Peers                  map[string]*NetworkEntity `yaml:"peers,omitempty"`
	Version                string                    `yaml:"version,omitempty"`
}

func WriteNetworkConfig(outputPath string) error {
	networkConfig := &FabricNetworkConfig{
		CertificateAuthorities: map[string]*NetworkEntity{
			"org1.example.com": {
				TLSCACerts: &Path{
					Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/ca/fabric_ca.org1.example.com-cert.pem",
				},
				URL: "http://fabric_ca:7054",
				Registrar: &Registrar{
					EnrollID:     "admin",
					EnrollSecret: "adminpw",
				},
			},
		},
		Channels: map[string]*Channel{
			"firefly": {
				Orderers: []string{"fabric_orderer"},
				Peers: map[string]*ChannelPeer{
					"fabric_peer": {
						ChaincodeQuery: true,
						EndorsingPeer:  true,
						EventSource:    true,
						LedgerQuery:    true,
					},
				},
			},
		},
		Client: &Client{
			BCCSP: &BCCSP{
				Security: &BCCSPSecurity{
					Default: &Provider{
						Provider: "SW",
					},
					Enabled:       true,
					HashAlgorithm: "SHA2",
					Level:         256,
					SoftVerify:    true,
				},
			},
			CredentialStore: &CredentialStore{
				CryptoStore: &Path{
					Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/msp",
				},
				Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/msp",
			},
			CryptoConfig: &Path{
				Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/msp",
			},
			Logging: &Logging{
				Level: "info",
			},
			Organization: "org1.example.com",
			TLSCerts: &TLSCerts{
				Client: &TLSCertsClient{
					Cert: &Path{
						Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/tls/client.crt",
					},
					Key: &Path{
						Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/tls/client.key",
					},
				},
			},
		},
		Orderers: map[string]*NetworkEntity{
			"fabric_orderer": {
				TLSCACerts: &Path{
					Path: "/etc/firefly/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
				},
				URL: "grpcs://fabric_orderer:7050",
			},
		},
		Organizations: map[string]*Organization{
			"org1.example.com": {
				CertificateAuthorities: []string{"org1.example.com"},
				CryptoPath:             "/tmp/msp",
				MSPID:                  "Org1MSP",
				Peers:                  []string{"fabric_peer"},
			},
		},
		Peers: map[string]*NetworkEntity{
			"fabric_peer": {
				TLSCACerts: &Path{
					Path: "/etc/firefly/organizations/peerOrganizations/org1.example.com/tlsca/tlsfabric_ca.org1.example.com-cert.pem",
				},
				URL: "grpcs://fabric_peer:7051",
			},
		},
		Version: "1.1.0%",
	}
	networkConfigBytes, _ := yaml.Marshal(networkConfig)
	return os.WriteFile(outputPath, networkConfigBytes, 0755)
}
