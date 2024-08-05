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

package quorum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum/tessera"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
)

var DockerEntrypoint = "docker-entrypoint.sh"
var QuorumPort = "8545"

func CreateQuorumEntrypoint(ctx context.Context, outputDirectory, consensus, stackName string, memberIndex, chainID, blockPeriodInSeconds int, privateTransactionManager fftypes.FFEnum) error {
	var discoveryCmd string
	connectTimeout := 15
	if memberIndex != 0 {
		discoveryCmd = fmt.Sprintf(`bootnode=$(curl http://quorum_0:%s -s --connect-timeout %[2]d --max-time %[2]d --retry 5 --retry-connrefused --retry-delay 0 --retry-max-time 60 --fail --header "Content-Type: application/json" --data '{"jsonrpc":"2.0", "method": "admin_nodeInfo", "params": [], "id": 1}' | grep -o "enode://[a-z0-9@.:]*")
BOOTNODE_CMD="--bootnodes $bootnode"
BOOTNODE_CMD=${BOOTNODE_CMD/127.0.0.1/quorum_0}`, QuorumPort, connectTimeout)
	} else {
		discoveryCmd = `BOOTNODE_CMD=""`
	}

	tesseraCmd := ""
	if !privateTransactionManager.Equals(types.PrivateTransactionManagerNone) {
		tesseraCmd = fmt.Sprintf(`TESSERA_URL=http://%[5]s_member%[1]dtessera
TESSERA_TP_PORT=%[2]s
TESSERA_Q2T_PORT=%[3]s
TESSERA_UPCHECK_URL=$TESSERA_URL:$TESSERA_TP_PORT/upcheck
ADDITIONAL_ARGS="${ADDITIONAL_ARGS:-} --ptm.timeout 5 --ptm.url ${TESSERA_URL}:${TESSERA_Q2T_PORT} --ptm.http.writebuffersize 4096 --ptm.http.readbuffersize 4096 --ptm.tls.mode off"

echo -n "Checking tessera is up ... "
curl --connect-timeout %[4]d --max-time %[4]d --retry 5 --retry-connrefused --retry-delay 0 --retry-max-time 60 --silent --fail "${TESSERA_UPCHECK_URL}"
echo ""
`, memberIndex, tessera.TmTpPort, tessera.TmQ2tPort, connectTimeout, stackName)
	}

	blockPeriod := blockPeriodInSeconds
	if blockPeriodInSeconds == -1 {
		blockPeriod = 5
	}
	blockPeriodInMs := blockPeriod * 60

	content := fmt.Sprintf(`#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

GOQUORUM_CONS_ALGO=%[1]s
if [ "ibft" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using istanbul for consensus algorithm..."
    export CONSENSUS_ARGS="--istanbul.blockperiod %[6]d --mine --miner.threads 1 --miner.gasprice 0 --emitcheckpoints"
    export QUORUM_API="istanbul"
elif [ "qbft" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using qbft for consensus algorithm..."
    export CONSENSUS_ARGS="--mine --miner.threads 1 --miner.gasprice 0 --emitcheckpoints"
    export QUORUM_API="istanbul"
elif [ "raft" == "$GOQUORUM_CONS_ALGO" ];
then
    echo "Using raft for consensus algorithm..."
    export CONSENSUS_ARGS="--raft --raftblocktime %[7]d --raftport 53000"
    export QUORUM_API="raft"
elif [ "clique" == "$GOQUORUM_CONS_ALGO" ];
then
	echo "Using clique for consensus algorithm..."
	export CONSENSUS_ARGS=""
	export QUORUM_API="clique"
fi

ADDITIONAL_ARGS=${ADDITIONAL_ARGS:-}
%[2]s

# discovery
%[3]s
echo "bootnode discovery command :: $BOOTNODE_CMD"
IP_ADDR=$(cat /etc/hosts | tail -n 1 | awk '{print $1}')

exec geth --datadir /data --nat extip:$IP_ADDR --syncmode 'full' --revertreason --port 30311 --http --http.addr "0.0.0.0" --http.corsdomain="*" -http.port %[4]s --http.vhosts "*" --http.api admin,personal,eth,net,web3,txpool,miner,debug,$QUORUM_API --networkid %[5]d --miner.gasprice 0 --password /data/password --mine --allow-insecure-unlock --verbosity 4 $CONSENSUS_ARGS $BOOTNODE_CMD $ADDITIONAL_ARGS`, consensus, tesseraCmd, discoveryCmd, QuorumPort, chainID, blockPeriod, blockPeriodInMs)
	filename := filepath.Join(outputDirectory, DockerEntrypoint)
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
	return nil
}

func CopyQuorumEntrypointToVolume(ctx context.Context, quorumEntrypointDirectory, volumeName string) error {
	if err := docker.CopyFileToVolume(ctx, volumeName, filepath.Join(quorumEntrypointDirectory, DockerEntrypoint), ""); err != nil {
		return err
	}
	return nil
}
