package stacks

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/briandowns/spinner"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/kaleido-io/firefly-cli/internal/contracts"
	"github.com/kaleido-io/firefly-cli/internal/docker"
	"golang.org/x/crypto/sha3"

	"gopkg.in/yaml.v2"
)

var homeDir, _ = os.UserHomeDir()
var StacksDir = path.Join(homeDir, ".firefly", "stacks")

type Stack struct {
	Name     string    `json:"name,omitempty"`
	Members  []*Member `json:"members,omitempty"`
	SwarmKey string    `json:"swarmKey,omitempty"`
}

type Member struct {
	ID                    string `json:"id,omitempty"`
	Index                 *int   `json:"index,omitempty"`
	Address               string `json:"address,omitempty"`
	PrivateKey            string `json:"privateKey,omitempty"`
	ExposedFireflyPort    int    `json:"exposedFireflyPort,omitempty"`
	ExposedEthconnectPort int    `json:"exposedEthconnectPort,omitempty"`
	ExosedUIPort          int    `json:"exposedUiPort ,omitempty"`
}

func InitStack(stackName string, memberCount int) {
	stack := &Stack{
		Name:     stackName,
		Members:  make([]*Member, memberCount),
		SwarmKey: GenerateSwarmKey(),
	}
	for i := 0; i < memberCount; i++ {
		stack.Members[i] = createMember(fmt.Sprint(i), i)
	}
	compose := CreateDockerCompose(stack)
	ensureDirectories(stack)
	writeDockerCompose(stack.Name, compose)
	writeConfigs(stack)
}

func CheckExists(stackName string) (bool, error) {
	_, err := os.Stat(path.Join(StacksDir, stackName))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func LoadStack(stackName string) (*Stack, error) {
	exists, err := CheckExists(stackName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("stack '%s' does not exist", stackName)
	}
	fmt.Printf("reading stack config...")
	if d, err := ioutil.ReadFile(path.Join(StacksDir, stackName, "stack.json")); err != nil {
		return nil, err
	} else {
		var stack *Stack
		if err := json.Unmarshal(d, &stack); err == nil {
			fmt.Printf("done\n")
		}
		return stack, err
	}

}

func ensureDirectories(stack *Stack) {

	dataDir := path.Join(StacksDir, stack.Name, "data")

	for _, member := range stack.Members {
		os.MkdirAll(path.Join(dataDir, "postgres_"+member.ID), 0755)
		os.MkdirAll(path.Join(dataDir, "ipfs_"+member.ID, "staging"), 0755)
		os.MkdirAll(path.Join(dataDir, "ipfs_"+member.ID, "data"), 0755)
	}
	os.MkdirAll(path.Join(dataDir, "ganache"), 0755)
}

func writeDockerCompose(stackName string, compose *DockerComposeConfig) {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		log.Fatal(err)
	}

	stackDir := path.Join(StacksDir, stackName)
	ioutil.WriteFile(path.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func writeConfigs(stack *Stack) {
	stackDir := path.Join(StacksDir, stack.Name)

	fireflyConfigs := NewFireflyConfigs(stack)
	for memberId, config := range fireflyConfigs {
		bytes, _ := yaml.Marshal(config)
		ioutil.WriteFile(path.Join(stackDir, "firefly_"+memberId+".core"), bytes, 0755)
	}
	bytes := []byte(`{"mounts":[{"mountpoint":"/blocks","path":"blocks","shardFunc":"/repo/flatfs/shard/v1/next-to-last/2","type":"flatfs"},{"mountpoint":"/","path":"datastore","type":"levelds"}],"type":"mount"}`)
	ioutil.WriteFile(path.Join(stackDir, "datastore_spec"), bytes, 0755)

	stackConfigBytes, _ := json.MarshalIndent(stack, "", " ")
	ioutil.WriteFile(path.Join(stackDir, "stack.json"), stackConfigBytes, 0755)
}

func createMember(id string, index int) *Member {
	privateKey, _ := secp256k1.NewPrivateKey(secp256k1.S256())
	privateKeyBytes := privateKey.Serialize()
	encodedPrivateKey := "0x" + hex.EncodeToString(privateKeyBytes)
	// Remove the "04" Suffix byte when computing the address. This byte indicates that it is an uncompressed public key.
	publicKeyBytes := privateKey.PubKey().SerializeUncompressed()[1:]
	// Take the hash of the public key to generate the address
	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes)
	// Ethereum addresses only use the lower 20 bytes, so toss the rest away
	encodedAddress := "0x" + hex.EncodeToString(hash.Sum(nil)[12:32])

	return &Member{
		ID:                    id,
		Index:                 &index,
		Address:               encodedAddress,
		PrivateKey:            encodedPrivateKey,
		ExposedFireflyPort:    5000 + index,
		ExposedEthconnectPort: 8080 + index,
		ExosedUIPort:          3000 + index,
	}
}

func (s *Stack) StartStack() error {
	fmt.Printf("starting FireFly stack '%s'... ", s.Name)
	workingDir := path.Join(StacksDir, s.Name)
	spinner := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spinner.FinalMSG = "done"
	if hasBeenRun, err := s.stackHasRunBefore(); !hasBeenRun && err == nil {
		fmt.Println("\nthis will take a few seconds longer since this is the first time you're running this stack...")
		spinner.Start()
		if err := s.runFirstTimeSetup(spinner); err != nil {
			spinner.Stop()
			return err
		}
		spinner.Stop()
		return nil
	} else if err == nil {
		spinner.Start()
		spinner.Suffix = " starting FireFly dependencies..."
		err := docker.RunDockerComposeCommand(workingDir, "up", "-d")
		spinner.Stop()
		return err
	} else {
		spinner.Stop()
		return err
	}
}

func (s *Stack) runFirstTimeSetup(spinner *spinner.Spinner) error {
	workingDir := path.Join(StacksDir, s.Name)
	spinner.Suffix = " starting FireFly dependencies..."
	if err := docker.RunDockerComposeCommand(workingDir, "up", "-d"); err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s_firefly_core_%s_1", s.Name, s.Members[0].ID)
	spinner.Suffix = " extracting smart contracts..."
	if err := s.extractContracts(containerName); err != nil {
		return err
	}
	spinner.Suffix = " deploying smart contracts..."
	if err := s.deployContracts(spinner); err != nil {
		return err
	}
	return nil
}

func (s *Stack) deployContracts(spinner *spinner.Spinner) error {
	contractDeployed := false
	paymentContract, err := contracts.ReadCompiledContract(path.Join(StacksDir, s.Name, "contracts", "Payment.json"))
	if err != nil {
		return err
	}
	fireflyContract, err := contracts.ReadCompiledContract(path.Join(StacksDir, s.Name, "contracts", "Firefly.json"))
	if err != nil {
		return err
	}
	var paymentContractAddress string
	var fireflyContractAddress string
	for _, member := range s.Members {
		var fireflyAbiId string
		ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
		if !contractDeployed {
			spinner.Suffix = fmt.Sprintf("publishing payment ABI to '%s'...", member.ID)
			publishPaymentResponse, err := contracts.PublishABI(ethconnectUrl, paymentContract)
			if err != nil {
				return err
			}
			paymentAbiId := publishPaymentResponse.ID
			// TODO: version the registered name
			spinner.Suffix = fmt.Sprintf("deploying payment contract to '%s'...", member.ID)
			deployPaymentResponse, err := contracts.DeployContract(ethconnectUrl, paymentAbiId, member.Address, map[string]string{"initialSupply": "100000000000000000000"}, "payment")
			if err != nil {
				return err
			}
			paymentContractAddress = deployPaymentResponse.ContractAddress

			spinner.Suffix = fmt.Sprintf("publishing FireFly ABI to '%s'...", member.ID)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId := publishFireflyResponse.ID

			// TODO: version the registered name
			spinner.Suffix = fmt.Sprintf("deploying FireFly contract to '%s'...", member.ID)
			deployFireflyResponse, err := contracts.DeployContract(ethconnectUrl, fireflyAbiId, member.Address, map[string]string{"paymentContract": paymentContractAddress}, "firefly")
			if err != nil {
				return err
			}
			fireflyContractAddress = deployFireflyResponse.ContractAddress

			contractDeployed = true
		} else {
			// TODO: Just load the ABI
			spinner.Suffix = fmt.Sprintf("publishing FireFly ABI to '%s'...", member.ID)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId = publishFireflyResponse.ID
		}
		// Register as "firefly"
		spinner.Suffix = fmt.Sprintf("registering FireFly contract on '%s'...", member.ID)
		_, err := contracts.RegisterContract(ethconnectUrl, fireflyAbiId, fireflyContractAddress, member.Address, "firefly", map[string]string{"paymentContract": paymentContractAddress})
		if err != nil {
			return err
		}
	}

	spinner.Suffix = " restarting FireFly nodes..."
	s.restartFireflyNodes()
	return nil
}

func (s *Stack) restartFireflyNodes() error {
	workingDir := path.Join(StacksDir, s.Name)
	for _, member := range s.Members {
		containerName := fmt.Sprintf("%s_firefly_core_%s_1", s.Name, member.ID)
		if err := docker.RunDockerCommand(workingDir, "start", containerName); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stack) extractContracts(containerName string) error {
	workingDir := path.Join(StacksDir, s.Name)
	if err := docker.RunDockerCommand(workingDir, "cp", containerName+":/firefly/contracts", workingDir); err != nil {
		return err
	}
	return nil
}

func (s *Stack) stackHasRunBefore() (bool, error) {
	files, err := ioutil.ReadDir(path.Join(StacksDir, s.Name, "data", "ganache"))
	if err != nil {
		return false, err
	}
	if len(files) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
