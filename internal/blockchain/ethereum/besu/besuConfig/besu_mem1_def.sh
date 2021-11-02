#!/bin/bash
ip=$(hostname -i)

 
for i in $(seq 1 100) 
do
	set -e
	response="$(curl -0 -m 10 172.16.239.26:9000/upcheck)"
	if [[ "I'm up!" == $response ]];
	then break
	else
	echo "Waiting for Tessera..."
	sleep 10
	fi
done



while [ ! -f "/opt/besu/public-keys/bootnode_pubkey" ]; do sleep 5; done ;
/opt/besu/bin/besu \
--config-file=/config/besu/config.toml \
--p2p-host=$ip \
--genesis-file=/config/besu/ibft2Genesis.json \
--node-private-key-file=/opt/besu/keys/key \
--min-gas-price=0 \
--privacy-enabled \
--privacy-url=http://member1tessera:9101 \
--privacy-public-key-file=/config/tessera/tm.pub \
--privacy-onchain-groups-enabled=true \
--rpc-http-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} \
--rpc-ws-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} ;
