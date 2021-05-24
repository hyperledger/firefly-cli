package stacks

import (
	"fmt"
	"os/exec"
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
	stackDir := path.Join(StacksDir, stack.Name)
	dataDir := path.Join(stackDir, "data")
	compose := &DockerCompose{
		Version:  "2.1",
		Services: make(map[string]*Service),
	}

	ganacheCommand := ""

	for _, member := range stack.Members {
		ganacheCommand += "--account " + member.PrivateKey + ",100000000000000000000 "
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

	for _, member := range stack.Members {
		compose.Services["firefly_core_"+member.ID] = &Service{
			Image:     "kaleidoinc/firefly",
			Ports:     []string{fmt.Sprint(member.ExposedFireflyPort) + ":5000"},
			Volumes:   []string{path.Join(stackDir, "firefly_"+member.ID+".core") + ":/etc/firefly/firefly.core"},
			DependsOn: map[string]map[string]string{"postgres_" + member.ID: {"condition": "service_healthy"}},
			Logging:   standardLogOptions,
		}

		compose.Services["postgres_"+member.ID] = &Service{
			Image:       "postgres",
			Environment: map[string]string{"POSTGRES_PASSWORD": "f1refly"},
			Volumes:     []string{path.Join(dataDir, "postgres_"+member.ID) + ":/var/lib/postgresql/data"},
			HealthCheck: &HealthCheck{
				Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
				Interval: "5s",
				Timeout:  "5s",
				Retries:  5,
			},
			Logging: standardLogOptions,
		}

		compose.Services["ethconnect_"+member.ID] = &Service{
			Image:     "kaleidoinc/ethconnect",
			Command:   "rest -U http://127.0.0.1:8080 -I / -r http://ganache:8545 -d 3",
			DependsOn: map[string]map[string]string{"ganache": {"condition": "service_started"}},
			Ports:     []string{fmt.Sprint(member.ExposedEthconnectPort) + ":8080"},
			Logging:   standardLogOptions,
		}

		compose.Services["ipfs_"+member.ID] = &Service{
			Image: "ipfs/go-ipfs",
			Volumes: []string{
				path.Join(dataDir, "ipfs_"+member.ID, "staging") + ":/export",
				path.Join(dataDir, "ipfs_"+member.ID, "data") + ":/data/ipfs",
			},
			Environment: map[string]string{
				"IPFS_SWARM_KEY":    stack.SwarmKey,
				"LIBP2P_FORCE_PNET": "1",
			},
			Logging: standardLogOptions,
		}
	}

	return compose
}

func RunDockerComposeCommand(stackName string, command ...string) error {
	stackDir := path.Join(StacksDir, stackName)
	dockerCmd := exec.Command("docker", append([]string{"compose"}, command...)...)
	dockerCmd.Dir = stackDir
	return dockerCmd.Run()
}

func RunDockerCommand(stackName string, command ...string) error {
	stackDir := path.Join(StacksDir, stackName)
	dockerCmd := exec.Command("docker", command...)
	dockerCmd.Dir = stackDir
	return dockerCmd.Run()
}
