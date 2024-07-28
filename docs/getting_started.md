# 🚀 Getting Started

This comprehensive guide covers how to manage workflows and run the engine using the [Command Line Interface (CLI)](../cmd/README.md). It includes everything from installation to workflow control and configuration settings.

## Installing from Source

First, let's set up the [CLI](../cmd/README.md) to control workflows, including the [built-in extensions](../ext/README.md). Before starting, ensure that [Go 1.22](https://go.dev/doc/install) or higher is installed on your system.

Clone the repository:

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

Once the build process is complete, the executable file will be created in the `dist` folder, ready for use.

## Configuration

Uniflow offers flexible configuration options through a `.uniflow.toml` file or system environment variables:

| TOML Key             | Environment Variable Key | Example                      |
|----------------------|--------------------------|------------------------------|
| `database.url`       | `DATABASE.URL`           | `mem://` or `mongodb://`     |
| `database.name`      | `DATABASE.NAME`          | -                            |
| `collection.nodes`   | `COLLECTION.NODES`       | `nodes`                      |
| `collection.secrets` | `COLLECTION.SECRETS`     | `secrets`                    |

If using [MongoDB](https://www.mongodb.com/), ensure to enable [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) so the engine can track changes in node and secret specifications. This requires setting up a [Replica Set](https://www.mongodb.com/docs/manual/replication/).

## CLI Commands

The CLI provides various commands for controlling workflows. To see all available commands, run:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command adds or updates node specifications or secrets in a namespace. Use the command as follows:

```sh
./dist/uniflow apply nodes --namespace default --filename examples/nodes.json
```

or

```sh
./dist/uniflow apply secrets --namespace default --filename examples/secrets.json
```

This command prints the results and uses the `default` namespace if none is specified.

### Delete

Use the `delete` command to remove node specifications or secrets from a namespace:

```sh
./dist/uniflow delete nodes --namespace default --filename examples/nodes.json
```

or

```sh
./dist/uniflow delete secrets --namespace default --filename examples/secrets.json
```

This command removes all node specifications or secrets defined in `examples/nodes.json` or `examples/secrets.json` from the specified namespace. It defaults to the `default` namespace if not specified.

### Get

The `get` command retrieves node specifications or secrets from a namespace:

```sh
./dist/uniflow get nodes --namespace default
```

or

```sh
./dist/uniflow get secrets --namespace default
```

This command displays all node specifications or secrets in the specified namespace. It defaults to the `default` namespace if not specified.

### Start

The `start` command launches the runtime with node specifications in a specified namespace:

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide initial node specifications using the `--from-nodes` flag:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.json
```

This command runs all node specifications in the specified namespace. It defaults to the `default` namespace if not specified.

## HTTP API Integration

To modify node specifications via the HTTP API, set up a workflow that exposes these functionalities. Utilize the `syscall` nodes included in the [basic extensions](../ext/README.md):

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

To get started, refer to the [workflow example](../examples/system.yaml). You can add authentication and authorization processes as needed. Typically, these runtime control workflows are defined in the `system` namespace.
