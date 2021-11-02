package docker

import (
	"fmt"
	"os/exec"
)

// CheckDockerConfig is a function to check docker and docker-compose configuration on the host
func CheckDockerConfig() error {

	dockerCmd := exec.Command("docker", "-v")
	_, err := dockerCmd.Output()
	if err != nil {
		return fmt.Errorf("an error occurred while running docker. Is docker installed on your computer?")
	}

	dockerComposeCmd := exec.Command("docker-compose", "-v")
	_, err = dockerComposeCmd.Output()

	if err != nil {
		return fmt.Errorf("an error occurred while running docker-compose. Is docker-compose installed on your computer?")
	}

	dockerDeamonCheck := exec.Command("docker", "ps")
	_, err = dockerDeamonCheck.Output()
	if err != nil {
		return fmt.Errorf("an error occurred while running docker. Is docker running on your computer?")
	}

	return nil
}
