#!/bin/bash

ethconnect rest -U http://127.0.0.1:8080 -I ./abis -r http://ethsigner:8545 -E ./events -d 3
