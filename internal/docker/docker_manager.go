// Copyright © 2024 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"context"
)

// DockerInterface combines all Docker-related operations into a single interface.
type IDockerManager interface {
	// Command Execution
	RunDockerCommand(ctx context.Context, workingDir string, command ...string) error
	RunDockerCommandLine(ctx context.Context, workingDir string, command string) error
	RunDockerComposeCommand(ctx context.Context, workingDir string, command ...string) error
	RunDockerCommandBuffered(ctx context.Context, workingDir string, command ...string) (string, error)
	RunDockerComposeCommandReturnsStdout(workingDir string, command ...string) ([]byte, error)

	// Image Inspection
	GetImageConfig(image string) (map[string]interface{}, error)
	GetImageLabel(image, label string) (string, error)
	GetImageDigest(image string) (string, error)

	// Volume Management
	CreateVolume(ctx context.Context, volumeName string) error
	CopyFileToVolume(ctx context.Context, volumeName string, sourcePath string, destPath string) error
	MkdirInVolume(ctx context.Context, volumeName string, directory string) error
	RemoveVolume(ctx context.Context, volumeName string) error

	// Container Interaction
	CopyFromContainer(ctx context.Context, containerName string, sourcePath string, destPath string) error
}

// DockerManager implements IDockerManager
type DockerManager struct{}

func NewDockerManager() *DockerManager {
	return &DockerManager{}
}

func (mgr *DockerManager) RunDockerCommand(ctx context.Context, workingDir string, command ...string) error {
	return RunDockerCommand(ctx, workingDir, command...)
}

func (mgr *DockerManager) RunDockerCommandLine(ctx context.Context, workingDir string, command string) error {
	return RunDockerCommandLine(ctx, workingDir, command)
}

func (mgr *DockerManager) RunDockerComposeCommand(ctx context.Context, workingDir string, command ...string) error {
	return RunDockerComposeCommand(ctx, workingDir, command...)
}

func (mgr *DockerManager) RunDockerCommandBuffered(ctx context.Context, workingDir string, command ...string) (string, error) {
	return RunDockerCommandBuffered(ctx, workingDir, command...)
}

func (mgr *DockerManager) RunDockerComposeCommandReturnsStdout(workingDir string, command ...string) ([]byte, error) {
	return RunDockerComposeCommandReturnsStdout(workingDir, command...)
}

func (mgr *DockerManager) GetImageConfig(image string) (map[string]interface{}, error) {
	return GetImageConfig(image)
}

func (mgr *DockerManager) GetImageLabel(image, label string) (string, error) {
	return GetImageLabel(image, label)
}

func (mgr *DockerManager) GetImageDigest(image string) (string, error) {
	return GetImageDigest(image)
}

func (mgr *DockerManager) CreateVolume(ctx context.Context, volumeName string) error {
	return CreateVolume(ctx, volumeName)
}

func (mgr *DockerManager) CopyFileToVolume(ctx context.Context, volumeName string, sourcePath string, destPath string) error {
	return CopyFileToVolume(ctx, volumeName, sourcePath, destPath)
}

func (mgr *DockerManager) MkdirInVolume(ctx context.Context, volumeName string, directory string) error {
	return MkdirInVolume(ctx, volumeName, directory)
}

func (mgr *DockerManager) RemoveVolume(ctx context.Context, volumeName string) error {
	return RemoveVolume(ctx, volumeName)
}

func (mgr *DockerManager) CopyFromContainer(ctx context.Context, containerName string, sourcePath string, destPath string) error {
	return CopyFromContainer(ctx, containerName, sourcePath, destPath)
}
