package besu

import (
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger-labs/firefly-cli/internal/core"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/internal/log"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type BesuProvider struct {
	Verbose bool
	Log     log.Logger
	Stack   *types.Stack
}

func (p *BesuProvider) WriteConfig() error {
	return nil
}

func (p *BesuProvider) RunFirstTimeSetup() error {
	return nil
}

func (p *BesuProvider) DeploySmartContracts() error {
	return ethereum.DeployContracts(p.Stack, p.Log, p.Verbose)
}

func (p *BesuProvider) PreStart() error {
	return nil
}

func (p *BesuProvider) PostStart() error {
	return nil
}

func (p *BesuProvider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 1)
	serviceDefinitions[0] = &docker.ServiceDefinition{
		ServiceName: "besu",
		Service:     &docker.Service{},
	}
	return serviceDefinitions
}

func (p *BesuProvider) GetFireflyConfig(m *types.Member) *core.BlockchainConfig {
	return &core.BlockchainConfig{}
}
