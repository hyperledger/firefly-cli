package stacks

type LogConfig struct {
	Level string `yaml:"level,omitempty"`
}

type HttpConfig struct {
	Port    int    `yaml:"port,omitempty"`
	Address string `yaml:"address,omitempty"`
}

type NodeConfig struct {
	Identity string `yaml:"identity,omitempty"`
}

type BasicAuth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type EthconnectConfig struct {
	URL                 string     `yaml:"url,omitempty"`
	Instance            string     `yaml:"instance,omitempty"`
	Topic               string     `yaml:"topic,omitempty"`
	SkipEventStreamInit bool       `yaml:"skipEventstreamInit,omitempty"`
	Auth                *BasicAuth `yaml:"auth,omitempty"`
}

type BlockchainConfig struct {
	Type       string            `yaml:"type,omitempty"`
	Ethconnect *EthconnectConfig `yaml:"ethconnect,omitempty"`
}

type DatabaseConfig struct {
	Type        string `yaml:"type,omitempty"`
	URL         string `yaml:"url,omitempty"`
	AutoMigrate bool   `yaml:"autoMigrate,omitempty"`
}

type P2PFSConfig struct {
	Type string `yaml:"type,omitempty"`
	URL  string `yaml:"url,omitempty"`
}

type FireflyConfig struct {
	Log        *LogConfig        `yaml:"log,omitempty"`
	Debug      *HttpConfig       `yaml:"debug,omitempty"`
	HTTP       *HttpConfig       `yaml:"http,omitempty"`
	Node       *NodeConfig       `yaml:"node,omitempty"`
	Blockchain *BlockchainConfig `yaml:"blockchain,omitempty"`
	Database   *DatabaseConfig   `yaml:"database,omitempty"`
	P2PFS      *P2PFSConfig      `yaml:"p2pfs,omitempty"`
}

func NewFireflyConfigs(stack *Stack) map[string]*FireflyConfig {
	configs := make(map[string]*FireflyConfig)

	for _, member := range stack.members {
		configs[member.id] = &FireflyConfig{
			Log: &LogConfig{
				Level: "debug",
			},
			Debug: &HttpConfig{
				Port: 6060,
			},
			HTTP: &HttpConfig{
				Port:    5000,
				Address: "0.0.0.0",
			},
			Node: &NodeConfig{
				Identity: member.address,
			},
			Blockchain: &BlockchainConfig{
				Type: "ethereum",
				Ethconnect: &EthconnectConfig{
					URL:                 "http://ethconnect_" + member.id + ":8080",
					Instance:            "/instances/1c197604587f046fd40684a8f21f4609fb811a7b",
					Topic:               member.id,
					SkipEventStreamInit: true,
				},
			},
			Database: &DatabaseConfig{
				Type:        "postgres",
				URL:         "postgres://postgres:f1refly@postgres_" + member.id + ":5432?sslmode=disable",
				AutoMigrate: true,
			},
			P2PFS: &P2PFSConfig{
				Type: "ipfs",
				URL:  "http://ipfs_" + member.id,
			},
		}
	}
	return configs
}
