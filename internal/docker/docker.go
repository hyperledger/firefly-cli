package docker

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func RunDockerCommand(workingDir string, showCommand bool, pipeStdout bool, command ...string) error {
	dockerCmd := exec.Command("docker", command...)
	dockerCmd.Dir = workingDir
	if showCommand {
		fmt.Println(dockerCmd.String())
	}
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	errChan := make(chan error)
	go runCommand(dockerCmd, stdoutChan, stderrChan, errChan)

	for {
		select {
		case s, ok := <-stdoutChan:
			if pipeStdout {
				if !ok {
					return nil
				}
				fmt.Print(s)
			}
		case s, ok := <-stderrChan:
			if !ok {
				return nil
			}
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

func runCommand(cmd *exec.Cmd, stdoutChan chan string, stderrChan chan string, errChan chan error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errChan <- err
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		errChan <- err
		return
	}
	cmd.Start()
	go readPipe(stdout, stdoutChan, errChan)
	go readPipe(stderr, stderrChan, errChan)
}

func readPipe(pipe io.ReadCloser, outputChan chan string, errChan chan error) {
	buf := bufio.NewReader(pipe)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				close(outputChan)
				return
			} else {
				errChan <- err
				close(outputChan)
				return
			}
		} else {
			outputChan <- line
		}
	}
}
