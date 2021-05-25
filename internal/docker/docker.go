package docker

import (
	"os/exec"
)

func RunDockerComposeCommand(workingDir string, command ...string) error {
	dockerCmd := exec.Command("docker", append([]string{"compose"}, command...)...)
	dockerCmd.Dir = workingDir
	return dockerCmd.Run()
}

func RunDockerCommand(workingDir string, command ...string) error {
	dockerCmd := exec.Command("docker", command...)
	dockerCmd.Dir = workingDir
	return dockerCmd.Run()
}
