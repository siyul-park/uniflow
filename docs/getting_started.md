# 🚀 Getting Started

This comprehensive walkthrough will introduce you to managing workflows and running the engine using our [Command-Line Interface (CLI)](../cmd/README.md). We'll cover everything from installation to workflow control and configuration setup.

## Installing from Source

Let's begin by setting up the [CLI](../cmd/README.md), which allows you to control workflows with [built-in extensions](../ext/README.md). Before we start, ensure you have [Go 1.22](https://go.dev/doc/install) or later installed on your system.

First, clone the repository:

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

Once the build process completes, you'll find the executable in the `dist` folder, ready for use.

## Configuration

Uniflow offers flexible configuration options through either a `.uniflow.toml` file or system environment variables:

| TOML Key           | Environment Variable | Example                  |
|--------------------|----------------------|--------------------------|
| `database.url`     | `DATABASE.URL`       | `mem://` or `mongodb://` |
| `database.name`    | `DATABASE.NAME`      | -                        |
| `collection.nodes` | `COLLECTION.NODES`   | `nodes`                  |

When using [MongoDB](https://www.mongodb.com/), enable [change streams](https://www.mongodb.com/docs/manual/changeStreams/) to allow the engine to track node specification changes. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## CLI Commands

Uniflow's CLI provides a suite of commands for workflow control. To see all available commands, use:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command adds or updates node specifications in a namespace:

```sh
./dist/uniflow apply --filename examples/ping.yaml
```

This command outputs the results and uses the `default` namespace if none is specified.

### Delete

Remove node specifications from a namespace with the `delete` command:

```sh
./dist/uniflow delete --filename examples/ping.yaml
```

This removes all node specifications defined in `examples/ping.yaml` from the specified (or default) namespace.

### Get

Retrieve node specifications from a namespace using the `get` command:

```sh
./dist/uniflow get
```

This displays all node specifications in the default (or specified) namespace.

### Start

Launch the runtime with node specifications from a specific namespace using the `start` command:

```sh
./dist/uniflow start
```

If the namespace is empty, you can provide initial node specifications using the `--filename` flag:

```sh
./dist/uniflow start --filename examples/ping.yaml
```

## HTTP API Integration

To modify node specifications via HTTP API, you'll need to set up a workflow that exposes these capabilities. Utilize the `syscall` node included in the default extensions:

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

Explore our [workflow example](../examples/system.yaml) to get started. You can enhance this workflow with authentication and authorization processes as needed. Typically, these runtime control workflows are defined in the `system` namespace.
