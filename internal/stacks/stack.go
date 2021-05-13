package stacks

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"

	"gopkg.in/yaml.v2"
)

var homeDir, _ = os.UserHomeDir()
var StacksDir = path.Join(homeDir, ".firefly", "stacks")

type Stack struct {
	name    string
	members []*member
}

type member struct {
	id             string
	index          int
	address        string
	privateKey     string
	exposedApiPort int
	exposedUiPort  int
}

func InitStack(stackName string, memberCount int) {
	stack := &Stack{
		name:    stackName,
		members: make([]*member, memberCount),
	}
	for i := 0; i < memberCount; i++ {
		stack.members[i] = createMember(fmt.Sprint(i), i)
	}

	compose := CreateDockerCompose(stack)

	ensureDirectories(stack)
	writeDockerCompose(stack.name, compose)
	writeConfigs(stack)
}

func CheckExists(stackName string) bool {
	_, err := os.Stat(path.Join(StacksDir, stackName))
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ensureDirectories(stack *Stack) {

	dataDir := path.Join(StacksDir, stack.name, "data")

	for _, member := range stack.members {
		os.MkdirAll(path.Join(dataDir, "postgres_"+member.id), 0755)
		os.MkdirAll(path.Join(dataDir, "ipfs_"+member.id, "staging"), 0755)
		os.MkdirAll(path.Join(dataDir, "ipfs_"+member.id, "data"), 0755)
	}
	os.MkdirAll(path.Join(dataDir, "ganache"), 0755)
}

func writeDockerCompose(stackName string, compose *DockerCompose) {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		log.Fatal(err)
	}

	stackDir := path.Join(StacksDir, stackName)
	ioutil.WriteFile(path.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func writeConfigs(stack *Stack) {
	configs := NewFireflyConfigs(stack)

	stackDir := path.Join(StacksDir, stack.name)

	for memberId, config := range configs {
		bytes, _ := yaml.Marshal(config)
		ioutil.WriteFile(path.Join(stackDir, "firefly_"+memberId+".core"), bytes, 0755)
	}
}

func createMember(id string, index int) *member {
	privateKey, _ := crypto.GenerateKey()

	privateKeyBytes := crypto.FromECDSA(privateKey)
	encodedPrivateKey := "0x" + hexutil.Encode(privateKeyBytes)[2:]

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes[1:])
	encodedAddress := hexutil.Encode(hash.Sum(nil)[12:])

	return &member{
		id:             id,
		index:          index,
		address:        encodedAddress,
		privateKey:     encodedPrivateKey,
		exposedApiPort: 5000 + index,
		exposedUiPort:  3000 + index,
	}
}
