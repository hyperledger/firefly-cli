package types

import (
	"encoding/hex"
	"testing"

	"github.com/hyperledger/firefly-signer/pkg/ethtypes"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestWrapHexAddress(t *testing.T) {
	tests := []struct {
		Name     string
		Hexvalue string `yaml:"hexvalue"`
	}{
		{
			Name:     "Case 16-Bits",
			Hexvalue: "0x1f2a000000000000000000000000000000000000",
		},
		{
			Name:     "Case 24-Bits",
			Hexvalue: "0x123abc0000000000000000000000000000000000",
		},
		{
			Name:     "Case 32-Bits",
			Hexvalue: "0xabcdeffedcba9876543210abcdeffedc00000000",
		},
		{
			Name:     "Case 64-Bits",
			Hexvalue: "0x1234567890abcdef012345670000000000000000",
		},
		{
			Name:     "Case 65-Bits",
			Hexvalue: "0x1234567890abcdef0123456789abcdef6789abcd",
		},
		{
			Name:     "Case 128-Bits",
			Hexvalue: "0x1234567890abcdef0123456789abcdef00000000",
		},
		{
			Name:     "Case 152-Bits",
			Hexvalue: "0x549b5f43a40e1a0522864a004cfff2b0ca473a65",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			var hexType HexAddress
			hexBytes, err := hex.DecodeString(tc.Hexvalue[2:])
			if err != nil {
				t.Log("Unable to decode values:", err)
			}
			if len(hexBytes) != 20 {
				t.Fatalf("expected 20 bytes, got %d bytes", len(hexBytes))
			}
			// Copy bytes to a fixed-size array
			var hexArray [20]byte
			copy(hexArray[:], hexBytes)
			result, err := hexType.WrapHexAddress([20]byte(hexArray))
			if err != nil {
				t.Log("error in generating result", err)
				t.Fail()
				return
			}
			assert.Equal(t, tc.Hexvalue, result)
		})

	}
}

// Wrong values to test the accuracy of the  WrapHexAddress()
func TestFailWrapHexAddress(t *testing.T) {
	testData := []struct {
		Name     string
		HexValue string `yaml:"hexvalue"`
	}{
		{
			Name:     "case-1",
			HexValue: "px123abc0000000000000000000000000000000000",
		},
		{
			Name:     "case-2",
			HexValue: "A0234567890abcdef0123456789abcdef00000000",
		},
		{
			Name:     "case-3",
			HexValue: "06942dc1fC868aF18132C0916dA3ae4ab58142a4",
		},
	}
	for _, tc := range testData {
		t.Run(tc.Name, func(t *testing.T) {
			var hexType HexAddress
			hexBytes, err := hex.DecodeString(tc.HexValue[2:])
			if err != nil {
				t.Log("unable to decode hexvalue:", err)
			}
			if len(hexBytes) != 20 {
				t.Fatalf("expected 20 bytes, but got %d bytes", len(hexBytes))
			}
			var hexArray [20]byte
			copy(hexArray[:], hexBytes)
			result, err := hexType.WrapHexAddress([20]byte(hexArray))
			if err != nil {
				t.Log("error in generating result", err)
				t.Fail()
				return
			}
			assert.Equal(t, tc.HexValue, result)
		})
	}
}

type HexAddr struct {
	ethtypes.Address0xHex
}

func TestYamlMarshall(t *testing.T) {
	testAddress := []struct {
		Name     string
		Hexvalue string `yaml:"hexvalue"`
	}{
		{
			Name:     "Case 16-Bits",
			Hexvalue: "0x1f2a000000000000000000000000000000000000",
		},
		{
			Name:     "Case 24-Bits",
			Hexvalue: "0x123abc0000000000000000000000000000000000",
		},
		{
			Name:     "Case 32-Bits",
			Hexvalue: "0xabcdeffedcba9876543210abcdeffedc00000000",
		},
		{
			Name:     "Case 64-Bits",
			Hexvalue: "0x1234567890abcdef012345670000000000000000",
		},
	}
	for _, tc := range testAddress {
		t.Run(tc.Name, func(t *testing.T) {
			var hexType HexAddress
			hexbyte, err := hex.DecodeString(tc.Hexvalue[2:])
			if err != nil {
				t.Log("unable to decode values")
			}
			var hexArray [20]byte
			copy(hexArray[:], hexbyte)
			YamlHex, err := hexType.WrapHexAddress([20]byte(hexArray))
			if err != nil {
				t.Log("unable to generate yaml string")
			}

			YamlNode := yaml.Node{
				Value: YamlHex,
			}
			assert.YAMLEq(t, YamlNode.Value, YamlHex)
		})

	}
}
