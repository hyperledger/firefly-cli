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

import "gopkg.in/yaml.v3"

type HexAddress string

// Explicitly converts a string representation of a hex address to HexAddress type.
// This function assumes that hexStr is a valid hex address.
func ConvertToHexAddress(hexStr string) HexAddress {
	return HexAddress(hexStr)
}

//describe the HexValue type to be in string format
type MyType struct {
	HexValue HexAddress `yaml:"hexvalue"`
}

// Explicitly quote hex addresses so that they are interpreted as string (not int)
func (mt *MyType) MarshalYAML() (interface{}, error) {
	hexAddr := ConvertToHexAddress(string(mt.HexValue))
	return yaml.Node{
		Value: string(hexAddr),
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}, nil
}
