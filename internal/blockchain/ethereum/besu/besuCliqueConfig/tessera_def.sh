#!/bin/sh
ip=$(hostname -i)
mkdir -p /var/log/tessera/;
mkdir -p /data/tm/;
cp /config/keys/tm.* /data/tm/ ;

	cat <<EOF > /data/tm/tessera-config-09.json
	{
            "mode": "orion",
            "useWhiteList": false,
            "jdbc": {
              "username": "sa",
              "password": "",
              "url": "jdbc:h2:./data/tm/db;MODE=Oracle;TRACE_LEVEL_SYSTEM_OUT=0",
              "autoCreateTables": true
            },
            "serverConfigs":[
            {
              "app":"ThirdParty",
              "enabled": true,
              "serverAddress": "http://$ip:9080",
              "communicationType" : "REST"
            },
            {
              "app":"Q2T",
              "enabled": true,
              "serverAddress": "http://$ip:9101",
              "sslConfig": {
                "tls": "OFF"
              },
              "communicationType" : "REST"
            },
            {
              "app":"P2P",
              "enabled": true,
              "serverAddress": "http://$ip:9000",
              "sslConfig": {
                "tls": "OFF"
              },
              "communicationType" : "REST"
            }
            ],
            "peer": [
                {
                    "url": "http://member1tessera:9000"
                },
                {
                    "url": "http://member2tessera:9000"
                },
                {
                    "url": "http://member3tessera:9000"
                }
            ],
            "keys": {
              "passwords": [],
              "keyData": [
                {
                  "config": $(cat /data/tm/tm.key),
                  "publicKey": "$(cat /data/tm/tm.pub)"
                }
              ]
            },
            "alwaysSendTo": []
          }
EOF
	cat /data/tm/tessera-config-09.json
	/tessera/bin/tessera -configfile /data/tm/tessera-config-09.json
  

