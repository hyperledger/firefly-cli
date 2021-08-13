package ethconnect

import (
	"fmt"

	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

func GetEthconnectServiceDefinitions(members []*types.Member) []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, len(members))
	for i, member := range members {
		serviceDefinitions[i] = &docker.ServiceDefinition{
			ServiceName: "ethconnect_" + member.ID,
			Service: &docker.Service{
				Image:     "ghcr.io/hyperledger-labs/firefly-ethconnect:latest",
				Command:   "rest -U http://127.0.0.1:8080 -I ./abis -r http://geth:8545 -E ./events -d 3",
				DependsOn: map[string]map[string]string{"geth": {"condition": "service_started"}},
				Ports:     []string{fmt.Sprintf("%d:8080", member.ExposedEthconnectPort)},
				Volumes: []string{
					fmt.Sprintf("ethconnect_abis_%s:/ethconnect/abis", member.ID),
					fmt.Sprintf("ethconnect_events_%s:/ethconnect/events", member.ID),
				},
				Logging: docker.StandardLogOptions,
			},
			VolumeNames: []string{"ethconnect_abis_" + member.ID, "ethconnect_events_" + member.ID},
		}
	}
	return serviceDefinitions
}
