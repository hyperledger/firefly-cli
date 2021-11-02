#!/bin/bash
ip=$(hostname -i)
echo "Here now ${BESU_CONS_API:-IBFT}"
echo "IP is $ip"
while [ ! -f "/opt/besu/public-keys/bootnode_pubkey" ]; do sleep 5; done ; \
/opt/besu/bin/besu \
--config-file=/config/besu/config.toml \
--p2p-host=$ip \
--genesis-file=/config/besu/ibft2Genesis.json \
--node-private-key-file=/opt/besu/keys/key \
--min-gas-price=0 \
--rpc-http-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} \
--rpc-ws-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} ;