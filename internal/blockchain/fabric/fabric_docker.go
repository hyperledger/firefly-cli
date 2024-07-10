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
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func GenerateDockerServiceDefinitions(s *types.Stack) []*docker.ServiceDefinition {
	serviceDefinitions := []*docker.ServiceDefinition{
		// Fabric CA
		{
			ServiceName: "fabric_ca",
			Service: &docker.Service{
				Image:         FabricCAImageName,
				ContainerName: fmt.Sprintf("%s_fabric_ca", s.Name),
				Environment: s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
					"FABRIC_CA_HOME":                            "/etc/hyperledger/fabric-ca-server",
					"FABRIC_CA_SERVER_CA_NAME":                  "fabric_ca",
					"FABRIC_CA_SERVER_PORT":                     "7054",
					"FABRIC_CA_SERVER_OPERATIONS_LISTENADDRESS": "0.0.0.0:17054",
					"FABRIC_CA_SERVER_CA_CERTFILE":              "/etc/firefly/organizations/peerOrganizations/org1.example.com/ca/fabric_ca.org1.example.com-cert.pem",
					"FABRIC_CA_SERVER_CA_KEYFILE":               "/etc/firefly/organizations/peerOrganizations/org1.example.com/ca/priv_sk",
				}),
				Ports: []string{
					"7054:7054",
					"17054:17054",
				},
				Command: "sh -c 'fabric-ca-server start -b admin:adminpw'",
				Volumes: []string{
					"firefly_fabric:/etc/firefly",
				},
			},
			VolumeNames: []string{"fabric_ca"},
		},

		// Fabric Orderer
		{
			ServiceName: "fabric_orderer",
			Service: &docker.Service{
				Image:         FabricOrdererImageName,
				ContainerName: fmt.Sprintf("%s_fabric_orderer", s.Name),
				Environment: s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
					"FABRIC_LOGGING_SPEC":                       "INFO",
					"ORDERER_GENERAL_LISTENADDRESS":             "0.0.0.0",
					"ORDERER_GENERAL_LISTENPORT":                "7050",
					"ORDERER_GENERAL_LOCALMSPID":                "OrdererMSP",
					"ORDERER_GENERAL_LOCALMSPDIR":               "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/msp",
					"ORDERER_GENERAL_TLS_ENABLED":               "true",
					"ORDERER_GENERAL_TLS_PRIVATEKEY":            "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.key",
					"ORDERER_GENERAL_TLS_CERTIFICATE":           "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.crt",
					"ORDERER_GENERAL_TLS_ROOTCAS":               "[/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/ca.crt]",
					"ORDERER_KAFKA_TOPIC_REPLICATIONFACTOR":     "1",
					"ORDERER_KAFKA_VERBOSE":                     "true",
					"ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE": "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.crt",
					"ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY":  "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.key",
					"ORDERER_GENERAL_CLUSTER_ROOTCAS":           "[/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/ca.crt]",
					"ORDERER_GENERAL_BOOTSTRAPMETHOD":           "none",
					"ORDERER_CHANNELPARTICIPATION_ENABLED":      "true",
					"ORDERER_ADMIN_TLS_ENABLED":                 "true",
					"ORDERER_ADMIN_TLS_CERTIFICATE":             "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.crt",
					"ORDERER_ADMIN_TLS_PRIVATEKEY":              "/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/server.key",
					"ORDERER_ADMIN_TLS_ROOTCAS":                 "[/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/ca.crt]",
					"ORDERER_ADMIN_TLS_CLIENTROOTCAS":           "[/etc/firefly/organizations/ordererOrganizations/example.com/orderers/fabric_orderer.example.com/tls/ca.crt]",
					"ORDERER_ADMIN_LISTENADDRESS":               "0.0.0.0:7053",
					"ORDERER_OPERATIONS_LISTENADDRESS":          "0.0.0.0:17050",
				}),
				WorkingDir: "/opt/gopath/src/github.com/hyperledger/fabric",
				Command:    "orderer",
				Volumes: []string{
					"firefly_fabric:/etc/firefly",
					"fabric_orderer:/var/hyperledger/production/orderer",
				},
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
				Image:         FabricPeerImageName,
				ContainerName: fmt.Sprintf("%s_fabric_peer", s.Name),
				Environment: s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
					"CORE_VM_ENDPOINT":                      "unix:///host/var/run/docker.sock",
					"CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE": fmt.Sprintf("%s_default", s.Name),
					"FABRIC_LOGGING_SPEC":                   "INFO",
					"CORE_PEER_TLS_ENABLED":                 "true",
					"CORE_PEER_PROFILE_ENABLED":             "false",
					"CORE_PEER_MSPCONFIGPATH":               "/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/msp",
					"CORE_PEER_TLS_CERT_FILE":               "/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/server.crt",
					"CORE_PEER_TLS_KEY_FILE":                "/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/server.key",
					"CORE_PEER_TLS_ROOTCERT_FILE":           "/etc/firefly/organizations/peerOrganizations/org1.example.com/peers/fabric_peer.org1.example.com/tls/ca.crt",
					"CORE_PEER_ID":                          "fabric_peer",
					"CORE_PEER_ADDRESS":                     "fabric_peer:7051",
					"CORE_PEER_LISTENADDRESS":               "0.0.0.0:7051",
					"CORE_PEER_CHAINCODEADDRESS":            "fabric_peer:7052",
					"CORE_PEER_CHAINCODELISTENADDRESS":      "0.0.0.0:7052",
					"CORE_PEER_GOSSIP_BOOTSTRAP":            "fabric_peer:7051",
					"CORE_PEER_GOSSIP_EXTERNALENDPOINT":     "fabric_peer:7051",
					"CORE_PEER_LOCALMSPID":                  "Org1MSP",
					"CORE_OPERATIONS_LISTENADDRESS":         "0.0.0.0:17051",
				}),
				Volumes: []string{
					"firefly_fabric:/etc/firefly",
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
