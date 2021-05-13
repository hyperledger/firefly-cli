# FireFly CLI

The FireFly CLI can be used to create a local FireFly stacks for offline development of blockchain apps. This allows developers to rapidly iterate on their idea without worrying about needing to set up a bunch of infrastructure before they can write the first line of code.

![FireFly CLI Screenshot](docs/firefly_screenshot.png)

## Prerequisites

In order to run the FireFly CLI, you will need a few things installed on your dev machine:

- [Docker](https://www.docker.com/)
- [Go](https://golang.org/)

## Install the CLI

```
$ go get https://github.com/kaleido-io/firefly-cli
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

This command clears all data in a stack, but leaves the stack itself. This is useful for testing when you want to start with a clean slate but don't want to actually recreate the resources in the stack itself. The stack must be stopped to run this command.

```
$ ff reset <stack_name>
```

## Completely delete a stack

This command will completely delete a stack, including all of its data and configuration. The stack must be stopped to run this command.

```
$ ff remove <stack_name>
```
