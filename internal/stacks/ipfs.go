package stacks

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func WriteSwarmKey(path string) {
	privKey, _, _ := crypto.GenerateKeyPair(crypto.ECDSA, 2048)
	privateKeyBytes, _ := privKey.Bytes()
	hexKey := hex.EncodeToString(privateKeyBytes)
	bytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + hexKey)
	ioutil.WriteFile(path, bytes, 0755)
}
