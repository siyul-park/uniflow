# 🚀 Getting Started

This comprehensive guide covers how to manage workflows and run the engine using the [Command-Line Interface (CLI)](../cmd/README_kr.md). It includes everything from installation to workflow control and configuration settings.

## Installing from Source

First, set up the [CLI](../cmd/README_kr.md), which allows you to control workflows along with the [built-in extensions](../ext/README_kr.md). Before you start, ensure your system has [Go 1.22](https://go.dev/doc/install) or later installed.

Start by cloning the repository:

```sh
git clone https://github.com/siyul-park/uniflow
```

Navigate to the cloned directory:

```sh
cd uniflow
```

Install dependencies and build the project:

```sh
make init
make build
```

Once the build process is complete, the executable files will be available in the `dist` folder, ready for use.

## Configuration

Uniflow offers flexible configuration options through the `.uniflow.toml` file or system environment variables:

| TOML Key              | Environment Variable Key | Example                     |
|-----------------------|--------------------------|-----------------------------|
| `database.url`        | `DATABASE.URL`           | `mem://` or `mongodb://`    |
| `database.name`       | `DATABASE.NAME`          | -                           |
| `collection.nodes`    | `COLLECTION.NODES`       | `nodes`                     |
| `collection.secrets`  | `COLLECTION.SECRETS`     | `secrets`                   |

If using [MongoDB](https://www.mongodb.com/), enable [change streams](https://www.mongodb.com/docs/manual/changeStreams/) to allow the engine to track changes to node specifications and secrets. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## CLI Commands

The CLI provides various commands for controlling workflows. To see all available commands, run:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command adds or updates node specifications or secrets in a namespace. Use it as follows:

```sh
./dist/uniflow apply nodes --namespace default --filename examples/nodes.yaml
```

or

```sh
./dist/uniflow apply secrets --namespace default --filename examples/secrets.yaml
```

This command outputs the result, and if a namespace is not specified, it uses the `default` namespace.

### Delete

The `delete` command removes node specifications or secrets from a namespace:

```sh
./dist/uniflow delete nodes --namespace default --filename examples/nodes.yaml
```

or

```sh
./dist/uniflow delete secrets --namespace default --filename examples/secrets.yaml
```

This command removes all node specifications or secrets defined in `examples/nodes.yaml` or `examples/secrets.yaml` from the specified namespace. If no namespace is specified, it defaults to the `default` namespace.

### Get

The `get` command retrieves node specifications or secrets from a namespace:

```sh
./dist/uniflow get nodes --namespace default
```

or

```sh
./dist/uniflow get secrets --namespace default
```

This command displays all node specifications or secrets in the specified namespace. If no namespace is specified, it defaults to the `default` namespace.

### Start

The `start` command initiates the runtime with node specifications from a specific namespace:

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can use the `--from-nodes` flag to provide initial node specifications:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.yaml
```

You can also use the `--from-secrets` flag to provide initial secrets:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

This command runs all node specifications in the specified namespace. If no namespace is specified, it defaults to the `default` namespace.

## Integrating HTTP API

To modify node specifications via HTTP API, set up a workflow that exposes this functionality. Utilize the `syscall` node included in the [basic extensions](../ext/README_kr.md):

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

To get started, refer to the [workflow example](../examples/system.yaml). You can add authentication and authorization processes to this workflow as needed. Typically, these runtime control workflows are defined in the `system` namespace.