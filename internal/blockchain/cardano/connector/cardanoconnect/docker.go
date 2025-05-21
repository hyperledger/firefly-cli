// Copyright © 2025 IOG Singapore and SundaeSwap, Inc.
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

package cardanoconnect

import (
	"fmt"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

func (c *Cardanoconnect) GetServiceDefinitions(s *types.Stack, dependentServices map[string]string) []*docker.ServiceDefinition {
	dependsOn := make(map[string]map[string]string)
	for dep, state := range dependentServices {
		dependsOn[dep] = map[string]string{"condition": state}
	}
	extraHosts := make([]string, 0)
	if strings.Contains(s.RemoteNodeURL, "host.docker.internal") {
		extraHosts = append(extraHosts, "host.docker.internal:host-gateway")
	}
	serviceDefinitions := make([]*docker.ServiceDefinition, len(s.Members))
	for i, member := range s.Members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "cardanoconnect_" + member.ID,
			Service: &docker.Service{
				Image:         s.VersionManifest.Cardanoconnect.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_cardanoconnect_%v", s.Name, i),
				Command:       "./firefly-cardanoconnect -f /cardanoconnect/config/config.yaml",
				DependsOn:     dependsOn,
				Ports:         []string{fmt.Sprintf("%d:%d", member.ExposedConnectorPort, c.Port())},
				ExtraHosts:    extraHosts,
				Volumes: []string{
					fmt.Sprintf("cardanoconnect_config_%s:/cardanoconnect/config", member.ID),
					fmt.Sprintf("cardanoconnect_contracts_%s:/cardanoconnect/contracts", member.ID),
					fmt.Sprintf("cardanoconnect_sqlite_%s:/cardanoconnect/sqlite", member.ID),
				},
				Logging: docker.StandardLogOptions,
			},
			VolumeNames: []string{
				fmt.Sprintf("cardanoconnect_config_%s", member.ID),
				fmt.Sprintf("cardanoconnect_contracts_%s", member.ID),
				fmt.Sprintf("cardanoconnect_sqlite_%s", member.ID),
			},
		}
		if s.Socket != "" {
			serviceDefinitions[i].Service.Volumes = append(serviceDefinitions[i].Service.Volumes, fmt.Sprintf("%s:/ipc/socket", s.Socket))
		}
	}
	return serviceDefinitions
}
