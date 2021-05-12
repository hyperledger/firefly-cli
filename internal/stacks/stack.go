package stacks

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

var homeDir, _ = os.UserHomeDir()
var FireflyDir = path.Join(homeDir, ".firefly")

type Stack struct {
	name    string
	members []*member
}

type member struct {
	id             string
	index          int
	exposedApiPort int
	exposedUiPort  int
}

func InitStack(stackName string, memberCount int) {
	stack := &Stack{
		name:    stackName,
		members: make([]*member, memberCount),
	}
	for i := 0; i < memberCount; i++ {
		stack.members[i] = createMember(fmt.Sprint(i), i)
	}

	compose := CreateDockerCompose(stack)

	ensureDirectories(stack)
	writeDockerCompose(stack.name, compose)
	writeConfigs(stack)
}

func CheckExists(stackName string) bool {
	_, err := os.Stat(path.Join(FireflyDir, stackName))
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ensureDirectories(stack *Stack) {

	dataDir := path.Join(FireflyDir, stack.name, "data")

	for _, member := range stack.members {
		os.MkdirAll(path.Join(dataDir, "postgres_"+member.id), 0755)
	}
	os.MkdirAll(path.Join(dataDir, "ganache"), 0755)
}

func writeDockerCompose(stackName string, compose *DockerCompose) {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		log.Fatal(err)
	}

	stackDir := path.Join(FireflyDir, stackName)
	ioutil.WriteFile(path.Join(stackDir, "docker-compose.yml"), bytes, 0755)
}

func writeConfigs(stack *Stack) {
	configs := NewFireflyConfigs(stack)

	stackDir := path.Join(FireflyDir, stack.name)

	for memberId, config := range configs {
		bytes, _ := yaml.Marshal(config)
		ioutil.WriteFile(path.Join(stackDir, "firefly_"+memberId+".core"), bytes, 0755)
	}
}

func createMember(id string, index int) *member {
	return &member{
		id:             id,
		index:          index,
		exposedApiPort: 5000 + index,
		exposedUiPort:  3000 + index,
	}
}
