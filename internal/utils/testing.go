// This file contains different Setup and tools, for the FireFly-CLI testing Environment
package utils

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

type TestHelper struct {
	FabricURL     string
	EthConnectURL string
	EvmConnectURL string
	BaseURL       string
}

var (
	FabricEndpoint     = "http://localhost:7054"
	EthConnectEndpoint = "http://localhost:8080"
	EvmConnectEndpoint = "http://localhost:5008"
)

func StartMockServer(t *testing.T) {
	httpmock.Activate()
}

// mockprotocol endpoints for testing
func NewTestEndPoint(t *testing.T) *TestHelper {
	return &TestHelper{
		FabricURL:     FabricEndpoint,
		EthConnectURL: EthConnectEndpoint,
		EvmConnectURL: EvmConnectEndpoint,
	}
}

func StopMockServer(_ *testing.T) {
	httpmock.DeactivateAndReset()
}
