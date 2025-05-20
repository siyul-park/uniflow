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
make build-all
```

After the build completes, the executable files will be available in the `dist` folder.

### Configuration

You can configure the settings using environment variables or a configuration file (`.toml`, `.yaml`, `.json`, `.hjson`, `.env`). The path to the configuration file is specified using the `UNIFLOW_CONFIG` environment variable. If not specified, the default `.uniflow.toml` file will be used.

```bash
export UNIFLOW_CONFIG=./uniflow.toml
```

The configuration file can define the following key settings:

```toml
[runtime]
namespace = "default"
language = "cel"

[database]
url = "memory://"

[collection]
specs = "specs"
values = "values"

[[plugins]]
path = "./dist/cel.so"
config.extensions = ["encoders", "math", "lists", "sets", "strings"]

[[plugins]]
path = "./dist/ecmascript.so"

[[plugins]]
path = "./dist/mongodb.so"

[[plugins]]
path = "./dist/reflect.so"

[[plugins]]
path = "./dist/ctl.so"

[[plugins]]
path = "./dist/net.so"

[[plugins]]
path = "./dist/sql.so"

[[plugins]]
path = "./dist/testing.so"
```

Environment variables are also automatically loaded, and they use the `UNIFLOW_` prefix. For example, the following environment variables can be set:

```env
UNIFLOW_DATABASE_URL=memory://
UNIFLOW_COLLECTION_SPECS=specs
UNIFLOW_COLLECTION_VALUES=values
UNIFLOW_LANGUAGE_DEFAULT=cel
```

If you are using [MongoDB](https://www.mongodb.com/), you will need to enable [change streams](https://www.mongodb.com/docs/manual/changeStreams/) to track resource changes in real-time. This requires setting up a [replica set](https://www.mongodb.com/docs/manual/replication/).

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

## Supported Commands

`uniflow` provides various commands used to start and manage the runtime environment of workflows.

### Start Command

The `start` command runs all node specifications within the specified namespace. If no namespace is specified, the
default `default` namespace will be used.

```sh
./dist/uniflow start --namespace default
```

If the namespace is empty, you can provide the initial node specification with the `--from-specs` flag.

```sh
./dist/uniflow start --namespace default --from-specs examples/specs.yaml
```

Initial variable files can be set using the `--from-values` flag.

```sh
./dist/uniflow start --namespace default --from-values examples/values.yaml
```

Environment variables can be specified with the `--environment` flag.

```sh
./dist/uniflow start --namespace default --environment DATABASE_URL=mongodb://localhost:27017 --environment DATABASE_NAME=mydb
```

### Test Command

The `test` command runs workflow tests within the specified namespace. If no namespace is specified, the default
`default` namespace will be used.

```sh
./dist/uniflow test --namespace default
```

To run specific tests, use regular expressions for filtering.

```sh
./dist/uniflow test ".*/my_test" --namespace default
```

If the namespace is empty, you can apply initial specifications and variables.

```sh
./dist/uniflow test --namespace default --from-specs examples/specs.yaml --from-values examples/values.yaml
```

Environment variables can be specified with the `--environment` flag.

```sh
./dist/uniflow test --namespace default --environment DATABASE_URL=mongodb://localhost:27017 --environment DATABASE_NAME=mydb
```

### Apply Command

The `apply` command applies the content of the specified file to the namespace. If no namespace is specified, the
default `default` namespace will be used.

```sh
./dist/uniflow apply nodes --namespace default --filename examples/specs.yaml
```

To apply variables:

```sh
./dist/uniflow apply values --namespace default --filename examples/values.yaml
```

### Delete Command

The `delete` command removes all resources defined in the specified file. If no namespace is specified, the default
`default` namespace will be used.

```sh
./dist/uniflow delete nodes --namespace default --filename examples/specs.yaml
```

To delete variables:

```sh
./dist/uniflow delete values --namespace default --filename examples/values.yaml
```

### Get Command

The `get` command retrieves all resources within the specified namespace. If no namespace is specified, the default
`default` namespace will be used.

```sh
./dist/uniflow get nodes --namespace default
```

To retrieve variables:

```sh
./dist/uniflow get values --namespace default
```
