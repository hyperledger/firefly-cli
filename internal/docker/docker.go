package docker

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func RunDockerCommand(workingDir string, showCommand bool, pipeStdout bool, command ...string) error {
	if showCommand {
		fmt.Println(append([]string{"docker"}, command...))
	}
	dockerCmd := exec.Command("docker", command...)
	dockerCmd.Dir = workingDir
	stdoutChan := make(chan string)
	errChan := make(chan error)
	go runScript(dockerCmd, stdoutChan, errChan)

	for {
		select {
		case s := <-stdoutChan:
			if pipeStdout {
				fmt.Print(s)
			}
		case err := <-errChan:
			return err
		}
	}
}

func RunDockerComposeCommand(workingDir string, showCommand bool, pipeStdout bool, command ...string) error {
	return RunDockerCommand(workingDir, showCommand, pipeStdout, append([]string{"compose"}, command...)...)
}

func runScript(cmd *exec.Cmd, stdoutChan chan string, errChan chan error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errChan <- err
		close(stdoutChan)
		return
	}
	cmd.Start()
	buf := bufio.NewReader(stdout)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				close(stdoutChan)
				close(errChan)
				return
			}
			errChan <- err
			close(stdoutChan)
			break
		} else {
			stdoutChan <- line
		}
	}
}
