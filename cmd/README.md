# 💻 Command Line Interface (CLI)

Effectively manage your workflows using the versatile Command Line Interface (CLI) designed for a variety of tasks. This CLI is provided as a base executable that includes [built-in extensions](../ext/README.md).

## Configuration

Before running commands, configure your system using environment variables. You can use either the `.uniflow.toml` file or system environment variables.

| TOML Key              | Environment Variable Key  | Example                    |
|-----------------------|----------------------------|----------------------------|
| `database.url`        | `DATABASE.URL`             | `mem://` or `mongodb://`   |
| `database.name`       | `DATABASE.NAME`            | -                          |
| `collection.charts`   | `COLLECTION.CHARTS`        | `charts`                   |
| `collection.nodes`    | `COLLECTION.NODES`         | `nodes`                    |
| `collection.secrets`  | `COLLECTION.SECRETS`       | `secrets`                  |

If using [MongoDB](https://www.mongodb.com/), ensure that [Change Streams](https://www.mongodb.com/docs/manual/changeStreams/) are enabled so that the engine can track changes to resources. To utilize Change Streams, set up a [Replica Set](https://www.mongodb.com/docs/manual/replication/#std-label-replication).
