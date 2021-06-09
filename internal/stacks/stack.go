package stacks

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	ID                      string `json:"id,omitempty"`
	Index                   *int   `json:"index,omitempty"`
	Address                 string `json:"address,omitempty"`
	PrivateKey              string `json:"privateKey,omitempty"`
	ExposedFireflyPort      int    `json:"exposedFireflyPort,omitempty"`
	ExposedEthconnectPort   int    `json:"exposedEthconnectPort,omitempty"`
	ExposedPostgresPort     int    `json:"exposedPostgresPort,omitempty"`
	ExposedDataexchangePort int    `json:"exposedDataexchangePort,omitempty"`
	ExposedIPFSApiPort      int    `json:"exposedIPFSApiPort,omitempty`
	ExposedIPFSGWPort       int    `json:"exposedIPFSGWPort,omitempty`
	ExposedUIPort           int    `json:"exposedUiPort ,omitempty"`
}

func InitStack(stackName string, memberCount int) error {
	stack := &Stack{
		Name:     stackName,
		Members:  make([]*Member, memberCount),
		SwarmKey: GenerateSwarmKey(),
	}
	for i := 0; i < memberCount; i++ {
		stack.Members[i] = createMember(fmt.Sprint(i), i)
	}
	compose := CreateDockerCompose(stack)
	if err := stack.ensureDirectories(); err != nil {
		return err
	}
	if err := stack.writeDockerCompose(compose); err != nil {
		return &json.UnmarshalFieldError{}
	}
	if err := stack.writeConfigs(); err != nil {
		return err
	}
	return stack.writeDataExchangeCerts()
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
	fmt.Printf("reading stack config... ")
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

func (s *Stack) ensureDirectories() error {

	dataDir := path.Join(StacksDir, s.Name, "data")

	for _, member := range s.Members {
		if err := os.MkdirAll(path.Join(dataDir, "postgres_"+member.ID), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(path.Join(dataDir, "ipfs_"+member.ID, "staging"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(path.Join(dataDir, "ipfs_"+member.ID, "data"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(path.Join(dataDir, "dataexchange_"+member.ID, "peer-certs"), 0755); err != nil {
			return err
		}
	}
	return os.MkdirAll(path.Join(dataDir, "ganache"), 0755)
}

func (s *Stack) writeDockerCompose(compose *DockerComposeConfig) error {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	stackDir := path.Join(StacksDir, s.Name)
	return ioutil.WriteFile(path.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func (s *Stack) writeConfigs() error {
	stackDir := path.Join(StacksDir, s.Name)

	fireflyConfigs := NewFireflyConfigs(s)
	for memberId, config := range fireflyConfigs {
		if err := WriteFireflyConfig(config, path.Join(stackDir, "firefly_"+memberId+".core")); err != nil {
			return err
		}
	}

	stackConfigBytes, _ := json.MarshalIndent(s, "", " ")
	if err := ioutil.WriteFile(path.Join(stackDir, "stack.json"), stackConfigBytes, 0755); err != nil {
		return err
	}
	return nil
}

func (s *Stack) writeDataExchangeCerts() error {
	stackDir := path.Join(StacksDir, s.Name)
	for _, member := range s.Members {

		// TODO: remove dependency on openssl here
		opensslCmd := exec.Command("openssl", "req", "-new", "-x509", "-nodes", "-days", "365", "-subj", fmt.Sprintf("/CN=dataexchange_%s/O=member_%s", member.ID, member.ID), "-keyout", "key.pem", "-out", "cert.pem")
		opensslCmd.Dir = path.Join(stackDir, "data", "dataexchange_"+member.ID)
		if err := opensslCmd.Run(); err != nil {
			return err
		}

		dataExchangeConfig := s.GenerateDataExchangeHTTPSConfig(member.ID)
		configBytes, err := json.Marshal(dataExchangeConfig)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(path.Join(stackDir, "data", "dataexchange_"+member.ID, "config.json"), configBytes, 0755)
	}
	return nil
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
		ID:                      id,
		Index:                   &index,
		Address:                 encodedAddress,
		PrivateKey:              encodedPrivateKey,
		ExposedFireflyPort:      5000 + index,
		ExposedEthconnectPort:   8080 + index,
		ExposedUIPort:           3000 + index,
		ExposedPostgresPort:     5434 + index,
		ExposedDataexchangePort: 3020 + index,
		ExposedIPFSApiPort:      6000 + index,
		ExposedIPFSGWPort:       6100 + index,
	}
}

func updateStatus(message string, spin *spinner.Spinner) {
	if spin != nil {
		spin.Suffix = fmt.Sprintf(" %s...", message)
	} else {
		fmt.Println(message)
	}
}

func (s *Stack) StartStack(fancyFeatures bool, verbose bool) error {
	fmt.Printf("starting FireFly stack '%s'... ", s.Name)
	workingDir := path.Join(StacksDir, s.Name)
	var spin *spinner.Spinner
	if fancyFeatures {
		spin = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		spin.FinalMSG = "done"
	}
	if hasBeenRun, err := s.stackHasRunBefore(); !hasBeenRun && err == nil {
		fmt.Println("\nthis will take a few seconds longer since this is the first time you're running this stack...")
		if spin != nil {
			spin.Start()
		}
		if err := s.runFirstTimeSetup(spin, verbose); err != nil {
			if spin != nil {
				spin.Stop()
			}
			return err
		}
		if spin != nil {
			spin.Stop()
		}
		return nil
	} else if err == nil {
		if spin != nil {
			spin.Start()
		}
		updateStatus("starting FireFly dependencies", spin)
		err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "up", "-d")
		if spin != nil {
			spin.Stop()
		}
		return err
	} else {
		if spin != nil {
			spin.Stop()
		}
		return err
	}
}

func (s *Stack) StopStack(verbose bool) error {
	return docker.RunDockerComposeCommand(path.Join(StacksDir, s.Name), verbose, verbose, "stop")
}

func (s *Stack) ResetStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(path.Join(StacksDir, s.Name), verbose, verbose, "down", "--rmi", "all"); err != nil {
		return err
	}
	if err := os.RemoveAll(path.Join(StacksDir, s.Name, "data")); err != nil {
		return err
	}
	return s.ensureDirectories()
}

func (s *Stack) RemoveStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(path.Join(StacksDir, s.Name), verbose, verbose, "rm", "-f"); err != nil {
		return err
	}
	return os.RemoveAll(path.Join(StacksDir, s.Name))
}

func (s *Stack) runFirstTimeSetup(spin *spinner.Spinner, verbose bool) error {
	workingDir := path.Join(StacksDir, s.Name)
	updateStatus("starting FireFly dependencies", spin)
	if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "up", "-d"); err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s_firefly_core_%s_1", s.Name, s.Members[0].ID)
	updateStatus("extracting smart contracts", spin)
	if err := s.extractContracts(containerName, verbose); err != nil {
		return err
	}
	updateStatus("deploying smart contracts", spin)
	if err := s.deployContracts(spin, verbose); err != nil {
		return err
	}
	updateStatus("registering FireFly identities", spin)
	if err := s.registerFireflyIdentities(spin, verbose); err != nil {
		return err
	}
	return nil
}

func (s *Stack) deployContracts(spin *spinner.Spinner, verbose bool) error {
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
			updateStatus(fmt.Sprintf("publishing payment ABI to '%s'", member.ID), spin)
			publishPaymentResponse, err := contracts.PublishABI(ethconnectUrl, paymentContract)
			if err != nil {
				return err
			}
			paymentAbiId := publishPaymentResponse.ID
			// TODO: version the registered name
			updateStatus(fmt.Sprintf("deploying payment contract to '%s'", member.ID), spin)
			deployPaymentResponse, err := contracts.DeployContract(ethconnectUrl, paymentAbiId, member.Address, map[string]string{"initialSupply": "100000000000000000000"}, "payment")
			if err != nil {
				return err
			}
			paymentContractAddress = deployPaymentResponse.ContractAddress

			updateStatus(fmt.Sprintf("publishing FireFly ABI to '%s'", member.ID), spin)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId := publishFireflyResponse.ID

			// TODO: version the registered name
			updateStatus(fmt.Sprintf("deploying FireFly contract to '%s'", member.ID), spin)
			deployFireflyResponse, err := contracts.DeployContract(ethconnectUrl, fireflyAbiId, member.Address, map[string]string{"paymentContract": paymentContractAddress}, "firefly")
			if err != nil {
				return err
			}
			fireflyContractAddress = deployFireflyResponse.ContractAddress

			contractDeployed = true
		} else {
			// TODO: Just load the ABI
			updateStatus(fmt.Sprintf("publishing FireFly ABI to '%s'", member.ID), spin)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId = publishFireflyResponse.ID
		}
		// Register as "firefly"
		updateStatus(fmt.Sprintf("registering FireFly contract on '%s'", member.ID), spin)
		_, err := contracts.RegisterContract(ethconnectUrl, fireflyAbiId, fireflyContractAddress, member.Address, "firefly", map[string]string{"paymentContract": paymentContractAddress})
		if err != nil {
			return err
		}
	}

	updateStatus("restarting FireFly nodes", spin)
	if err := s.patchConfigAndRestartFireflyNodes(verbose); err != nil {
		return err
	}

	return nil
}

func (s *Stack) patchConfigAndRestartFireflyNodes(verbose bool) error {
	for _, member := range s.Members {
		containerName := fmt.Sprintf("%s_firefly_core_%s_1", s.Name, member.ID)
		if err := s.stopFirelyNode(containerName, verbose); err != nil {
			return err
		}
		configFilePath := path.Join(StacksDir, s.Name, "firefly_"+member.ID+".core")
		config, err := ReadFireflyConfig(configFilePath)
		if err != nil {
			return err
		}
		config.Blockchain.Ethereum.Ethconnect.SkipEventStreamInit = false
		WriteFireflyConfig(config, configFilePath)
		if err := s.startFireflyNode(containerName, verbose); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stack) restartFireflyNodes(verbose bool) error {
	for _, member := range s.Members {
		containerName := fmt.Sprintf("%s_firefly_core_%s_1", s.Name, member.ID)
		if err := s.restartFireflyNode(containerName, verbose); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stack) stopFirelyNode(containerName string, verbose bool) error {
	workingDir := path.Join(StacksDir, s.Name)
	return docker.RunDockerCommand(workingDir, verbose, verbose, "stop", containerName)
}

func (s *Stack) startFireflyNode(containerName string, verbose bool) error {
	workingDir := path.Join(StacksDir, s.Name)
	return docker.RunDockerCommand(workingDir, verbose, verbose, "start", containerName)
}

func (s *Stack) restartFireflyNode(containerName string, verbose bool) error {
	workingDir := path.Join(StacksDir, s.Name)
	return docker.RunDockerCommand(workingDir, verbose, verbose, "restart", containerName)
}

func (s *Stack) extractContracts(containerName string, verbose bool) error {
	workingDir := path.Join(StacksDir, s.Name)
	destinationDir := path.Join(workingDir, "contracts")
	if err := docker.RunDockerCommand(workingDir, verbose, verbose, "cp", containerName+":/firefly/contracts", destinationDir); err != nil {
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
