# 🚀 Getting Started

This guide will walk you through installing, configuring, and managing workflows using the [Command Line Interface (CLI)](../cmd/README.md). It covers the full process, from installation to controlling and configuring workflows.

## Installing from Source

First, set up the [CLI](../cmd/README.md) along with the [core extensions](../ext/README.md). Make sure your system has [Go 1.23](https://go.dev/doc/install) or a later version installed.

### Clone the Repository

To download the source code, run the following command in your terminal:

```sh
git clone https://github.com/siyul-park/uniflow
```

Navigate to the downloaded folder:

```sh
cd uniflow
```

### Install Dependencies and Build

To install the required dependencies and build the project, run the following commands:

```sh
make init
make build
```

After the build completes, the executable files will be available in the `dist` folder.

### Configuration

Settings can be modified using the `.uniflow.toml` file or system environment variables. The key configuration options are:

| TOML Key             | Environment Variable Key | Example                     |
|----------------------|--------------------------|-----------------------------|
| `database.url`       | `DATABASE.URL`           | `mem://` or `mongodb://`    |
| `database.name`      | `DATABASE.NAME`          | -                           |
| `collection.charts`  | `COLLECTION.CHARTS`      | `charts`                    |
| `collection.nodes`   | `COLLECTION.NODES`       | `nodes`                     |
| `collection.secrets` | `COLLECTION.SECRETS`     | `secrets`                   |

If you are using [MongoDB](https://www.mongodb.com/), enable [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) to track resource changes in real time. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## Using Uniflow

`uniflow` is primarily used to start and manage the runtime environment.

### Start Command

The `start` command executes all node specifications in the specified namespace. If no namespace is provided, the default namespace (`default`) is used.

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide an initial node specification using the `--from-specs` flag:

```sh
./dist/uniflow start --namespace default --from-specs examples/nodes.yaml
```

You can specify an initial secrets file with the `--from-secrets` flag:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

Charts can be initialized using the `--from-charts` flag:

```sh
./dist/uniflow start --namespace default --from-charts examples/charts.yaml
```

## Using Uniflowctl

`uniflowctl` is a command used to manage resources within a namespace.

### Apply Command

The `apply` command applies the contents of a specified file to the namespace. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/nodes.yaml
```

To apply secrets:

```sh
./dist/uniflowctl apply secrets --namespace default --filename examples/secrets.yaml
```

To apply charts:

```sh
./dist/uniflowctl apply charts --namespace default --filename examples/charts.yaml
```

### Delete Command

The `delete` command removes all resources defined in the specified file. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/nodes.yaml
```

To delete secrets:

```sh
./dist/uniflowctl delete secrets --namespace default --filename examples/secrets.yaml
```

To delete charts:

```sh
./dist/uniflowctl delete charts --namespace default --filename examples/charts.yaml
```

### Get Command

The `get` command retrieves all resources in the specified namespace. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflowctl get nodes --namespace default
```

To retrieve secrets:

```sh
./dist/uniflowctl get secrets --namespace default
```

To retrieve charts:

```sh
./dist/uniflowctl get charts --namespace default
```

## Integrating HTTP API

To modify node specifications via the HTTP API, set up workflows accordingly. You can use the `native` node provided in the [core extensions](../ext/README.md):

```yaml
kind: native
opcode: nodes.create # or nodes.read, nodes.update, nodes.delete
```

Refer to the [workflow examples](../examples/system.yaml) to get started. If needed, you can add authentication and authorization processes. These runtime control workflows are typically defined in the `system` namespace.
