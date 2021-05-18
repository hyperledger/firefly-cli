package stacks

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type AddressesConfig struct {
	API        string   `json:"API"`
	Announce   []string `json:"Announce"`
	Gateway    string   `json:"Gateway"`
	NoAnnounce []string `json:"NoAnnounce"`
	Swarm      []string `json:"Swarm"`
}

type MountChildConfig struct {
	Path        string `json:"path,omitempty"`
	ShardFunc   string `json:"shardFunc,omitempty"`
	Sync        bool   `json:"sync,omitempty"`
	Type        string `json:"type,omitempty"`
	Compression string `json:"compression,omitempty"`
}

type MountConfig struct {
	Child      *MountChildConfig `json:"child"`
	MountPoint string            `json:"mountpoint"`
	Prefix     string            `json:"prefix"`
	Type       string            `json:"type"`
}

type SpecConfig struct {
	Mounts []*MountConfig `json:"mounts"`
	Type   string         `json:"type"`
}

type DatastoreConfig struct {
	BloomFilterSize    int         `json:"BloomFilterSize"`
	GCPeriod           string      `json:"GCPeriod"`
	HashOnRead         bool        `json:"HashOnRead"`
	Spec               *SpecConfig `json:"Spec"`
	StorageGCWatermark int         `json:"StorageGCWatermark"`
	StorageMax         string      `json:"StorageMax"`
}

type IdentityConfig struct {
	PeerID  string `json:"PeerID"`
	PrivKey string `json:"PrivKey"`
}

type IPFSConfig struct {
	Addresses *AddressesConfig `json:"Addresses"`
	Bootstrap []string         `json:"Bootstrap"`
	Datastore *DatastoreConfig `json:"Datastore"`
	Identity  *IdentityConfig  `json:"Identity"`
}

func GenerateSwarmKey() string {
	privKey, _, _ := crypto.GenerateKeyPair(crypto.ECDSA, 2048)
	privateKeyBytes, _ := privKey.Bytes()
	hexKey := hex.EncodeToString(privateKeyBytes)
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

func NewIpfsConfigs(stack *Stack) map[string]*IPFSConfig {
	configs := make(map[string]*IPFSConfig)
	bootstrapAddresses := make([]string, len(stack.members))
	for i, member := range stack.members {
		bootstrapAddresses[i] = fmt.Sprintf("/dnsaddr/ipfs_%s/p2p/%s", member.id, member.ipfsIdentity.PeerID)
	}

	for _, member := range stack.members {

		configs[member.id] = &IPFSConfig{
			Addresses: &AddressesConfig{
				API:        "/ip4/0.0.0.0/tcp/5001",
				Announce:   []string{},
				Gateway:    "/ip4/0.0.0.0/tcp/8080",
				NoAnnounce: []string{},
				Swarm:      []string{},
			},
			Bootstrap: bootstrapAddresses,
			Datastore: &DatastoreConfig{
				BloomFilterSize: 0,
				GCPeriod:        "1h",
				HashOnRead:      false,
				Spec: &SpecConfig{
					Mounts: []*MountConfig{
						{
							Child: &MountChildConfig{
								Path:      "blocks",
								ShardFunc: "/repo/flatfs/shard/v1/next-to-last/2",
								Sync:      true,
								Type:      "flatfs",
							},
							MountPoint: "/blocks",
							Prefix:     "flatfs.datastore",
							Type:       "measure",
						},
						{
							Child: &MountChildConfig{
								Compression: "none",
								Path:        "datastore",
								Type:        "levelds",
							},
							MountPoint: "/",
							Prefix:     "develdb.datastore",
							Type:       "measure",
						},
					},
					Type: "mount",
				},
				StorageGCWatermark: 90,
				StorageMax:         "10GB",
			},
			Identity: &IdentityConfig{
				PrivKey: member.ipfsIdentity.PrivKey,
				PeerID:  member.ipfsIdentity.PeerID,
			},
		}
	}
	return configs
}
