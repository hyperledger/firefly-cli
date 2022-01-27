#!/bin/bash
/opt/besu/bin/besu public-key export --to=/tmp/bootnode_pubkey; \
/opt/besu/bin/besu \
--Xdns-enabled=true \
--Xdns-update-enabled=true \
--config-file=/config/besu/config.toml \
--genesis-file=/config/besu/CliqueGenesis.json \
--node-private-key-file=/opt/besu/keys/key \
--min-gas-price=0 \
--rpc-http-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} \
--rpc-ws-api=EEA,WEB3,ETH,NET,PRIV,PERM,${BESU_CONS_API:-IBFT} ;