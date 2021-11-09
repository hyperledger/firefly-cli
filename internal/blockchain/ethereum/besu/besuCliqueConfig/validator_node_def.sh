#!/bin/bash
while [ ! -f "/opt/besu/public-keys/bootnode_pubkey" ]; do sleep 5; done ; \
/opt/besu/bin/besu \
--Xdns-enabled=true \
--Xdns-update-enabled=true \
--config-file=/config/besu/config.toml \
--genesis-file=/config/besu/CliqueGenesis.json \
--node-private-key-file=/opt/besu/keys/key \
--min-gas-price=0 \
--rpc-http-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} \
--rpc-ws-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} ;