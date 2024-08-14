// Copyright © 2024 Kaleido, Inc.
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

package fabric

// Note: Any change of image tag should be checked if it is published in arm64 format.
// Refer to this commit for when arm64 support was added and the code workaround was removed:
// https://github.com/hyperledger/firefly-cli/pull/323/commits/71237b73b07bfee72b355dea83af9cd874b2a2de
var FabricToolsImageName = "hyperledger/fabric-tools:2.5.6"
var FabricCAImageName = "hyperledger/fabric-ca:1.5"
var FabricOrdererImageName = "hyperledger/fabric-orderer:2.5"
var FabricPeerImageName = "hyperledger/fabric-peer:2.5"
