package blockchain

import "github.com/hyperledger-labs/firefly-cli/internal/docker"

type IBlockchainProvider interface {
	WriteConfig() error
	Init() error
	PreStart() error
	PostStart() error
	GetDockerServiceDefinition() (serviceName string, serviceDefinition *docker.Service)
}
