// Copyright © 2025 Kaleido, Inc.
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

package docker

import (
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type DependsOn map[string]map[string]string

type HealthCheck struct {
	Test     []string `yaml:"test,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

type LoggingConfig struct {
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

type ServiceDefinition struct {
	ServiceName string
	Service     *Service
	VolumeNames []string
}

type Service struct {
	ContainerName string                       `yaml:"container_name,omitempty"`
	Image         string                       `yaml:"image,omitempty"`
	Build         string                       `yaml:"build,omitempty"`
	User          string                       `yaml:"user,omitempty"`
	Command       string                       `yaml:"command,omitempty"`
	Environment   map[string]interface{}       `yaml:"environment,omitempty"`
	Volumes       []string                     `yaml:"volumes,omitempty"`
	Ports         []string                     `yaml:"ports,omitempty"`
	DependsOn     map[string]map[string]string `yaml:"depends_on,omitempty"`
	HealthCheck   *HealthCheck                 `yaml:"healthcheck,omitempty"`
	Logging       *LoggingConfig               `yaml:"logging,omitempty"`
	WorkingDir    string                       `yaml:"working_dir,omitempty"`
	EntryPoint    []string                     `yaml:"entrypoint,omitempty"`
	EnvFile       string                       `yaml:"env_file,omitempty"`
	Expose        []int                        `yaml:"expose,omitempty"`
	Deploy        map[string]interface{}       `yaml:"deploy,omitempty"`
	Platform      string                       `yaml:"platform,omitempty"`
	ExtraHosts    []string                     `yaml:"extra_hosts,omitempty"`
}

type DockerComposeConfig struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
	Volumes  map[string]struct{} `yaml:"volumes,omitempty"`
}

var StandardLogOptions = &LoggingConfig{
	Driver: "json-file",
	Options: map[string]string{
		"max-size": "10m",
		"max-file": "1",
	},
}

func CreateDockerCompose(s *types.Stack) *DockerComposeConfig {
	compose := &DockerComposeConfig{
		Version:  "2.4",
		Services: make(map[string]*Service),
		Volumes:  make(map[string]struct{}),
	}
	for _, member := range s.Members {

		// Look at the VersionManifest to see if a specific version of FireFly was provided, else use latest, assuming a locally built image
		const fireflyCore = "firefly_core"
		if !member.External {
			configFile := filepath.Join(s.RuntimeDir, "config", fmt.Sprintf("%s_%s.yml", fireflyCore, member.ID))
			compose.Services[fireflyCore+"_"+member.ID] = &Service{
				Image:         s.VersionManifest.FireFly.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_%s_%s", s.Name, fireflyCore, member.ID),
				Ports: []string{
					fmt.Sprintf("%d:%d", member.ExposedFireflyPort, member.ExposedFireflyPort),
					fmt.Sprintf("%d:%d", member.ExposedFireflyAdminSPIPort, member.ExposedFireflyAdminSPIPort),
				},
				Volumes: []string{
					fmt.Sprintf("%s:/etc/firefly/firefly.core.yml:ro", configFile),
					fmt.Sprintf("%s_data_%s:/etc/firefly/data", fireflyCore, member.ID),
				},
				DependsOn:   map[string]map[string]string{},
				Logging:     StandardLogOptions,
				Environment: s.EnvironmentVars,
				HealthCheck: &HealthCheck{
					Test: []string{
						"CMD",
						"curl",
						"--fail",
						fmt.Sprintf("http://localhost:%d/api/v1/status", member.ExposedFireflyPort),
					},
					Interval: "15s", // 6000 requests in a day
					Retries:  30,
				},
			}
			compose.Volumes[fmt.Sprintf("%s_data_%s", fireflyCore, member.ID)] = struct{}{}
			compose.Services[fireflyCore+"_"+member.ID].DependsOn["dataexchange_"+member.ID] = map[string]string{"condition": "service_started"}
			compose.Services[fireflyCore+"_"+member.ID].DependsOn["ipfs_"+member.ID] = map[string]string{"condition": "service_healthy"}
		}
		if s.Database == "postgres" {
			compose.Services["postgres_"+member.ID] = &Service{
				Image:         constants.PostgresImageName,
				ContainerName: fmt.Sprintf("%s_postgres_%s", s.Name, member.ID),
				Ports:         []string{fmt.Sprintf("%d:5432", member.ExposedDatabasePort)},
				Environment: s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
					"POSTGRES_PASSWORD": "f1refly",
					"PGDATA":            "/var/lib/postgresql/data/pgdata"}),
				Volumes: []string{fmt.Sprintf("postgres_%s:/var/lib/postgresql/data", member.ID)},
				HealthCheck: &HealthCheck{
					Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
					Interval: "5s",
					Timeout:  "3s",
					Retries:  12,
				},
				Logging: StandardLogOptions,
			}
			compose.Volumes[fmt.Sprintf("postgres_%s", member.ID)] = struct{}{}
			if service, ok := compose.Services[fmt.Sprintf("%s_%s", fireflyCore, member.ID)]; ok {
				service.DependsOn["postgres_"+member.ID] = map[string]string{"condition": "service_healthy"}
			}
		}
		sharedStorage := &Service{
			Image:         constants.IPFSImageName,
			ContainerName: fmt.Sprintf("%s_ipfs_%s", s.Name, member.ID),
			Ports: []string{
				fmt.Sprintf("%d:5001", member.ExposedIPFSApiPort),
				fmt.Sprintf("%d:8080", member.ExposedIPFSGWPort),
			},
			Volumes: []string{
				fmt.Sprintf("ipfs_staging_%s:/export", member.ID),
				fmt.Sprintf("ipfs_data_%s:/data/ipfs", member.ID),
			},
			Logging: StandardLogOptions,
			HealthCheck: &HealthCheck{
				Test:     []string{"CMD-SHELL", `wget --post-data= http://127.0.0.1:5001/api/v0/id -O - -q`},
				Interval: "5s",
				Timeout:  "3s",
				Retries:  12,
			},
		}
		if s.IPFSMode.Equals(types.IPFSModePrivate) {
			sharedStorage.Environment = s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
				"IPFS_SWARM_KEY":    s.SwarmKey,
				"LIBP2P_FORCE_PNET": "1",
			},
			)
		} else {
			sharedStorage.Environment = s.EnvironmentVars
		}
		compose.Services["ipfs_"+member.ID] = sharedStorage
		compose.Volumes[fmt.Sprintf("ipfs_staging_%s", member.ID)] = struct{}{}
		compose.Volumes[fmt.Sprintf("ipfs_data_%s", member.ID)] = struct{}{}
		compose.Services["dataexchange_"+member.ID] = &Service{
			Image:         s.VersionManifest.DataExchange.GetDockerImageString(),
			ContainerName: fmt.Sprintf("%s_dataexchange_%s", s.Name, member.ID),
			Ports:         []string{fmt.Sprintf("%d:3000", member.ExposedDataexchangePort)},
			Volumes:       []string{fmt.Sprintf("dataexchange_%s:/data", member.ID)},
			Logging:       StandardLogOptions,
			Environment:   s.EnvironmentVars,
		}
		compose.Volumes[fmt.Sprintf("dataexchange_%s", member.ID)] = struct{}{}
		if s.SandboxEnabled {
			compose.Services["sandbox_"+member.ID] = &Service{
				Image:         constants.SandboxImageName,
				ContainerName: fmt.Sprintf("%s_sandbox_%s", s.Name, member.ID),
				Ports:         []string{fmt.Sprintf("%d:3001", member.ExposedSandboxPort)},
				Environment: s.ConcatenateWithProvidedEnvironmentVars(map[string]interface{}{
					"FF_ENDPOINT": fmt.Sprintf("http://firefly_core_%d:%d", *member.Index, member.ExposedFireflyPort),
				}),
			}
		}
	}

	if s.PrometheusEnabled {
		compose.Services["prometheus"] = &Service{
			Image:         constants.PrometheusImageName,
			ContainerName: fmt.Sprintf("%s_prometheus", s.Name),
			Ports:         []string{fmt.Sprintf("%d:9090", s.ExposedPrometheusPort)},
			Volumes:       []string{"prometheus_data:/prometheus", "prometheus_config:/etc/prometheus"},
			Logging:       StandardLogOptions,
			Environment:   s.EnvironmentVars,
		}
		compose.Volumes["prometheus_data"] = struct{}{}
		compose.Volumes["prometheus_config"] = struct{}{}
	}

	return compose
}
