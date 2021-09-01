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
	"fmt"
	"path"

	"github.com/hyperledger-labs/firefly-cli/internal/constants"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

func GenerateDockerServiceDefinitions(s *types.Stack) []*docker.ServiceDefinition {
	stackDir := path.Join(constants.StacksDir, s.Name)
	serviceDefinitions := []*docker.ServiceDefinition{
		// Fabric CA
		{
			ServiceName: "fabric_ca",
			Service: &docker.Service{
				Image: "hyperledger/fabric-ca:1.5",
				Environment: map[string]string{
					"FABRIC_CA_HOME":                            "/etc/hyperledger/fabric-ca-server",
					"FABRIC_CA_SERVER_CA_NAME":                  "fabric_ca",
					"FABRIC_CA_SERVER_TLS_ENABLED":              "true",
					"FABRIC_CA_SERVER_PORT":                     "7054",
					"FABRIC_CA_SERVER_OPERATIONS_LISTENADDRESS": "0.0.0.0:17054",
					"FABRIC_CA_SERVER_CA_CERTFILE":              "/etc/hyperledger/fabric-ca-server-config/fabric_ca.org1.example.com-cert.pem",
					"FABRIC_CA_SERVER_CA_KEYFILE":               "/etc/hyperledger/fabric-ca-server-config/priv_sk",
				},
				// TODO: Figure out how to increment ports here
				Ports: []string{
					"7054:7054",
					"17054:17054",
				},
				Command: "sh -c 'fabric-ca-server start -b admin:adminpw -d'",
				Volumes: []string{
					fmt.Sprintf("%s:/etc/hyperledger/fabric-ca-server-config", path.Join(stackDir, "blockchain", "organizations", "peerOrganizations", "org1.example.com", "ca")),
				},
				Hostname: "fabric_ca",
			},
			VolumeNames: []string{"fabric_ca"},
		},

		// Fabric Orderer
		{
			ServiceName: "fabric_orderer",
			Service: &docker.Service{
				Image: "hyperledger/fabric-orderer:2.3",
				Environment: map[string]string{
					"FABRIC_LOGGING_SPEC":                       "INFO",
					"ORDERER_GENERAL_LISTENADDRESS":             "0.0.0.0",
					"ORDERER_GENERAL_LISTENPORT":                "7050",
					"ORDERER_GENERAL_LOCALMSPID":                "OrdererMSP",
					"ORDERER_GENERAL_LOCALMSPDIR":               "/var/hyperledger/orderer/msp",
					"ORDERER_GENERAL_TLS_ENABLED":               "true",
					"ORDERER_GENERAL_TLS_PRIVATEKEY":            "/var/hyperledger/orderer/tls/server.key",
					"ORDERER_GENERAL_TLS_CERTIFICATE":           "/var/hyperledger/orderer/tls/server.crt",
					"ORDERER_GENERAL_TLS_ROOTCAS":               "[/var/hyperledger/orderer/tls/ca.crt]",
					"ORDERER_KAFKA_TOPIC_REPLICATIONFACTOR":     "1",
					"ORDERER_KAFKA_VERBOSE":                     "true",
					"ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE": "/var/hyperledger/orderer/tls/server.crt",
					"ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY":  "/var/hyperledger/orderer/tls/server.key",
					"ORDERER_GENERAL_CLUSTER_ROOTCAS":           "[/var/hyperledger/orderer/tls/ca.crt]",
					"ORDERER_GENERAL_BOOTSTRAPMETHOD":           "none",
					"ORDERER_CHANNELPARTICIPATION_ENABLED":      "true",
					"ORDERER_ADMIN_TLS_ENABLED":                 "true",
					"ORDERER_ADMIN_TLS_CERTIFICATE":             "/var/hyperledger/orderer/tls/server.crt",
					"ORDERER_ADMIN_TLS_PRIVATEKEY":              "/var/hyperledger/orderer/tls/server.key",
					"ORDERER_ADMIN_TLS_ROOTCAS":                 "[/var/hyperledger/orderer/tls/ca.crt]",
					"ORDERER_ADMIN_TLS_CLIENTROOTCAS":           "[/var/hyperledger/orderer/tls/ca.crt]",
					"ORDERER_ADMIN_LISTENADDRESS":               "0.0.0.0:7053",
					"ORDERER_OPERATIONS_LISTENADDRESS":          "0.0.0.0:17050",
				},
				WorkingDir: "/opt/gopath/src/github.com/hyperledger/fabric",
				Command:    "orderer",
				Volumes: []string{
					// fmt.Sprintf("%s:/etc/hyperledger/fabric/genesisblock", path.Join(stackDir, "blockchain", "genesis_block.pb")),
					fmt.Sprintf("%s:/var/hyperledger/orderer/msp", path.Join(stackDir, "blockchain", "organizations", "ordererOrganizations", "example.com", "orderers", "fabric_orderer.example.com", "msp")),
					fmt.Sprintf("%s:/var/hyperledger/orderer/tls", path.Join(stackDir, "blockchain", "organizations", "ordererOrganizations", "example.com", "orderers", "fabric_orderer.example.com", "tls")),
					"fabric_orderer:/var/hyperledger/production/orderer",
				},
				// TODO: Figure out how to increment ports here
				Ports: []string{
					"7050:7050",
					"7053:7053",
					"17050:17050",
				},
			},
			VolumeNames: []string{"fabric_orderer"},
		},

		// Fabric Peer
		{
			ServiceName: "fabric_peer",
			Service: &docker.Service{
				Image: "hyperledger/fabric-peer:2.3",
				Environment: map[string]string{
					"CORE_VM_ENDPOINT":                      "unix:///host/var/run/docker.sock",
					"CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE": fmt.Sprintf("%s_default", s.Name),
					"FABRIC_LOGGING_SPEC":                   "INFO",
					"CORE_PEER_TLS_ENABLED":                 "true",
					"CORE_PEER_PROFILE_ENABLED":             "false",
					"CORE_PEER_TLS_CERT_FILE":               "/etc/hyperledger/fabric/tls/server.crt",
					"CORE_PEER_TLS_KEY_FILE":                "/etc/hyperledger/fabric/tls/server.key",
					"CORE_PEER_TLS_ROOTCERT_FILE":           "/etc/hyperledger/fabric/tls/ca.crt",
					"CORE_PEER_ID":                          "fabric_peer",
					"CORE_PEER_ADDRESS":                     "fabric_peer:7051",
					"CORE_PEER_LISTENADDRESS":               "0.0.0.0:7051",
					"CORE_PEER_CHAINCODEADDRESS":            "fabric_peer:7052",
					"CORE_PEER_CHAINCODELISTENADDRESS":      "0.0.0.0:7052",
					"CORE_PEER_GOSSIP_BOOTSTRAP":            "fabric_peer:7051",
					"CORE_PEER_GOSSIP_EXTERNALENDPOINT":     "fabric_peer:7051",
					"CORE_PEER_LOCALMSPID":                  "Org1MSP",
					"CORE_OPERATIONS_LISTENADDRESS":         "0.0.0.0:17051",
				},
				Volumes: []string{
					fmt.Sprintf("%s:/etc/hyperledger/fabric/msp", path.Join(stackDir, "blockchain", "organizations", "peerOrganizations", "org1.example.com", "peers", "fabric_peer.org1.example.com", "msp")),
					fmt.Sprintf("%s:/etc/hyperledger/fabric/tls", path.Join(stackDir, "blockchain", "organizations", "peerOrganizations", "org1.example.com", "peers", "fabric_peer.org1.example.com", "tls")),
					"fabric_peer:/var/hyperledger/production",
					"/var/run/docker.sock:/host/var/run/docker.sock",
				},
				Ports: []string{
					"7051:7051",
					"17051:17051",
				},
			},
			VolumeNames: []string{"fabric_peer"},
		},
	}
	return serviceDefinitions
}
