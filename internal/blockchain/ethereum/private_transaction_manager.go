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

package ethereum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hyperledger/firefly-cli/internal/docker"
)

var entrypoint = "docker-entrypoint.sh"
var tmpath = "/qdata/tm"

func CreateTesseraKeys(ctx context.Context, image, outputDirectory, prefix, name, password string) (string, error) {
	// generates both .pub and .key files used by Tessera
	var filename string
	if prefix != "" {
		filename = fmt.Sprintf("%v_%s", prefix, name)
	} else {
		filename = name
	}
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return "", err
	}
	fmt.Println("generating tessera keys")
	err := docker.RunDockerCommand(ctx, outputDirectory, "run", "--rm", "-v", fmt.Sprintf("%s:/keystore", outputDirectory), image, "-keygen", "-filename", fmt.Sprintf("/keystore/%s", filename))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", outputDirectory, filename), nil
}

func CopyTesseraKeysToVolume(ctx context.Context, tesseraKeyPath, volumeName string) error {
	if err := docker.MkdirInVolume(ctx, volumeName, tmpath); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(ctx, volumeName, filepath.Join(tesseraKeyPath, "tm.pub"), tmpath); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(ctx, volumeName, filepath.Join(tesseraKeyPath, "tm.key"), tmpath); err != nil {
		return err
	}
	return nil
}

func CreateTesseraEntrypoint(ctx context.Context, outputDirectory, volumeName, memberCount string) error {
	// only tessera v09 onwards is supported
	var sb strings.Builder
	memberCountInt, _ := strconv.Atoi(memberCount)
	for i := 0; i < memberCountInt; i++ {
		sb.WriteString(fmt.Sprintf("{\"url\":\"http://member%dtessera:9000\"},", i)) // construct peer list
	}
	peerList := strings.TrimSuffix(sb.String(), ",")
	content := fmt.Sprintf(`export JAVA_OPTS="-Xms128M -Xmx128M"
DDIR=/data/qdata/dd
mkdir -p ${DDIR}
[ -f ${DDIR}/dd/tm.ipc ] && rm -f ${DDIR}/dd/tm.ipc
cat <<EOF > ${DDIR}/tessera-config-09.json
	{
		"useWhiteList": false,
		"jdbc": {
			"username": "sa",
			"password": "",
			"url": "jdbc:h2:./${DDIR}/db;TRACE_LEVEL_SYSTEM_OUT=0",
			"autoCreateTables": true
		},
		"serverConfigs":[
			{
				"app":"ThirdParty",
				"enabled": true,
				"serverAddress": "http://$(hostname -i):9080",
				"communicationType" : "REST"
			},
			{
				"app":"Q2T",
				"enabled": true,
				"serverAddress": "unix:${DDIR}/tm.ipc",
				"sslConfig": {
					"tls": "OFF"
				},
				"communicationType" : "REST"
			},
			{
				"app":"P2P",
				"enabled": true,
				"serverAddress": "http://$(hostname -i):9000",
				"sslConfig": {
					"tls": "OFF"
			},
				"communicationType" : "REST"
			}
		],
		"peer": [
			%s
		],
		"keys": {
		"passwords": [],
		"keyData": [
				{
					"privateKeyPath": "${DDIR}/keystore/tm.key",
					"publicKeyPath": "${DDIR}/keystore/tm.pub"
				}
			]
		},
		"alwaysSendTo": [],
		"bootstrapNode": false,
		"features": {
			"enableRemoteKeyValidation": false,
			"enablePrivacyEnhancements": true
		}
	}
EOF
/tessera/bin/tessera -configfile ${DDIR}/tessera-config-09.json
`, peerList)
	filename := filepath.Join(outputDirectory, entrypoint)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	CopyTesseraEntrypointToVolume(ctx, outputDirectory, volumeName)
	return nil
}

func CopyTesseraEntrypointToVolume(ctx context.Context, tesseraEntrypointDirectory, volumeName string) error {
	if err := docker.MkdirInVolume(ctx, volumeName, ""); err != nil {
		return err
	}
	if err := docker.CopyFileToVolume(ctx, volumeName, filepath.Join(tesseraEntrypointDirectory, entrypoint), ""); err != nil {
		return err
	}
	return nil
}

func CreateQuorumEntrypoint(ctx context.Context, outputDirectory, volumeName, memberIndex string, chainId int) error {
	discoveryCmd := "BOOTNODE_CMD=\"\""
	if memberIndex != "0" {
		discoveryCmd = `bootnode=$(curl http://geth_0:8545 -s --connect-timeout 5 --max-time 10 --retry 5 --retry-connrefused --retry-delay 0 --retry-max-time 60 --fail --header "Content-Type: application/json" --data '{"jsonrpc":"2.0", "method": "admin_nodeInfo", "params": [], "id": 1}' | grep -o "enode://[a-z0-9@.:]*")
BOOTNODE_CMD="--bootnodes $bootnode"
BOOTNODE_CMD=${BOOTNODE_CMD/127.0.0.1/geth_0}`
	}

	content := fmt.Sprintf(`#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

GOQUORUM_CONS_ALGO=$(echo "${GOQUORUM_CONS_ALGO:-clique}" | tr '[:lower:]')
if [ "istanbul" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using istanbul for consensus algorithm..."
    export CONSENSUS_ARGS="--istanbul.blockperiod 5 --mine --miner.threads 1 --miner.gasprice 0 --emitcheckpoints"
    export QUORUM_API="istanbul"
elif [ "qbft" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using qbft for consensus algorithm..."
    export CONSENSUS_ARGS="--mine --miner.threads 1 --miner.gasprice 0 --emitcheckpoints"
    export QUORUM_API="istanbul"
elif [ "raft" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using raft for consensus algorithm..."
    export CONSENSUS_ARGS="--raft --raftblocktime 300 --raftport 53000"
    export QUORUM_API="raft"
elif [ "clique" == "$GOQUORUM_CONS_ALGO" ];
then
	echo "Using clique for consensus algorithm..."
	export CONSENSUS_ARGS=""
	export QUORUM_API="clique"
fi

TESSERA_URL=http://member%stessera:9000/upcheck
ADDITIONAL_ARGS="${ADDITIONAL_ARGS:-} --ptm.timeout 5 --ptm.url ${TESSERA_URL} --ptm.http.writebuffersize 4096 --ptm.http.readbuffersize 4096 --ptm.tls.mode off"

echo -n "Checking tessera is up ... "
curl --connect-timeout 5 --max-time 10 --retry 5 --retry-connrefused --retry-delay 0 --retry-max-time 60 --silent --fail "$TESSERA_URL"
echo ""

# discovery
%s
echo "bootnode discovery command :: $BOOTNODE_CMD"
IP_ADDR=$(cat /etc/hosts | tail -n 1 | awk '{print $1}')

exec geth --datadir /data --nat extip:$IP_ADDR --syncmode 'full' --revertreason --port 30311 --http --http.addr "0.0.0.0" --http.corsdomain="*" -http.port 8545 --http.vhosts "*" --http.api admin,personal,eth,net,web3,txpool,miner,debug,$QUORUM_API --networkid %d --miner.gasprice 0 --password /data/password --mine --allow-insecure-unlock --verbosity 4 $CONSENSUS_ARGS --miner.gaslimit 16777215 $BOOTNODE_CMD`, memberIndex, discoveryCmd, chainId)
	filename := filepath.Join(outputDirectory, entrypoint)
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	CopyQuorumEntrypointToVolume(ctx, outputDirectory, volumeName)
	return nil
}

func CopyQuorumEntrypointToVolume(ctx context.Context, quorumEntrypointDirectory, volumeName string) error {
	if err := docker.CopyFileToVolume(ctx, volumeName, filepath.Join(quorumEntrypointDirectory, entrypoint), ""); err != nil {
		return err
	}
	return nil
}
