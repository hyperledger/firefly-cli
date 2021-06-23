#!/bin/sh

GANACHE_PORT=8545
TIMEOUT_CODE=143

output=$(timeout 2s curl --stderr - -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Host: localhost" -H "Origin: http://localhost:$GANACHE_PORT" -H "Sec-WebSocket-Key: super_secret_key" -H "Sec-WebSocket-Version: 13" http://localhost:$GANACHE_PORT)
ret=$?

if [ $ret -eq $TIMEOUT_CODE ]; then
    echo $output | grep "101 Switching Protocols" >> /dev/null
else
    exit 1
fi