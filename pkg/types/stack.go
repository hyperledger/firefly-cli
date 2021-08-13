package types

type Stack struct {
	Name                  string    `json:"name,omitempty"`
	Members               []*Member `json:"members,omitempty"`
	SwarmKey              string    `json:"swarmKey,omitempty"`
	ExposedBlockchainPort int       `json:"exposedGethPort,omitempty"`
	Database              string    `json:"database"`
	BlockchainProvider    string    `json:"blockchainProvider"`
}

type Member struct {
	ID                      string `json:"id,omitempty"`
	Index                   *int   `json:"index,omitempty"`
	Address                 string `json:"address,omitempty"`
	PrivateKey              string `json:"privateKey,omitempty"`
	ExposedFireflyPort      int    `json:"exposedFireflyPort,omitempty"`
	ExposedFireflyAdminPort int    `json:"exposedFireflyAdminPort,omitempty"`
	ExposedEthconnectPort   int    `json:"exposedEthconnectPort,omitempty"`
	ExposedPostgresPort     int    `json:"exposedPostgresPort,omitempty"`
	ExposedDataexchangePort int    `json:"exposedDataexchangePort,omitempty"`
	ExposedIPFSApiPort      int    `json:"exposedIPFSApiPort,omitempty"`
	ExposedIPFSGWPort       int    `json:"exposedIPFSGWPort,omitempty"`
	ExposedUIPort           int    `json:"exposedUiPort,omitempty"`
	ExposedTokensPort       int    `json:"exposedTokensPort,omitempty"`
	External                bool   `json:"external,omitempty"`
}
