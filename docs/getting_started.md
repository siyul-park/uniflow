# 🚀 Getting Started

This comprehensive guide covers how to manage workflows and run the engine using the [Command Line Interface (CLI)](../cmd/README.md). It includes everything from installation to workflow control and configuration settings.

## Installing from Source

First, set up the CLI, which allows you to control workflows with the [built-in extensions](../ext/README.md). Before starting, ensure that [Go 1.22](https://go.dev/doc/install) or later is installed on your system.

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

Once the build process is complete, the executable will be created in the `dist` folder and will be ready for use.

## Configuration

Uniflow offers flexible configuration options through the `.uniflow.toml` file or system environment variables:

| TOML Key             | Environment Variable Key | Example                    |
|----------------------|--------------------------|----------------------------|
| `database.url`       | `DATABASE.URL`           | `mem://` or `mongodb://`   |
| `database.name`      | `DATABASE.NAME`          | -                          |
| `collection.nodes`   | `COLLECTION.NODES`       | `nodes`                    |
| `collection.secrets` | `COLLECTION.SECRETS`     | `secrets`                  |

When using [MongoDB](https://www.mongodb.com/), you need to enable [change streams](https://www.mongodb.com/docs/manual/changeStreams/) so the engine can track changes to node and secret specifications. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## CLI Commands

The CLI offers various commands for controlling workflows. To see all available commands, run:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command adds or updates node specifications or secrets in a namespace. You can use the command as follows:

```sh
./dist/uniflow apply nodes --namespace default --filename examples/nodes.json
```

Or

```sh
./dist/uniflow apply secrets --namespace default --filename examples/secrets.json
```

This command outputs the results and uses the `default` namespace if none is specified.

### Delete

The `delete` command removes node specifications or secrets from a namespace:

```sh
./dist/uniflow delete nodes --namespace default --filename examples/nodes.json
```

Or

```sh
./dist/uniflow delete secrets --namespace default --filename examples/secrets.json
```

This command removes all node specifications or secrets defined in `examples/nodes.json` or `examples/secrets.json` from the specified namespace. If no namespace is specified, the `default` namespace is used.

### Get

The `get` command retrieves node specifications or secrets from a namespace:

```sh
./dist/uniflow get nodes --namespace default
```

Or

```sh
./dist/uniflow get secrets --namespace default
```

This command displays all node specifications or secrets in the specified namespace. If no namespace is specified, the `default` namespace is used.

### Start

The `start` command initiates the runtime with node specifications in a specific namespace:

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide initial node specifications using the `--from-nodes` flag:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.json
```

You can also provide initial secrets using the `--from-secrets` flag:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.json
```

This command runs all node specifications in the specified namespace. If no namespace is specified, the `default` namespace is used.

## HTTP API Integration

To modify node specifications through the HTTP API, you need to set up a workflow that exposes this functionality. You can use the `syscall` node included in the [default extensions](../ext/README.md):

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

To get started, refer to the [workflow example](../examples/system.yaml). You can add authentication and authorization processes to this workflow as needed. Generally, these runtime control workflows are defined in the `system` namespace.
