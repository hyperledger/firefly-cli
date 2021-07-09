package stacks

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger-labs/firefly-cli/internal/contracts"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"golang.org/x/crypto/sha3"

	"gopkg.in/yaml.v2"
)

var homeDir, _ = os.UserHomeDir()
var StacksDir = filepath.Join(homeDir, ".firefly", "stacks")

//go:embed ganache/Dockerfile
var dockerfile []byte

//go:embed ganache/healthcheck.sh
var healthcheck []byte

type DatabaseSelection int

const (
	PostgreSQL DatabaseSelection = iota
	SQLite3
	SQLiteGo
)

var DBSelectionStrings = []string{"postgres", "sqlite3", "sqlitego"}

func (db DatabaseSelection) String() string {
	return DBSelectionStrings[db]
}

func DatabaseSelectionFromString(s string) (DatabaseSelection, error) {
	for i, dbSelection := range DBSelectionStrings {
		if strings.ToLower(s) == dbSelection {
			return DatabaseSelection(i), nil
		}
	}
	return SQLite3, fmt.Errorf("\"%s\" is not a valid database selection. valid options are: %v", s, DBSelectionStrings)
}

type Stack struct {
	Name               string    `json:"name,omitempty"`
	Members            []*Member `json:"members,omitempty"`
	SwarmKey           string    `json:"swarmKey,omitempty"`
	ExposedGanachePort int       `json:"exposedGanachePort,omitempty"`
	Database           string    `json:"database"`
}

type Member struct {
	ID                      string `json:"id,omitempty"`
	Index                   *int   `json:"index,omitempty"`
	Address                 string `json:"address,omitempty"`
	PrivateKey              string `json:"privateKey,omitempty"`
	ExposedFireflyPort      int    `json:"exposedFireflyPort,omitempty"`
	ExposedFireflyAdminPort int    `json:"exposedFireflyAdminPort,omitempty"`
	ExposedEthconnectPort   int    `json:"exposedEthconnectPort,omitempty"`
	ExposedPostgresPort     int    `json:"exposedPostgresPort,omitempty"`
	ExposedDataexchangePort int    `json:"exposedDataexchangePort,omitempty"`
	ExposedIPFSApiPort      int    `json:"exposedIPFSApiPort,omitempty"`
	ExposedIPFSGWPort       int    `json:"exposedIPFSGWPort,omitempty"`
	ExposedUIPort           int    `json:"exposedUiPort ,omitempty"`
}

type StartOptions struct {
	NoPull bool
}

type InitOptions struct {
	FireFlyBasePort   int
	ServicesBasePort  int
	DatabaseSelection string
}

func ListStacks() ([]string, error) {
	files, err := ioutil.ReadDir(StacksDir)
	if err != nil {
		return nil, err
	}

	stacks := make([]string, 0)
	i := 0
	for _, f := range files {
		if f.IsDir() {
			if exists, err := CheckExists(f.Name()); err == nil && exists {
				stacks = append(stacks, f.Name())
				i++
			}
		}
	}
	return stacks, nil
}

func InitStack(stackName string, memberCount int, options *InitOptions) error {

	dbSelection, err := DatabaseSelectionFromString(options.DatabaseSelection)
	if err != nil {
		return err
	}

	stack := &Stack{
		Name:               stackName,
		Members:            make([]*Member, memberCount),
		SwarmKey:           GenerateSwarmKey(),
		ExposedGanachePort: options.ServicesBasePort,
		Database:           dbSelection.String(),
	}

	for i := 0; i < memberCount; i++ {
		stack.Members[i] = createMember(fmt.Sprint(i), i, options)
	}
	compose := CreateDockerCompose(stack)
	if err := stack.ensureDirectories(); err != nil {
		return err
	}
	if err := stack.writeDockerCompose(compose); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %s", err)
	}
	return stack.writeConfigs()
}

func CheckExists(stackName string) (bool, error) {
	_, err := os.Stat(filepath.Join(StacksDir, stackName, "stack.json"))
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
	if d, err := ioutil.ReadFile(filepath.Join(StacksDir, stackName, "stack.json")); err != nil {
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

	stackDir := filepath.Join(StacksDir, s.Name)
	dataDir := filepath.Join(stackDir, "data")

	if err := os.MkdirAll(filepath.Join(stackDir, "ganache"), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(stackDir, "configs"), 0755); err != nil {
		return err
	}

	for _, member := range s.Members {
		if err := os.MkdirAll(filepath.Join(dataDir, "dataexchange_"+member.ID, "peer-certs"), 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stack) writeDockerCompose(compose *DockerComposeConfig) error {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	stackDir := filepath.Join(StacksDir, s.Name)

	if err := ioutil.WriteFile(filepath.Join(stackDir, "ganache", "Dockerfile"), dockerfile, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(stackDir, "ganache", "healthcheck.sh"), healthcheck, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func (s *Stack) writeConfigs() error {
	stackDir := filepath.Join(StacksDir, s.Name)

	fireflyConfigs := NewFireflyConfigs(s)
	for memberId, config := range fireflyConfigs {
		if err := WriteFireflyConfig(config, filepath.Join(stackDir, "configs", fmt.Sprintf("firefly_core_%s.yml", memberId))); err != nil {
			return err
		}
	}

	stackConfigBytes, _ := json.MarshalIndent(s, "", " ")
	if err := ioutil.WriteFile(filepath.Join(stackDir, "stack.json"), stackConfigBytes, 0755); err != nil {
		return err
	}
	return nil
}

func (s *Stack) writeDataExchangeCerts(verbose bool) error {
	stackDir := filepath.Join(StacksDir, s.Name)
	for _, member := range s.Members {

		memberDXDir := path.Join(stackDir, "data", "dataexchange_"+member.ID)

		// TODO: remove dependency on openssl here
		opensslCmd := exec.Command("openssl", "req", "-new", "-x509", "-nodes", "-days", "365", "-subj", fmt.Sprintf("/CN=dataexchange_%s/O=member_%s", member.ID, member.ID), "-keyout", "key.pem", "-out", "cert.pem")
		opensslCmd.Dir = filepath.Join(stackDir, "data", "dataexchange_"+member.ID)
		if err := opensslCmd.Run(); err != nil {
			return err
		}

		dataExchangeConfig := s.GenerateDataExchangeHTTPSConfig(member.ID)
		configBytes, err := json.Marshal(dataExchangeConfig)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(path.Join(memberDXDir, "config.json"), configBytes, 0755)

		// Copy files into docker volumes
		volumeName := fmt.Sprintf("%s_dataexchange_%s", s.Name, member.ID)
		docker.MkdirInVolume(volumeName, "peer-certs", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "config.json"), "/config.json", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "cert.pem"), "/cert.pem", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "key.pem"), "/key.pem", verbose)
	}
	return nil
}

func createMember(id string, index int, options *InitOptions) *Member {
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

	serviceBase := options.ServicesBasePort + (index * 100)
	return &Member{
		ID:                      id,
		Index:                   &index,
		Address:                 encodedAddress,
		PrivateKey:              encodedPrivateKey,
		ExposedFireflyPort:      options.FireFlyBasePort + index,
		ExposedFireflyAdminPort: serviceBase + 1, // note shared ganache is on zero
		ExposedEthconnectPort:   serviceBase + 2,
		ExposedUIPort:           serviceBase + 3,
		ExposedPostgresPort:     serviceBase + 4,
		ExposedDataexchangePort: serviceBase + 5,
		ExposedIPFSApiPort:      serviceBase + 6,
		ExposedIPFSGWPort:       serviceBase + 7,
	}
}

func updateStatus(message string, spin *spinner.Spinner) {
	if spin != nil {
		spin.Suffix = fmt.Sprintf(" %s...", message)
	} else {
		fmt.Println(message)
	}
}

func (s *Stack) StartStack(fancyFeatures bool, verbose bool, options *StartOptions) error {
	fmt.Printf("starting FireFly stack '%s'... ", s.Name)
	// Check to make sure all of our ports are available
	if err := s.checkPortsAvailable(); err != nil {
		return err
	}
	workingDir := filepath.Join(StacksDir, s.Name)
	var spin *spinner.Spinner
	if fancyFeatures && !verbose {
		spin = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		spin.FinalMSG = "done"
	}
	if hasBeenRun, err := s.StackHasRunBefore(); !hasBeenRun && err == nil {
		fmt.Println("\nthis will take a few seconds longer since this is the first time you're running this stack...")
		if spin != nil {
			spin.Start()
		}
		if err := s.runFirstTimeSetup(spin, verbose, options); err != nil {
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
	return docker.RunDockerComposeCommand(filepath.Join(StacksDir, s.Name), verbose, verbose, "stop")
}

func (s *Stack) ResetStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(filepath.Join(StacksDir, s.Name), verbose, verbose, "down", "--rmi", "all", "--volumes"); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(StacksDir, s.Name, "data")); err != nil {
		return err
	}
	return s.ensureDirectories()
}

func (s *Stack) RemoveStack(verbose bool) error {
	if err := s.ResetStack(verbose); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(StacksDir, s.Name))
}

func (s *Stack) checkPortsAvailable() error {
	ports := make([]int, 1)
	ports[0] = s.ExposedGanachePort
	for _, member := range s.Members {
		ports = append(ports, member.ExposedDataexchangePort)
		ports = append(ports, member.ExposedEthconnectPort)
		ports = append(ports, member.ExposedFireflyAdminPort)
		ports = append(ports, member.ExposedFireflyPort)
		ports = append(ports, member.ExposedIPFSApiPort)
		ports = append(ports, member.ExposedIPFSGWPort)
		ports = append(ports, member.ExposedPostgresPort)
		ports = append(ports, member.ExposedUIPort)
	}
	for _, port := range ports {
		if err := checkPortAvailable(port); err != nil {
			return err
		}
	}
	return nil
}

/* This function checks if a TCP port is available by trying to connect to it
* This means the code actually expects an error to be returned when trying to connect
* If an error (of the expected type) is returned, the func will return nil. If it is
* able to connect to something, or an unexpected error occurs, an error will be returned.
 */
func checkPortAvailable(port int) error {
	timeout := time.Millisecond * 500
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)), timeout)

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return nil
	}

	switch t := err.(type) {

	case *net.OpError:
		switch t := t.Unwrap().(type) {
		case *os.SyscallError:
			if t.Syscall == "connect" {
				return nil
			}
		}
		if t.Op == "dial" {
			return err
		} else if t.Op == "read" {
			return nil
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return nil
		}
	}

	if conn != nil {
		defer conn.Close()
		return fmt.Errorf("port %d is unavailable. please check to see if another process is listening on that port.", port)
	}
	return nil
}

func (s *Stack) runFirstTimeSetup(spin *spinner.Spinner, verbose bool, options *StartOptions) error {
	workingDir := filepath.Join(StacksDir, s.Name)
	updateStatus("writing data exchange certs", spin)
	if err := s.writeDataExchangeCerts(verbose); err != nil {
		return err
	}

	// write firefly configs to volumes
	for _, member := range s.Members {
		updateStatus(fmt.Sprintf("copying firefly.core to firefly_core_%s", member.ID), spin)
		volumeName := fmt.Sprintf("%s_firefly_core_%s", s.Name, member.ID)
		if err := docker.CopyFileToVolume(volumeName, path.Join(workingDir, "configs", fmt.Sprintf("firefly_core_%s.yml", member.ID)), "/firefly.core", verbose); err != nil {
			return err
		}
	}

	if !options.NoPull {
		updateStatus("pulling latest versions", spin)
		if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "pull"); err != nil {
			return err
		}
	}

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

func (s *Stack) UpgradeStack(verbose bool) error {
	workingDir := filepath.Join(StacksDir, s.Name)
	if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "down"); err != nil {
		return err
	}
	return docker.RunDockerComposeCommand(workingDir, verbose, verbose, "pull")
}

func (s *Stack) PrintStackInfo(verbose bool) error {
	workingDir := filepath.Join(StacksDir, s.Name)
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(workingDir, verbose, true, "images"); err != nil {
		return err
	}
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(workingDir, verbose, true, "ps"); err != nil {
		return err
	}
	fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(StacksDir, s.Name, "docker-compose.yml"))
	return nil
}

func (s *Stack) deployContracts(spin *spinner.Spinner, verbose bool) error {
	contractDeployed := false
	paymentContract, err := contracts.ReadCompiledContract(filepath.Join(StacksDir, s.Name, "contracts", "Payment.json"))
	if err != nil {
		return err
	}
	fireflyContract, err := contracts.ReadCompiledContract(filepath.Join(StacksDir, s.Name, "contracts", "Firefly.json"))
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
			// Just load the ABI
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
		fmt.Println("setting pre-init false")
		configRecordUrl := fmt.Sprintf("http://localhost:%d/admin/api/v1/config/records/admin", member.ExposedFireflyAdminPort)
		if err := s.httpJSONWithRetry("PUT", configRecordUrl, "{\"preInit\": false}", nil); err != nil && err != io.EOF {
			return err
		}
		fmt.Println("resetting config")
		resetUrl := fmt.Sprintf("http://localhost:%d/admin/api/v1/config/reset", member.ExposedFireflyAdminPort)
		if err := s.httpJSONWithRetry("POST", resetUrl, "{}", nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stack) extractContracts(containerName string, verbose bool) error {
	workingDir := filepath.Join(StacksDir, s.Name)
	destinationDir := filepath.Join(workingDir, "contracts")
	if err := docker.RunDockerCommand(workingDir, verbose, verbose, "cp", containerName+":/firefly/contracts", destinationDir); err != nil {
		return err
	}
	return nil
}

func (s *Stack) StackHasRunBefore() (bool, error) {
	path := filepath.Join(StacksDir, s.Name, "data", fmt.Sprintf("dataexchange_%s", s.Members[0].ID), "cert.pem")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
