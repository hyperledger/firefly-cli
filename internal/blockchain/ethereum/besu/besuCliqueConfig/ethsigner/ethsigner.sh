#!/bin/bash
apt-get update
yes | apt-get install curl

bin/ethsigner --chain-id=1337 --http-listen-host=0.0.0.0 --downstream-http-port=8555 --downstream-http-host=rpcnode multikey-signer --directory=/SignerConfig
