# 💻 Command Line Interface (CLI)

Effectively manage the versatile Command Line Interface (CLI) designed for various workflow management tasks. This CLI is provided with the core executable file, including [built-in extensions](../ext/README.md).

## Configuration

Before executing commands, configure the system using environment variables. You can utilize either the `.uniflow.toml` file or system environment variables.

| TOML Key             | Environment Variable | Example                   |
|----------------------|----------------------|---------------------------|
| `database.url`       | `DATABASE.URL`       | `mem://` or `mongodb://`  |
| `database.name`      | `DATABASE.NAME`      | -                         |
| `collection.nodes`   | `COLLECTION.NODES`   | `nodes`                   |
| `collection.secrets` | `COLLECTION.SECRETS` | `secrets`                 |

When using [MongoDB](https://www.mongodb.com/), ensure that [change streams](https://www.mongodb.com/docs/manual/changeStreams/) are enabled so the engine can track changes to node specifications. To utilize change streams, set up a [replica set](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication).
