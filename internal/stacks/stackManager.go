package stacks

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain"
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum/besu"
	"github.com/hyperledger-labs/firefly-cli/internal/blockchain/ethereum/geth"
	"github.com/hyperledger-labs/firefly-cli/internal/constants"
	"github.com/hyperledger-labs/firefly-cli/internal/contracts"
	"github.com/hyperledger-labs/firefly-cli/internal/docker"
	"github.com/hyperledger-labs/firefly-cli/pkg/types"
	"golang.org/x/crypto/sha3"

	"gopkg.in/yaml.v2"
)

type StackManager struct {
	Stack *types.Stack
}

type StartOptions struct {
	NoPull     bool
	NoRollback bool
}

type InitOptions struct {
	FireFlyBasePort    int
	ServicesBasePort   int
	DatabaseSelection  DatabaseSelection
	Verbose            bool
	ExternalProcesses  int
	BlockchainProvider BlockchainProvider
}

func ListStacks() ([]string, error) {
	files, err := ioutil.ReadDir(constants.StacksDir)
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

func NewStackManager() *StackManager {
	return &StackManager{}
}

func (s *StackManager) InitStack(stackName string, memberCount int, options *InitOptions) error {
	s.Stack = &types.Stack{
		Name:                  stackName,
		Members:               make([]*types.Member, memberCount),
		SwarmKey:              GenerateSwarmKey(),
		ExposedBlockchainPort: options.ServicesBasePort,
		Database:              options.DatabaseSelection.String(),
		BlockchainProvider:    options.BlockchainProvider.String(),
	}

	blockchainProvider := s.getBlockchainProvider(false)

	for i := 0; i < memberCount; i++ {
		externalProcess := i < options.ExternalProcesses
		s.Stack.Members[i] = createMember(fmt.Sprint(i), i, options, externalProcess)
	}
	compose := docker.CreateDockerCompose(s.Stack)
	blockchainServiceName, blockchainServiceDefinition := blockchainProvider.GetDockerServiceDefinition()
	compose.Services[blockchainServiceName] = blockchainServiceDefinition
	compose.Volumes[blockchainServiceName] = struct{}{}

	if err := s.ensureDirectories(); err != nil {
		return err
	}
	if err := s.writeDockerCompose(compose); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %s", err)
	}
	return s.writeConfigs(options.Verbose)
}

func CheckExists(stackName string) (bool, error) {
	_, err := os.Stat(filepath.Join(constants.StacksDir, stackName, "stack.json"))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (s *StackManager) LoadStack(stackName string) error {
	exists, err := CheckExists(stackName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("stack '%s' does not exist", stackName)
	}
	fmt.Printf("reading stack config... ")
	if d, err := ioutil.ReadFile(filepath.Join(constants.StacksDir, stackName, "stack.json")); err != nil {
		return err
	} else {
		var stack *types.Stack
		if err := json.Unmarshal(d, &stack); err == nil {
			fmt.Printf("done\n")
		}
		s.Stack = stack
	}
	return nil
}

func (s *StackManager) ensureDirectories() error {

	stackDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	dataDir := filepath.Join(stackDir, "data")

	if err := os.MkdirAll(filepath.Join(stackDir, "configs"), 0755); err != nil {
		return err
	}

	for _, member := range s.Stack.Members {
		if err := os.MkdirAll(filepath.Join(dataDir, "dataexchange_"+member.ID, "peer-certs"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(stackDir, "blockchain", member.ID), 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) writeDockerCompose(compose *docker.DockerComposeConfig) error {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	stackDir := filepath.Join(constants.StacksDir, s.Stack.Name)

	return ioutil.WriteFile(filepath.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func (s *StackManager) writeConfigs(verbose bool) error {
	stackDir := filepath.Join(constants.StacksDir, s.Stack.Name)

	fireflyConfigs := NewFireflyConfigs(s.Stack)
	for memberId, config := range fireflyConfigs {
		if err := WriteFireflyConfig(config, filepath.Join(stackDir, "configs", fmt.Sprintf("firefly_core_%s.yml", memberId))); err != nil {
			return err
		}
	}

	stackConfigBytes, _ := json.MarshalIndent(s, "", " ")
	if err := ioutil.WriteFile(filepath.Join(stackDir, "stack.json"), stackConfigBytes, 0755); err != nil {
		return err
	}

	p := s.getBlockchainProvider(verbose)
	if err := p.WriteConfig(); err != nil {
		return err
	}

	return nil
}

func (s *StackManager) writeDataExchangeCerts(verbose bool) error {
	stackDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	for _, member := range s.Stack.Members {

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
		volumeName := fmt.Sprintf("%s_dataexchange_%s", s.Stack.Name, member.ID)
		docker.MkdirInVolume(volumeName, "peer-certs", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "config.json"), "/config.json", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "cert.pem"), "/cert.pem", verbose)
		docker.CopyFileToVolume(volumeName, path.Join(memberDXDir, "key.pem"), "/key.pem", verbose)
	}
	return nil
}

func createMember(id string, index int, options *InitOptions, external bool) *types.Member {
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
	return &types.Member{
		ID:                      id,
		Index:                   &index,
		Address:                 encodedAddress,
		PrivateKey:              encodedPrivateKey,
		ExposedFireflyPort:      options.FireFlyBasePort + index,
		ExposedFireflyAdminPort: serviceBase + 1, // note shared blockchain node is on zero
		ExposedEthconnectPort:   serviceBase + 2,
		ExposedUIPort:           serviceBase + 3,
		ExposedPostgresPort:     serviceBase + 4,
		ExposedDataexchangePort: serviceBase + 5,
		ExposedIPFSApiPort:      serviceBase + 6,
		ExposedIPFSGWPort:       serviceBase + 7,
		External:                external,
	}
}

func updateStatus(message string, spin *spinner.Spinner) {
	if spin != nil {
		spin.Suffix = fmt.Sprintf(" %s...", message)
	} else {
		fmt.Println(message)
	}
}

func (s *StackManager) StartStack(fancyFeatures bool, verbose bool, options *StartOptions) error {
	blockchainProvider := s.getBlockchainProvider(verbose)
	fmt.Printf("starting FireFly stack '%s'... ", s.Stack.Name)
	// Check to make sure all of our ports are available
	if err := s.checkPortsAvailable(); err != nil {
		return err
	}
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
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
			// Something bad happened during setup
			if options.NoRollback {
				return err
			} else {
				// Rollback changes
				updateStatus("an error occurred - rolling back changes", spin)
				resetErr := s.ResetStack(verbose)
				if spin != nil {
					spin.Stop()
				}

				var finalErr error

				if resetErr != nil {
					finalErr = fmt.Errorf("%s - error resetting stack: %s", err.Error(), resetErr.Error())
				} else {
					finalErr = fmt.Errorf("%s - all changes rolled back", err.Error())
				}

				return finalErr
			}
		}
		if spin != nil {
			spin.Stop()
		}
		return nil
	} else if err == nil {
		if spin != nil {
			spin.Start()
		}

		if err := blockchainProvider.PreStart(); err != nil {
			return err
		}

		updateStatus("starting FireFly dependencies", spin)
		if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "up", "-d"); err != nil {
			return err
		}

		if err := blockchainProvider.PostStart(); err != nil {
			return err
		}

		if err := s.ensureFireflyNodesUp(false, spin); err != nil {
			return err
		}

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

func (s *StackManager) StopStack(verbose bool) error {
	return docker.RunDockerComposeCommand(filepath.Join(constants.StacksDir, s.Stack.Name), verbose, verbose, "stop")
}

func (s *StackManager) ResetStack(verbose bool) error {
	if err := docker.RunDockerComposeCommand(filepath.Join(constants.StacksDir, s.Stack.Name), verbose, verbose, "down", "--volumes"); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(constants.StacksDir, s.Stack.Name, "data")); err != nil {
		return err
	}
	return s.ensureDirectories()
}

func (s *StackManager) RemoveStack(verbose bool) error {
	if err := s.ResetStack(verbose); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(constants.StacksDir, s.Stack.Name))
}

func (s *StackManager) checkPortsAvailable() error {
	ports := make([]int, 1)
	ports[0] = s.Stack.ExposedBlockchainPort
	for _, member := range s.Stack.Members {
		ports = append(ports, member.ExposedDataexchangePort)
		ports = append(ports, member.ExposedEthconnectPort)
		if !member.External {
			ports = append(ports, member.ExposedFireflyAdminPort)
			ports = append(ports, member.ExposedFireflyPort)
		}
		ports = append(ports, member.ExposedIPFSApiPort)
		ports = append(ports, member.ExposedIPFSGWPort)
		ports = append(ports, member.ExposedPostgresPort)
		ports = append(ports, member.ExposedUIPort)
	}
	for _, port := range ports {
		available, err := checkPortAvailable(port)
		if err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("port %d is unavailable. please check to see if another process is listening on that port", port)
		}
	}
	return nil
}

func checkPortAvailable(port int) (bool, error) {
	timeout := time.Millisecond * 500
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)), timeout)

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true, nil
	}

	switch t := err.(type) {

	case *net.OpError:
		switch t := t.Unwrap().(type) {
		case *os.SyscallError:
			if t.Syscall == "connect" {
				return true, nil
			}
		}
		if t.Op == "dial" {
			return false, err
		} else if t.Op == "read" {
			return true, nil
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return true, nil
		}
	}

	if conn != nil {
		defer conn.Close()
		return false, nil
	}
	return true, nil
}

func (s *StackManager) runFirstTimeSetup(spin *spinner.Spinner, verbose bool, options *StartOptions) error {
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	blockchainProvider := s.getBlockchainProvider(verbose)

	updateStatus("initializing blockchain node", spin)
	if err := blockchainProvider.Init(); err != nil {
		return err
	}

	updateStatus("writing data exchange certs", spin)
	if err := s.writeDataExchangeCerts(verbose); err != nil {
		return err
	}

	// write firefly configs to volumes
	for _, member := range s.Stack.Members {
		if !member.External {
			updateStatus(fmt.Sprintf("copying firefly.core to firefly_core_%s", member.ID), spin)
			volumeName := fmt.Sprintf("%s_firefly_core_%s", s.Stack.Name, member.ID)
			if err := docker.CopyFileToVolume(volumeName, path.Join(workingDir, "configs", fmt.Sprintf("firefly_core_%s.yml", member.ID)), "/firefly.core", verbose); err != nil {
				return err
			}
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

	if err := blockchainProvider.PostStart(); err != nil {
		return err
	}

	var containerName string
	for _, member := range s.Stack.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_firefly_core_%s_1", s.Stack.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return errors.New("unable to extract contracts from container - no valid firefly core containers found in stack")
	}
	updateStatus("extracting smart contracts", spin)
	if err := s.extractContracts(containerName, verbose); err != nil {
		return err
	}

	if err := s.ensureFireflyNodesUp(true, spin); err != nil {
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

func (s *StackManager) ensureFireflyNodesUp(firstTimeSetup bool, spin *spinner.Spinner) error {
	for _, member := range s.Stack.Members {
		if member.External {
			configFilename := path.Join(constants.StacksDir, s.Stack.Name, "configs", fmt.Sprintf("firefly_core_%v.yml", member.ID))
			var port int
			if firstTimeSetup {
				port = member.ExposedFireflyAdminPort
			} else {
				port = member.ExposedFireflyPort
			}
			// Check process running
			available, err := checkPortAvailable(port)
			if err != nil {
				return err
			}
			if available {
				updateStatus(fmt.Sprintf("please start your firefly core with the config file for this stack: firefly -f %s  ", configFilename), spin)
				if err := s.waitForFireflyStart(port); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *StackManager) waitForFireflyStart(port int) error {
	retries := 120
	retryPeriod := 1000 // ms
	retriesRemaining := retries
	for retriesRemaining > 0 {
		time.Sleep(time.Duration(retryPeriod) * time.Millisecond)
		available, err := checkPortAvailable(port)
		if err != nil {
			return err
		}
		if !available {
			return nil
		}
		retriesRemaining--
	}
	return fmt.Errorf("waited for %v seconds for firefly to start on port %v but it was never available", retries*retryPeriod/1000, port)
}

func (s *StackManager) UpgradeStack(verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	if err := docker.RunDockerComposeCommand(workingDir, verbose, verbose, "down"); err != nil {
		return err
	}
	return docker.RunDockerComposeCommand(workingDir, verbose, verbose, "pull")
}

func (s *StackManager) PrintStackInfo(verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(workingDir, verbose, true, "images"); err != nil {
		return err
	}
	fmt.Print("\n")
	if err := docker.RunDockerComposeCommand(workingDir, verbose, true, "ps"); err != nil {
		return err
	}
	fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(constants.StacksDir, s.Stack.Name, "docker-compose.yml"))
	return nil
}

func (s *StackManager) deployContracts(spin *spinner.Spinner, verbose bool) error {
	contractDeployed := false
	fireflyContract, err := contracts.ReadCompiledContract(filepath.Join(constants.StacksDir, s.Stack.Name, "contracts", "Firefly.json"))
	if err != nil {
		return err
	}
	var fireflyContractAddress string
	for _, member := range s.Stack.Members {
		ethconnectUrl := fmt.Sprintf("http://127.0.0.1:%v", member.ExposedEthconnectPort)
		if !contractDeployed {
			updateStatus(fmt.Sprintf("publishing FireFly ABI to '%s'", member.ID), spin)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId := publishFireflyResponse.ID

			// TODO: version the registered name
			updateStatus(fmt.Sprintf("deploying FireFly contract to '%s'", member.ID), spin)
			deployFireflyResponse, err := contracts.DeployContract(ethconnectUrl, fireflyAbiId, member.Address, map[string]string{}, "firefly")
			if err != nil {
				return err
			}
			fireflyContractAddress = deployFireflyResponse.ContractAddress

			contractDeployed = true
		} else {
			updateStatus(fmt.Sprintf("publishing FireFly ABI to '%s'", member.ID), spin)
			publishFireflyResponse, err := contracts.PublishABI(ethconnectUrl, fireflyContract)
			if err != nil {
				return err
			}
			fireflyAbiId := publishFireflyResponse.ID

			updateStatus(fmt.Sprintf("registering FireFly contract on '%s'", member.ID), spin)
			_, err = contracts.RegisterContract(ethconnectUrl, fireflyAbiId, fireflyContractAddress, member.Address, "firefly", map[string]string{})
			if err != nil {
				return err
			}
		}
	}

	if err := s.patchConfigAndRestartFireflyNodes(verbose, spin); err != nil {
		return err
	}

	return nil
}

func (s *StackManager) patchConfigAndRestartFireflyNodes(verbose bool, spin *spinner.Spinner) error {
	for _, member := range s.Stack.Members {
		updateStatus(fmt.Sprintf("applying configuration changes to %s", member.ID), spin)
		configRecordUrl := fmt.Sprintf("http://localhost:%d/admin/api/v1/config/records/admin", member.ExposedFireflyAdminPort)
		if err := s.httpJSONWithRetry("PUT", configRecordUrl, "{\"preInit\": false}", nil); err != nil && err != io.EOF {
			return err
		}
		resetUrl := fmt.Sprintf("http://localhost:%d/admin/api/v1/config/reset", member.ExposedFireflyAdminPort)
		if err := s.httpJSONWithRetry("POST", resetUrl, "{}", nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *StackManager) extractContracts(containerName string, verbose bool) error {
	workingDir := filepath.Join(constants.StacksDir, s.Stack.Name)
	destinationDir := filepath.Join(workingDir, "contracts")
	if err := docker.RunDockerCommand(workingDir, verbose, verbose, "cp", containerName+":/firefly/contracts", destinationDir); err != nil {
		return err
	}
	return nil
}

func (s *StackManager) StackHasRunBefore() (bool, error) {
	path := filepath.Join(constants.StacksDir, s.Stack.Name, "data", fmt.Sprintf("dataexchange_%s", s.Stack.Members[0].ID), "cert.pem")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (s *StackManager) getBlockchainProvider(verbose bool) blockchain.IBlockchainProvider {
	switch s.Stack.BlockchainProvider {
	case GoEthereum.String():
		return &geth.GethProvider{
			Verbose: verbose,
			Stack:   s.Stack,
		}
	case HyperledgerBesu.String():
		return &besu.BesuProvider{
			Verbose: verbose,
			Stack:   s.Stack,
		}
	default:
		return nil
	}

}
