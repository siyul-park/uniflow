# 🚀 Getting Started

This guide provides detailed instructions on how to install, configure, and manage workflows using the [Command Line Interface (CLI)](../cmd/README.md). It covers the entire process from installation to workflow control and configuration.

## Installing from Source

To begin, you need to set up the [CLI](../cmd/README.md) along with the [built-in extensions](../ext/README.md). Before starting the installation, ensure that [Go 1.23](https://go.dev/doc/install) or higher is installed on your system.

### Cloning the Repository

To clone the source code, run the following command:

```sh
git clone https://github.com/siyul-park/uniflow
```

Navigate to the cloned directory:

```sh
cd uniflow
```

### Installing Dependencies and Building

To install dependencies and build the project, execute the following commands:

```sh
make init
make build
```

Once the build is complete, the executable will be located in the `dist` folder.

### Configuration

You can flexibly modify settings via the `.uniflow.toml` file or system environment variables. Key configuration options include:

| TOML Key              | Environment Variable Key  | Example                    |
|-----------------------|----------------------------|----------------------------|
| `database.url`        | `DATABASE.URL`             | `mem://` or `mongodb://`   |
| `database.name`       | `DATABASE.NAME`            | -                          |
| `collection.nodes`    | `COLLECTION.NODES`         | `nodes`                    |
| `collection.secrets`  | `COLLECTION.SECRETS`       | `secrets`                  |

If using [MongoDB](https://www.mongodb.com/), enable [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) so that the engine can track node specifications and secret changes. This requires setting up a [Replica Set](https://www.mongodb.com/docs/manual/replication/).

## Uniflow

`uniflow` is primarily used to start and manage runtime environments.

### Start

The `start` command initiates the runtime with node specifications for a specific namespace. The basic usage is as follows:

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide initial node specifications using the `--from-nodes` flag:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.yaml
```

To provide initial secrets, use the `--from-secrets` flag:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

This command will execute all node specifications for the specified namespace. If no namespace is specified, the `default` namespace is used.

## Uniflowctl

`uniflowctl` is used to manage node specifications and secrets within a namespace.

### Apply

The `apply` command adds or updates node specifications or secrets in a namespace. Usage examples are:

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/nodes.yaml
```

or

```sh
./dist/uniflowctl apply secrets --namespace default --filename examples/secrets.yaml
```

This command applies the contents of the specified file to the namespace. If no namespace is specified, the `default` namespace is used by default.

### Delete

The `delete` command removes node specifications or secrets from a namespace. Usage examples are:

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/nodes.yaml
```

or

```sh
./dist/uniflowctl delete secrets --namespace default --filename examples/secrets.yaml
```

This command removes all node specifications or secrets defined in the specified file. If no namespace is specified, the `default` namespace is used.

### Get

The `get` command retrieves node specifications or secrets from a namespace. Usage examples are:

```sh
./dist/uniflowctl get nodes --namespace default
```

or

```sh
./dist/uniflowctl get secrets --namespace default
```

This command displays all node specifications or secrets for the specified namespace. If no namespace is specified, the `default` namespace is used.

## HTTP API Integration

To modify node specifications through the HTTP API, you need to set up a workflow that exposes this functionality. You can use the `native` node included in the [basic extensions](../ext/README.md):

```yaml
kind: native
opcode: nodes.create # or nodes.read, nodes.update, nodes.delete
```

To get started, refer to the [workflow example](../examples/system.yaml). You may need to add authentication and authorization processes to this workflow as needed. Typically, such runtime control workflows are defined in the `system` namespace.
