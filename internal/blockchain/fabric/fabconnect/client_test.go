package fabconnect

import (
	"testing"

	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateIdentity(t *testing.T) {
	utils.StartMockServer(t)

	testContext := utils.NewTestEndPoint(t)
	tests := []struct {
		Name             string
		FabconnectURL    string
		Signer           string
		ApiResponse      string
		Method           string
		ExpectedResponse *CreateIdentityResponse
	}{
		{
			Name:          "TestIdentity-1",
			FabconnectURL: testContext.FabricURL + "/fabconnect/identities",
			Signer:        "user-1",
			Method:        "POST",
			ApiResponse: `
			{
			   "Name": "fabric_user-1",
			   "Secret": "9876543210987654321098765432109876543210987654321098765432109876"
			}   
			`,
			ExpectedResponse: &CreateIdentityResponse{
				Name:   "fabric_user-1",
				Secret: "9876543210987654321098765432109876543210987654321098765432109876",
			},
		},
		{
			Name:          "TestIdentity-2",
			FabconnectURL: testContext.FabricURL + "/fabconnect/identities",
			Signer:        "user-2",
			Method:        "POST",
			ApiResponse: `
			{
				"Name": "fabric_user-2",
				"Secret": "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff"
			}
			`,
			ExpectedResponse: &CreateIdentityResponse{
				Name:   "fabric_user-2",
				Secret: "aabbccddeeff0011223344556677889900112233445566778899aabbccddeeff",
			},
		},
		{
			Name:          "TestIdentity-3",
			FabconnectURL: testContext.FabricURL + "/fabconnect/identities",
			Signer:        "user-3",
			Method:        "POST",
			ApiResponse: `
			{
			   "Name": "fabric_user-3",
			   "Secret": "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
			}   
			`,
			ExpectedResponse: &CreateIdentityResponse{
				Name:   "fabric_user-3",
				Secret: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			//mockResponse
			httpmock.RegisterResponder(tc.Method, tc.FabconnectURL,
				httpmock.NewStringResponder(200, tc.ApiResponse))

			identityResp, err := CreateIdentity(tc.FabconnectURL, tc.Signer)
			if err != nil {
				t.Fatalf("unable to create identity: %v", err)
			}
			assert.NotNil(t, identityResp)
			assert.Equal(t, tc.ExpectedResponse, identityResp)
		})
	}
	utils.StartMockServer(t)
}
