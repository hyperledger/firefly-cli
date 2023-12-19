// Copyright © 2022 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type HexAddress string

// checks whether a string is a valid Hexaddress
func IsValidHex(s string) (HexAddress, error) {
	//remove the 0x value if present
	hexDigits := strings.TrimPrefix(s, "0x")
	// Check if the remaining characters are valid hexadecimal digits
	match, err := regexp.MatchString("^[0-9a-fA-F]+$", hexDigits)
	if err != nil {
		return "", fmt.Errorf("error in validating hex address: %v", err)
	}
	if match {
		return HexAddress("0x" + hexDigits), nil
	} else {
		return "", fmt.Errorf("invalid hex address: %s", s)
	}
}

// Explicitly converts hex address to HexAddress type.
func ConvertToHexAddress(hexStr string) (HexAddress, error) {
	validAddress, err := IsValidHex(hexStr)
	if err != nil {
		return "", err
	}
	return validAddress, nil
}

// describe the HexValue type to be in string format
type MyType struct {
	HexValue HexAddress `yaml:"hexvalue"`
}

// Explicitly quote hex addresses so that they are interpreted as string (not int)
func (mt *MyType) MarshalYAML() (interface{}, error) {
	hexAddr, err := ConvertToHexAddress(string(mt.HexValue))
	if err != nil {
		return nil, err
	}
	return yaml.Node{
		Value: string(hexAddr),
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}, nil
}
