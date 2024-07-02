package quorum

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestUnlockAccount(t *testing.T) {
	tests := []struct {
		Name        string
		RPCUrl      string
		Address     string
		Password    string
		StatusCode  int
		ApiResponse *JSONRPCResponse
	}{
		{
			Name:       "TestUnlockAccount-1",
			RPCUrl:     "http://127.0.0.1:8545",
			Address:    "user-1",
			Password:   "POST",
			StatusCode: 200,
			ApiResponse: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      0,
				Error:   nil,
				Result:  "mock result",
			},
		},
		{
			Name:       "TestUnlockAccountError-2",
			RPCUrl:     "http://127.0.0.1:8545",
			Address:    "user-1",
			Password:   "POST",
			StatusCode: 200,
			ApiResponse: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      0,
				Error:   &JSONRPCError{500, "invalid account"},
				Result:  "mock result",
			},
		},
		{
			Name:       "TestUnlockAccountHTTPError-3",
			RPCUrl:     "http://localhost:8545",
			Address:    "user-1",
			Password:   "POST",
			StatusCode: 500,
			ApiResponse: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      0,
				Error:   nil,
				Result:  "mock result",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			apiResponse, _ := json.Marshal(tc.ApiResponse)
			// mockResponse
			httpmock.RegisterResponder("POST", tc.RPCUrl,
				httpmock.NewStringResponder(tc.StatusCode, string(apiResponse)))
			client := NewQuorumClient(tc.RPCUrl)
			utils.StartMockServer(t)
			err := client.UnlockAccount(tc.Address, tc.Password)
			utils.StopMockServer(t)

			// expect errors when returned status code != 200 or ApiResponse comes back with non nil error
			if tc.StatusCode != 200 || tc.ApiResponse.Error != nil {
				assert.NotNil(t, err, "expects error to be returned when either quorum returns an application error or non 200 http response")
			} else {
				assert.NoError(t, err, fmt.Sprintf("unable to unlock account: %v", err))
			}
		})
	}
}
