# FireFly CLI

The FireFly CLI can be used to create a local FireFly stacks for offline development of blockchain apps. This allows developers to rapidly iterate on their idea without worrying about needing to set up a bunch of infrastructure before they can write the first line of code.

![FireFly CLI Screenshot](docs/firefly_screenshot.png)

## Prerequisites

In order to run the FireFly CLI, you will need a few things installed on your dev machine:

- [Docker](https://www.docker.com/)
- [Go](https://golang.org/)

## Install the CLI

On Go 1.16 and newer:

```
$ go install github.com/hyperledger-labs/firefly-cli/ff@latest
```

On earlier versions of Go:

```
$ go get github.com/hyperledger-labs/firefly-cli/ff
```

## Running on Linux

There are a couple of things to be aware of if you're running the FireFly CLI on Linux:

1. Because the FireFly CLI uses Docker, you may encounter some permission issues, depending on how your dev machine is set up. Unless you have set up your user to run Docker without root, it is recommended that you run FireFly CLI commands with `sudo`. For example, to create a new stack run:

```
$ sudo ff init <stack_name>
```

For more information about Docker permissions on Linux, please see [Docker's documentation on the topic](https://docs.docker.com/engine/install/linux-postinstall/).

2. By default, `go install` will install the `ff` binary at `~/go/bin/ff`. If you are running `ff` with `sudo`, the root user will not be able to find the `ff` binary on the path. It is recommended to create a symlink so that the root user can find the `ff` binary on the path.

```
$ sudo ln -s ~/go/bin/ff /usr/bin/ff
```

## Create a new stack

```
$ ff init <stack_name>
```

## Start a stack

```
$ ff start <stack_name>
```

## View logs

```
$ ff logs <stack_name>
```

> **NOTE**: You can use the `-f` flag on the `logs` command to follow the log output from all nodes in the stack

## Stop a stack

```
$ ff stop <stack_name>
```

## Clear all data from a stack

This command clears all data in a stack, but leaves the stack itself. This is useful for testing when you want to start with a clean slate but don't want to actually recreate the resources in the stack itself. Note: this will also stop the stack if it is running.

```
$ ff reset <stack_name>
```

## Completely delete a stack

This command will completely delete a stack, including all of its data and configuration.

```
$ ff remove <stack_name>
```

## Get stack info

This command will print out information about a particular stack, including whether it is running or not.

```
$ ff info <stack_name>
```

## List all stacks

This command will list all stacks that have been created on your machine.

```
$ ff ls
```
