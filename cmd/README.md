# 💻 Command-Line Interface (CLI)

Effectively manage various workflows with the versatile Command-Line Interface (CLI) designed for system interaction.

### Setup

Environment configuration is managed through environment variables or the `.uniflow.toml` file. Below is how to register
and configure the default plugins:

```toml
[database]
url = "memory://"

[collection]
specs = "specs"
values = "values"

[language]
default = "cel"

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
path = "./dist/testing.so"
```

If you're using [MongoDB](https://www.mongodb.com/), you must
enable [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) to track real-time changes in resources. For
this, you will need to configure a [Replica Set](https://www.mongodb.com/docs/manual/replication/).

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

Environment variables can be specified with the `--env` flag.

```sh
./dist/uniflow start --namespace default --env DATABASE_URL=mongodb://localhost:27017 --env DATABASE_NAME=mydb
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

Environment variables can be specified with the `--env` flag.

```sh
./dist/uniflow test --namespace default --env DATABASE_URL=mongodb://localhost:27017 --env DATABASE_NAME=mydb
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
