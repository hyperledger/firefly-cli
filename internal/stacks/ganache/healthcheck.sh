#!/bin/sh

timeout 5s curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Host: localhost" -H "Origin: http://localhost:8545" -H "Sec-WebSocket-Key: super_secret_key" -H "Sec-WebSocket-Version: 13" http://localhost:8545

if [ $? -eq 143 ]; then exit 0; else exit 1; fi