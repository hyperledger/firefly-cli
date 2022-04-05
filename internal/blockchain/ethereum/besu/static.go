// Copyright © 2021 Kaleido, Inc.
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

package besu

import (
	_ "embed"
	"io/ioutil"
	"os"
	"path/filepath"
)

//go:embed besuCliqueConfig/besu/networkFiles/member1/keys/key.pub
var mem1key_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/member1/keys/key
var mem1key_priv []byte

//go:embed besuCliqueConfig/besu/networkFiles/validator1/keys/key.pub
var val1key_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/validator1/keys/key
var val1key_priv []byte

//go:embed besuCliqueConfig/besu/networkFiles/rpcnode/keys/key.pub
var rpckey_pub []byte

//go:embed besuCliqueConfig/besu/networkFiles/rpcnode/keys/key
var rpckey_priv []byte

//go:embed besuCliqueConfig/besu/.env
var besu_env []byte

//go:embed besuCliqueConfig/besu/config.toml
var besu_config []byte

//go:embed besuCliqueConfig/besu/permissions_config.toml
var besu_perm_config []byte

//go:embed besuCliqueConfig/besu/static-nodes.json
var static_nodes []byte

//go:embed besuCliqueConfig/besu/log-config-splunk.xml
var log_config_splunk []byte

//go:embed besuCliqueConfig/besu/log-config.xml
var log_config []byte

//go:embed besuCliqueConfig/ethsigner/createKeyFile.js
var createKeyFile []byte

//go:embed besuCliqueConfig/ethsigner/ethsigner.sh
var ethsigner_sh []byte

//go:embed besuCliqueConfig/ethsigner/Nodejs.sh
var nodejs_sh []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member1/tm.key
var tessera_mem1_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member1/tm.pub
var tessera_mem1_tmpub []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member2/tm.key
var tessera_mem2_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member2/tm.pub
var tessera_mem2_tmpub []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member3/tm.key
var tessera_mem3_tmkey []byte

//go:embed besuCliqueConfig/tessera/networkFiles/member3/tm.pub
var tessera_mem3_tmpub []byte

//go:embed besuCliqueConfig/besu_mem1_def.sh
var mem1_entrypt_sh []byte

//go:embed besuCliqueConfig/bootnode_def.sh
var bootnode_def []byte

//go:embed besuCliqueConfig/validator_node_def.sh
var validator_def []byte

//go:embed besuCliqueConfig/tessera_def.sh
var tessera_def []byte

func (p *BesuProvider) writeStaticFiles() error {
	GetPath := func(elem ...string) string {
		return filepath.Join(append([]string{p.Stack.InitDir, "config"}, elem...)...)
	}
	if err := os.Mkdir(filepath.Join(p.Stack.InitDir, "logs"), 0755); err != nil {
		return err
	}
	log_members := []string{"besu", "tessera"}
	for _, members := range log_members {
		if err := os.Mkdir(filepath.Join(p.Stack.InitDir, "logs", members), 0755); err != nil {
			return err
		}
	}
	if err := os.Mkdir(GetPath("besu"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("besu", "networkFiles"), 0755); err != nil {
		return err
	}
	member_directories := []string{"member1", "rpcnode", "validator1"}
	for _, file := range member_directories {
		if err := os.Mkdir(GetPath("besu", "networkFiles", file), 0755); err != nil {
			return err
		}
		if err := os.Mkdir(GetPath("besu", "networkFiles", file, "keys"), 0755); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "member1", "keys", "key.pub"), mem1key_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "member1", "keys", "key"), mem1key_priv, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "validator1", "keys", "key.pub"), val1key_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "validator1", "keys", "key"), val1key_priv, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "rpcnode", "keys", "key.pub"), rpckey_pub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "networkFiles", "rpcnode", "keys", "key"), rpckey_priv, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", ".env"), besu_env, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "config.toml"), besu_config, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "permissions_config.toml"), besu_perm_config, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "static-nodes.json"), static_nodes, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "log-config-splunk.xml"), log_config_splunk, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu", "log-config.xml"), log_config, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("EthConnect"), 0755); err != nil {
		return err
	}

	if err := os.Mkdir(GetPath("ethsigner"), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "createKeyFile.js"), createKeyFile, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "ethsigner.sh"), ethsigner_sh, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("ethsigner", "Nodejs.sh"), nodejs_sh, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("tessera"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(GetPath("tessera", "networkFiles"), 0755); err != nil {
		return err
	}
	tessera_members := []string{"member1", "member2", "member3"}
	for _, member := range tessera_members {
		if err := os.Mkdir(GetPath("tessera", "networkFiles", member), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member1", "tm.key"), tessera_mem1_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member1", "tm.pub"), tessera_mem1_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member2", "tm.key"), tessera_mem2_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member2", "tm.pub"), tessera_mem2_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member3", "tm.key"), tessera_mem3_tmkey, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera", "networkFiles", "member3", "tm.pub"), tessera_mem3_tmpub, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("besu_mem1_def.sh"), mem1_entrypt_sh, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(GetPath("bootnode_def.sh"), bootnode_def, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("validator_node_def.sh"), validator_def, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(GetPath("tessera_def.sh"), tessera_def, 0755); err != nil {
		return err
	}
	return nil
}
