package stacks

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func GenerateSwarmKey() string {
	key := make([]byte, 32)
	rand.Read(key)
	hexKey := hex.EncodeToString(key)
	return "/key/swarm/psk/1.0.0/\n/base16/\n" + hexKey
}

func GenerateKeyAndPeerId() (privateKey string, peerId string) {
	privKey, publicKey, _ := crypto.GenerateKeyPair(crypto.Ed25519, 2048)
	privateKeyBytes, _ := privKey.Bytes()
	privateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)
	peer, _ := peer.IDFromPublicKey(publicKey)
	peerId = peer.String()
	return privateKey, peerId
}
