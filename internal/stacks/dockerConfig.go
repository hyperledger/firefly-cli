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

type Service struct {
	Image       string                       `yaml:"image,omitempty"`
	Build       string                       `yaml:"build,omitempty"`
	Command     string                       `yaml:"command,omitempty"`
	Environment map[string]string            `yaml:"environment,omitempty"`
	Volumes     []string                     `yaml:"volumes,omitempty"`
	Ports       []string                     `yaml:"ports,omitempty"`
	DependsOn   map[string]map[string]string `yaml:"depends_on,omitempty"`
	HealthCheck *HealthCheck                 `yaml:"healthcheck,omitempty"`
	Logging     *LoggingConfig               `yaml:"logging,omitempty"`
}

type DockerComposeConfig struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
	Volumes  map[string]struct{} `yaml:"volumes,omitempty"`
}

func CreateDockerCompose(stack *Stack) *DockerComposeConfig {
	compose := &DockerComposeConfig{
		Version:  "2.1",
		Services: make(map[string]*Service),
		Volumes:  make(map[string]struct{}),
	}

	standardLogOptions := &LoggingConfig{
		Driver: "json-file",
		Options: map[string]string{
			"max-size": "10m",
			"max-file": "1",
		},
	}

	addresses := ""
	for i, member := range stack.Members {
		addresses = addresses + member.Address
		if i+1 < len(stack.Members) {
			addresses = addresses + ","
		}
	}
	gethCommand := fmt.Sprintf(`--datadir /data --syncmode 'full' --port 30311 --rpcvhosts=* --rpccorsdomain "*" --miner.gastarget 804247552 --rpc --rpcaddr "0.0.0.0" --rpcport 8545 --rpcapi 'admin,personal,db,eth,net,web3,txpool,miner,clique' --networkid 2021 --miner.gasprice 0 --unlock '%s' --password /data/password --mine --nousb --allow-insecure-unlock --nodiscover`, addresses)

	compose.Services["geth"] = &Service{
		Image:   "ethereum/client-go:release-1.9",
		Command: gethCommand,
		Volumes: []string{"geth:/data"},
		Logging: standardLogOptions,
		Ports:   []string{fmt.Sprintf("%d:8545", stack.ExposedGethPort)},
	}

	compose.Volumes["geth"] = struct{}{}

	for _, member := range stack.Members {

		if !member.External {
			compose.Services["firefly_core_"+member.ID] = &Service{
				Image: "ghcr.io/hyperledger-labs/firefly:latest",
				Ports: []string{
					fmt.Sprintf("%d:%d", member.ExposedFireflyPort, member.ExposedFireflyPort),
					fmt.Sprintf("%d:%d", member.ExposedFireflyAdminPort, member.ExposedFireflyAdminPort),
				},
				Volumes: []string{fmt.Sprintf("firefly_core_%s:/etc/firefly", member.ID)},
				DependsOn: map[string]map[string]string{
					"ethconnect_" + member.ID:   {"condition": "service_started"},
					"dataexchange_" + member.ID: {"condition": "service_started"},
				},
				Logging: standardLogOptions,
			}

			compose.Volumes["firefly_core_"+member.ID] = struct{}{}
		}

		if stack.Database == "postgres" {
			compose.Services["postgres_"+member.ID] = &Service{
				Image: "postgres",
				Ports: []string{fmt.Sprintf("%d:5432", member.ExposedPostgresPort)},
				Environment: map[string]string{
					"POSTGRES_PASSWORD": "f1refly",
					"PGDATA":            "/var/lib/postgresql/data/pgdata",
				},
				Volumes: []string{fmt.Sprintf("postgres_%s:/var/lib/postgresql/data", member.ID)},
				HealthCheck: &HealthCheck{
					Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
					Interval: "5s",
					Timeout:  "3s",
					Retries:  12,
				},
				Logging: standardLogOptions,
			}

			compose.Volumes["postgres_"+member.ID] = struct{}{}

			compose.Services["firefly_core_"+member.ID].DependsOn["postgres_"+member.ID] = map[string]string{"condition": "service_healthy"}
		}

		compose.Services["ethconnect_"+member.ID] = &Service{
			Image:     "ghcr.io/hyperledger-labs/firefly-ethconnect:latest",
			Command:   "rest -U http://127.0.0.1:8080 -I ./abis -r http://geth:8545 -E ./events -d 3",
			DependsOn: map[string]map[string]string{"geth": {"condition": "service_started"}},
			Ports:     []string{fmt.Sprintf("%d:8080", member.ExposedEthconnectPort)},
			Volumes: []string{
				fmt.Sprintf("ethconnect_abis_%s:/ethconnect/abis", member.ID),
				fmt.Sprintf("ethconnect_events_%s:/ethconnect/events", member.ID),
			},
			Logging: standardLogOptions,
		}

		compose.Volumes["ethconnect_abis_"+member.ID] = struct{}{}
		compose.Volumes["ethconnect_events_"+member.ID] = struct{}{}

		compose.Services["ipfs_"+member.ID] = &Service{
			Image: "ipfs/go-ipfs",
			Ports: []string{
				fmt.Sprintf("%d:5001", member.ExposedIPFSApiPort),
				fmt.Sprintf("%d:8080", member.ExposedIPFSGWPort),
			},
			Environment: map[string]string{
				"IPFS_SWARM_KEY":    stack.SwarmKey,
				"LIBP2P_FORCE_PNET": "1",
			},
			Volumes: []string{
				fmt.Sprintf("ipfs_staging_%s:/export", member.ID),
				fmt.Sprintf("ipfs_data_%s:/data/ipfs", member.ID),
			},
			Logging: standardLogOptions,
		}

		compose.Volumes["ipfs_staging_"+member.ID] = struct{}{}
		compose.Volumes["ipfs_data_"+member.ID] = struct{}{}

		compose.Services["dataexchange_"+member.ID] = &Service{
			Image:   "ghcr.io/hyperledger-labs/firefly-dataexchange-https:latest",
			Ports:   []string{fmt.Sprintf("%d:3000", member.ExposedDataexchangePort)},
			Volumes: []string{fmt.Sprintf("dataexchange_%s:/data", member.ID)},
			Logging: standardLogOptions,
		}

		compose.Volumes["dataexchange_"+member.ID] = struct{}{}

	}

	return compose
}
