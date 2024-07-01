// DockerManager is a mock that implements IDockerManager
package mocks

import "context"

type DockerManager struct{}

func NewDockerManager() *DockerManager {
	return &DockerManager{}
}

func (mgr *DockerManager) RunDockerCommand(ctx context.Context, workingDir string, command ...string) error {
	return nil
}

func (mgr *DockerManager) RunDockerCommandLine(ctx context.Context, workingDir string, command string) error {
	return nil
}

func (mgr *DockerManager) RunDockerComposeCommand(ctx context.Context, workingDir string, command ...string) error {
	return nil
}

func (mgr *DockerManager) RunDockerCommandBuffered(ctx context.Context, workingDir string, command ...string) (string, error) {
	return "", nil
}

func (mgr *DockerManager) RunDockerComposeCommandReturnsStdout(workingDir string, command ...string) ([]byte, error) {
	return nil, nil
}

func (mgr *DockerManager) GetImageConfig(image string) (map[string]interface{}, error) {
	return nil, nil
}

func (mgr *DockerManager) GetImageLabel(image, label string) (string, error) {
	return "", nil
}

func (mgr *DockerManager) GetImageDigest(image string) (string, error) {
	return "", nil
}

func (mgr *DockerManager) CreateVolume(ctx context.Context, volumeName string) error {
	return nil
}

func (mgr *DockerManager) CopyFileToVolume(ctx context.Context, volumeName string, sourcePath string, destPath string) error {
	return nil
}

func (mgr *DockerManager) MkdirInVolume(ctx context.Context, volumeName string, directory string) error {
	return nil
}

func (mgr *DockerManager) RemoveVolume(ctx context.Context, volumeName string) error {
	return nil
}

func (mgr *DockerManager) CopyFromContainer(ctx context.Context, containerName string, sourcePath string, destPath string) error {
	return nil
}
