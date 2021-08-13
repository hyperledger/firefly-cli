package blockchain

import (
	"github.com/hyperledger-labs/firefly-cli/internal/core"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type IBlockchainProvider interface {
	WriteConfig() error
	RunFirstTimeSetup() error
	DeploySmartContracts() error
	PreStart() error
	PostStart() error
	GetDockerServiceDefinitions() []*docker.ServiceDefinition
	GetFireflyConfig(m *types.Member) *core.BlockchainConfig
}
