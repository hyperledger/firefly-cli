// Copyright © 2022 Kaleido, Inc.
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
