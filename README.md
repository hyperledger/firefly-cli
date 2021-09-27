# FireFly CLI

![build](https://github.com/hyperledger/firefly-cli/actions/workflows/build.yml/badge.svg?branch=main)

The FireFly CLI can be used to create local [FireFly](https://github.com/hyperledger/firefly) stacks
for offline development of blockchain apps. This allows developers to rapidly iterate on their idea without
needing to set up a bunch of infrastructure before they can write the first line of code.

![FireFly CLI Screenshot](docs/firefly_screenshot.png)

## Prerequisites

In order to run the FireFly CLI, you will need a few things installed on your dev machine:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/)
- openssl

## Install the CLI

On Go 1.16 and newer:

```
$ go install github.com/hyperledger/firefly-cli/ff@latest
```

On earlier versions of Go:

```
$ go get github.com/hyperledger/firefly-cli/ff
```

> **NOTE**: For Linux users, it is recommended that you add your user to the `docker` group so that you do not have to run `ff` or `docker` as `root` or with `sudo`. For more information about Docker permissions on Linux, please see [Docker's documentation on the topic](https://docs.docker.com/engine/install/linux-postinstall/).

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
