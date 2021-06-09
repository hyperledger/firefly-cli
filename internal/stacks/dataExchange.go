package stacks

import (
	"fmt"
)

type DataExchangeListenerConfig struct {
	Hostname string `json:"hostname,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Port     int    `json:"port,omitempty"`
}

type PeerConfig struct {
	ID       string `json:"id,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type DataExchangePeerConfig struct {
	API   *DataExchangeListenerConfig `json:"api,omitempty"`
	P2P   *DataExchangeListenerConfig `json:"p2p,omitempty"`
	Peers []*PeerConfig               `json:"peers"`
}

func (s *Stack) GenerateDataExchangeHTTPSConfig(memberId string) *DataExchangePeerConfig {
	return &DataExchangePeerConfig{
		API: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3000,
		},
		P2P: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3001,
			Endpoint: fmt.Sprintf("https://dataexchange_%s:3001", memberId),
		},
		Peers: []*PeerConfig{},
	}
}
