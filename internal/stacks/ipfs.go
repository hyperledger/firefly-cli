package stacks

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/crypto"
)

func WriteSwarmKey(path string) {
	key, _ := crypto.GenerateKey()
	privateKeyBytes := crypto.FromECDSA(key)
	hexKey := hex.EncodeToString(privateKeyBytes)
	bytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + hexKey)
	ioutil.WriteFile(path, bytes, 0755)
}
