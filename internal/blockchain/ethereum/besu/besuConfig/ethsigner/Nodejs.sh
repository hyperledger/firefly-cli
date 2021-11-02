#!/bin/bash
cd /usr/local/bin 
npm install web3 
npm install fs 
cp /ethSigner/createKeyFile.js /usr/local/bin/createKeyFile.js 
cp -r /ethSigner/accounts /usr/local/bin/accounts 
node createKeyFile.js
