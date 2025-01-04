# 💻 Command Line Interface (CLI)

Effectively manage your workflows using the versatile Command Line Interface (CLI) designed for a variety of tasks. This CLI is provided as a base executable that includes [built-in extensions](../ext/README.md).

### Configuration

Settings can be modified using the `.uniflow.toml` file or system environment variables. The key configuration options are:

| TOML Key            | Environment Variable Key | Example                  |
|---------------------|--------------------------|--------------------------|
| `database.url`      | `DATABASE_URL`           | `mem://` or `mongodb://` |
| `database.name`     | `DATABASE_NAME`          | -                        |
| `collection.specs`  | `COLLECTION_SPECS`       | `specs`                  |
| `collection.values` | `COLLECTION_VALUES`      | `values`                 |

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
./dist/uniflow start --namespace default --from-specs examples/specs.yaml
```

You can specify an initial values file with the `--from-values` flag:

```sh
./dist/uniflow start --namespace default --from-values examples/values.yaml
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

### Delete Command

The `delete` command removes all resources defined in the specified file. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/specs.yaml
```

To delete values:

```sh
./dist/uniflowctl delete values --namespace default --filename examples/values.yaml
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
