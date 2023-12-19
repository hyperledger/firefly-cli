package types

import (
	"testing"
)

func TestConvertToHexAddress(t *testing.T) {
	tests := []struct {
		Name     string
		Hexvalue string
	}{
		{
			Name:     "Case 24-Bits",
			Hexvalue: "0x123abc",
		},
		{
			Name:     "Case 32-Bits",
			Hexvalue: "0xABCDEFFEDCBA9876543210ABCDEFFEDC",
		},
		{
			Name:     "Case 64-Bits",
			Hexvalue: "0x1234567890ABCDEF0123456789ABCDEF",
		},
		{
			Name:     "Case 128-Bits",
			Hexvalue: "0x1234567890ABCDEF0123456789ABCDEF",
		},
		{
			Name:     "Case 152-Bits",
			Hexvalue: "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
		},
		{
			Name:     "Case 16-Bits",
			Hexvalue: "0x1F2A",
		},
		{
			Name:     "Case 65-Bits",
			Hexvalue: "0x1234567890ABCDEF0123456789ABCDEF6789ABCDEF0123456789ABCDEF012345",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := ConvertToHexAddress(tc.Hexvalue)
			if err != nil {
				t.Log("error in generating result", err)
			}
			t.Log(result)

		})

	}
}
