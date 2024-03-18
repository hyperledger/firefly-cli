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

package tezossigner

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server ServerConfig             `yaml:"server"`
	Vaults VaultsConfig             `yaml:"vaults"`
	Tezos  map[string]AccountConfig `yaml:"tezos"`
}

type ServerConfig struct {
	Address        string `yaml:"address,omitempty"`
	UtilityAddress string `yaml:"utility_address,omitempty"`
}

type VaultsConfig struct {
	LocalSecret LocalSecretConfig `yaml:"local_secret,omitempty"`
}

type LocalSecretConfig struct {
	Driver string     `yaml:"driver,omitempty"`
	File   FileConfig `yaml:"config,omitempty"`
}

type FileConfig struct {
	SecretPath string `yaml:"file,omitempty"`
}

type AccountConfig struct {
	LogPayloads bool                      `yaml:"log_payloads"`
	Allow       AllowedTransactionsConfig `yaml:"allow"`
}

type AllowedTransactionsConfig struct {
	Block          []string `yaml:"block"`
	Endorsement    []string `yaml:"endorsement"`
	Preendorsement []string `yaml:"preendorsement"`
	Generic        []string `yaml:"generic"`
}

func (c *Config) WriteConfig(filename string) error {
	configYamlBytes, _ := yaml.Marshal(c)
	return os.WriteFile(filename, configYamlBytes, 0755)
}

func GenerateSignerConfig(accountsAddresses []string) *Config {
	config := &Config{
		Server: ServerConfig{
			Address:        ":6732",
			UtilityAddress: ":9583",
		},
		Vaults: VaultsConfig{
			LocalSecret: LocalSecretConfig{
				Driver: "file",
				File: FileConfig{
					SecretPath: "/etc/secret.json",
				},
			},
		},
	}

	addresses := map[string]AccountConfig{}
	// Give accounts the rights to sign certain types of transactions
	for _, address := range accountsAddresses {
		addresses[address] = AccountConfig{
			LogPayloads: true,
			Allow: AllowedTransactionsConfig{
				/* List of [activate_account, ballot, delegation, double_baking_evidence, double_endorsement_evidence,
				double_preendorsement_evidence, endorsement, failing_noop, origination, preendorsement, proposals,
				register_global_constant, reveal, sc_rollup_add_messages, sc_rollup_cement, sc_rollup_originate,
				sc_rollup_publish, seed_nonce_revelation, set_deposits_limit, transaction, transfer_ticket,
				tx_rollup_commit, tx_rollup_dispatch_tickets, tx_rollup_finalize_commitment, tx_rollup_origination,
				tx_rollup_rejection, tx_rollup_remove_commitment, tx_rollup_return_bond, tx_rollup_submit_batch]*/
				Generic: []string{
					"transaction",
					"endorsement",
					"reveal",
					"origination",
				},
			},
		}
	}
	config.Tezos = addresses

	return config
}
