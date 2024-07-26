# 🚀 Getting Started

This comprehensive guide introduces how to manage workflows and run the engine using the [Command Line Interface (CLI)](../cmd/README.md). It covers everything from installation to workflow control and configuration settings.

## Installing from Source

First, let's set up the [CLI](../cmd/README.md) that controls workflows along with the [built-in extensions](../ext/README.md). Ensure that you have [Go 1.22](https://go.dev/doc/install) or higher installed on your system before starting.

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

Once the build process is complete, the executable file will be generated in the `dist` folder and will be ready for use.

## Configuration

Uniflow provides flexible configuration options via a `.uniflow.toml` file or system environment variables:

| TOML Key           | Environment Variable | Example                   |
|--------------------|----------------------|---------------------------|
| `database.url`     | `DATABASE.URL`       | `mem://` or `mongodb://`  |
| `database.name`    | `DATABASE.NAME`      | -                         |
| `collection.nodes` | `COLLECTION.NODES`   | `nodes`                   |

When using [MongoDB](https://www.mongodb.com/), you need to enable [change streams](https://www.mongodb.com/docs/manual/changeStreams/) to allow the engine to track node specification changes. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## CLI Commands

The CLI provides various commands for controlling workflows. To see all available commands, run:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command adds or updates node specifications in a namespace:

```sh
./dist/uniflow apply --filename examples/ping.yaml
```

This command outputs the result and uses the `default` namespace if none is specified.

### Delete

The `delete` command removes node specifications from a namespace:

```sh
./dist/uniflow delete --filename examples/ping.yaml
```

This command removes all node specifications defined in `examples/ping.yaml` from the specified namespace. If no namespace is specified, it uses the `default` namespace.

### Get

The `get` command retrieves node specifications from a namespace:

```sh
./dist/uniflow get
```

This command displays all node specifications in the specified namespace. If no namespace is specified, it uses the `default` namespace.

### Start

The `start` command initiates the runtime with node specifications from a specific namespace:

```sh
./dist/uniflow start
```

If the namespace is empty, you can provide initial node specifications using the `--filename` flag:

```sh
./dist/uniflow start --filename examples/ping.yaml
```

This command start all node specifications defined in `examples/ping.yaml` from the specified namespace. If no namespace is specified, it uses the `default` namespace.

## HTTP API Integration

To modify node specifications via HTTP API, you need to set up workflows that expose these functionalities. Use the `syscall` node included in the [basic extensions](../ext/README.md):

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

To get started, refer to the [workflow example](../examples/system.yaml). You can add authentication and authorization processes to this workflow as needed. Typically, such runtime control workflows are defined in the `system` namespace.
