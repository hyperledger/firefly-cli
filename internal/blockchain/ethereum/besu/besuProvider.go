package besu

import (
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type BesuProvider struct {
	Verbose bool
	Stack   *types.Stack
}

func (p *BesuProvider) WriteConfig() error {
	return nil
}

func (p *BesuProvider) Init() error {
	return nil
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart() error {
	return nil
}

func (p *BesuProvider) GetDockerServiceDefinition() (serviceName string, serviceDefinition *docker.Service) {
	serviceDefinition = &docker.Service{}
	return "besu", serviceDefinition
}
