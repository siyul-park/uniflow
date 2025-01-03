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

| TOML Key            | Environment Variable Key | Example                  |
|---------------------|--------------------------|--------------------------|
| `database.url`      | `DATABASE_URL`           | `mem://` or `mongodb://` |
| `database.name`     | `DATABASE_NAME`          | -                        |
| `collection.charts` | `COLLECTION_CHARTS`      | `charts`                 |
| `collection.specs`  | `COLLECTION_SPECS`       | `nodes`                  |
| `collection.values` | `COLLECTION_VALUES`      | `values`                 |

If you are using [MongoDB](https://www.mongodb.com/), enable [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) to track resource changes in real time. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

## Running an Example

To run a basic HTTP request handler example using [ping.yaml](./examples/ping.yaml):

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: GET
      path: /ping
      port: out[0]
  ports:
    out[0]:
      - name: pong
        port: in

- kind: snippet
  name: pong
  language: text
  code: pong
```

To start the workflow, run:

```sh
uniflow start --from-specs example/ping.yaml
```

Verify it's running by calling the HTTP endpoint:

```sh
curl localhost:8000/ping
pong#
```

## Using Uniflow

`uniflow` is primarily used to start and manage the runtime environment.

### Start Command

The `start` command executes all node specifications in the specified namespace. If no namespace is provided, the default namespace (`default`) is used.

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide an initial node specification using the `--from-specs` flag:

```sh
./dist/uniflow start --namespace default --from-specs examples/specs.yaml
```

You can specify an initial values file with the `--from-values` flag:

```sh
./dist/uniflow start --namespace default --from-values examples/values.yaml
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
./dist/uniflowctl apply nodes --namespace default --filename examples/specs.yaml
```

To apply values:

```sh
./dist/uniflowctl apply values --namespace default --filename examples/values.yaml
```

To apply charts:

```sh
./dist/uniflowctl apply charts --namespace default --filename examples/charts.yaml
```

### Delete Command

The `delete` command removes all resources defined in the specified file. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/specs.yaml
```

To delete values:

```sh
./dist/uniflowctl delete values --namespace default --filename examples/values.yaml
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

To retrieve values:

```sh
./dist/uniflowctl get values --namespace default
```

To retrieve charts:

```sh
./dist/uniflowctl get charts --namespace default
```

## Integrating HTTP API

To modify node specifications via the HTTP API, set up workflows accordingly. You can use the `syscall` node provided in
the [core extensions](../ext/README.md):

```yaml
kind: syscall
opcode: specs.create # or specs.read, specs.update, specs.delete
```

Refer to the [workflow examples](../examples/system.yaml) to get started. If needed, you can add authentication and authorization processes. These runtime control workflows are typically defined in the `system` namespace.
