package stacks

import (
	"fmt"
	"path"
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
	Command     string                       `yaml:"command,omitempty"`
	Environment map[string]string            `yaml:"environment,omitempty"`
	Volumes     []string                     `yaml:"volumes,omitempty"`
	Ports       []string                     `yaml:"ports,omitempty"`
	DependsOn   map[string]map[string]string `yaml:"depends_on,omitempty"`
	HealthCheck *HealthCheck                 `yaml:"healthcheck,omitempty"`
	Logging     *LoggingConfig               `yaml:"logging,omitempty"`
}

type DockerCompose struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
}

func CreateDockerCompose(stack *Stack) *DockerCompose {
	stackDir := path.Join(StacksDir, stack.name)
	dataDir := path.Join(stackDir, "data")
	compose := &DockerCompose{
		Version:  "2.1",
		Services: make(map[string]*Service),
	}

	ganacheCommand := ""

	for _, member := range stack.members {
		ganacheCommand += "--account " + member.privateKey + ",100000000000000000000 "
	}
	ganacheCommand += "--db /data/ganache"

	standardLogOptions := &LoggingConfig{
		Driver: "json-file",
		Options: map[string]string{
			"max-size": "10m",
			"max-file": "1",
		},
	}

	compose.Services["ganache"] = &Service{
		Image:   "trufflesuite/ganache-cli",
		Command: ganacheCommand,
		Volumes: []string{dataDir + ":/data"},
		Logging: standardLogOptions,
	}

	for _, member := range stack.members {
		compose.Services["firefly_core_"+member.id] = &Service{
			Image:     "kaleido-io/firefly",
			Ports:     []string{fmt.Sprint(member.exposedApiPort) + ":5000"},
			Volumes:   []string{path.Join(stackDir, "firefly_"+member.id+".core") + ":/etc/firefly/firefly.core"},
			DependsOn: map[string]map[string]string{"postgres_" + member.id: {"condition": "service_healthy"}},
			Logging:   standardLogOptions,
		}

		compose.Services["postgres_"+member.id] = &Service{
			Image:       "postgres",
			Environment: map[string]string{"POSTGRES_PASSWORD": "f1refly"},
			Volumes:     []string{path.Join(dataDir, "postgres_"+member.id) + ":/var/lib/postgresql/data"},
			HealthCheck: &HealthCheck{
				Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
				Interval: "5s",
				Timeout:  "5s",
				Retries:  5,
			},
			Logging: standardLogOptions,
		}

		compose.Services["ethconnect_"+member.id] = &Service{
			Image:     "kaleido-io/ethconnect",
			Command:   "rest -U http://127.0.0.1:8080 -I / -r http://ganache_" + member.id + ":8545",
			DependsOn: map[string]map[string]string{"ganache": {"condition": "service_started"}},
			Logging:   standardLogOptions,
		}

		compose.Services["ipfs_"+member.id] = &Service{
			Image: "ipfs/go-ipfs",
			Volumes: []string{
				path.Join(dataDir, "ipfs_"+member.id, "staging") + ":/export",
				path.Join(dataDir, "ipfs_"+member.id, "data") + ":/data/ipfs",
			},
			Logging: standardLogOptions,
		}
	}

	return compose
}
