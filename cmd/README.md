# 💻 Command-Line Interface (CLI)

Effectively manage various workflows with the versatile Command-Line Interface (CLI) designed for system interaction.

### Configuration

You can configure the settings using environment variables or a configuration file (`.toml`, `.yaml`, `.json`, `.hjson`, `.env`). The path to the configuration file is specified using the `UNIFLOW_CONFIG` environment variable. If not specified, the default `.uniflow.toml` file will be used.

```bash
export UNIFLOW_CONFIG=./config/uniflow.toml
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
path = "./dist/ctrl.so"

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
