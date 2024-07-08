# Command Line Interface (CLI)

Effectively manage the Command Line Interface (CLI) with a versatile set of commands designed for seamless workflow management. This CLI serves as the default executable with [built-in extensions](../ext/README.md).

## Configuration Options

Before executing any commands, configure the system using environment variables. You can utilize `.uniflow.toml` or system environment variables.

| TOML Key           | Environment Variable Key | Example                  |
|--------------------|--------------------------|--------------------------|
| `database.url`     | `DATABASE.URL`           | `mem://` or `mongodb://` |
| `database.name`    | `DATABASE.NAME`          | -                        |
| `collection.nodes` | `COLLECTION.NODES`       | `nodes`                  |

If using [MongoDB](https://www.mongodb.com/), ensure that [change streams](https://www.mongodb.com/docs/manual/changeStreams/) are enabled to track modifications in node specifications. Utilize a [replica set](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication) for change streams.
