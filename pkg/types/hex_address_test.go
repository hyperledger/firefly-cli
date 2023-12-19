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
			Name:     "TestCase1",
			Hexvalue: "0x123abc",
		},
		{
			Name:     "TestCase2",
			Hexvalue: "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			expectedHexAddress := HexAddress(tc.Hexvalue)
			result := ConvertToHexAddress(tc.Hexvalue)

			if result != expectedHexAddress {
				t.Errorf("ConvertToHexAddress(%s) = %s; want %s", tc.Hexvalue, result, expectedHexAddress)
			}
		})

	}
}
