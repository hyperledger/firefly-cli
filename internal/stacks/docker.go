package stacks

import (
	"fmt"
	"os"
	"path"
)

type DependsOn map[string]map[string]string

type HealthCheck struct {
	Test     []string `yaml:"test,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

type Service struct {
	Image       string                       `yaml:"image,omitempty"`
	Command     string                       `yaml:"command,omitempty"`
	Environment map[string]string            `yaml:"environment,omitempty"`
	Volumes     []string                     `yaml:"volumes,omitempty"`
	Ports       []string                     `yaml:"ports,omitempty"`
	DependsOn   map[string]map[string]string `yaml:"depends_on,omitempty"`
	HealthCheck *HealthCheck                 `yaml:"healthcheck,omitempty"`
}

type DockerCompose struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
}

func CreateDockerCompose(stack *Stack) *DockerCompose {
	homeDir, _ := os.UserHomeDir()
	stackDir := path.Join(homeDir, ".firefly", stack.name)
	dataDir := path.Join(stackDir, "data")
	compose := &DockerCompose{
		Version:  "2.1",
		Services: make(map[string]*Service),
	}
	compose.Services["ganache"] = &Service{
		Image:   "trufflesuite/ganache-cli",
		Command: "--account 0xce8acaf351099a6e0351116c49f309306dd5029bea5efad328779f33bffaa218,100000000000000000000 --db /data/ganache",
		Volumes: []string{dataDir + ":/data"},
	}

	for _, member := range stack.members {
		compose.Services["firefly_core_"+member.id] = &Service{
			Image:     "kaleido-io/firefly",
			Ports:     []string{fmt.Sprint(member.exposedApiPort) + ":5000"},
			Volumes:   []string{path.Join(stackDir, "firefly_"+member.id+".core") + ":/etc/firefly/firefly.core"},
			DependsOn: map[string]map[string]string{"postgres_" + member.id: {"condition": "service_healthy"}},
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
		}

		compose.Services["ethconnect"+member.id] = &Service{
			Image:   "kaleido-io/ethconnect",
			Command: "rest -U http://127.0.0.1:8080 -I / -r http://ganache_" + member.id + ":8545",
		}

		compose.Services["ipfs_1"+member.id] = &Service{
			Image: "ipfs/go-ipfs",
		}
	}

	return compose
}
