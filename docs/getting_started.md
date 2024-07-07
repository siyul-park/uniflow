# Getting Started

This guide explains how to manage workflows and run the engine using the Command-Line Interface (CLI). Follow the steps from installing the CLI to controlling workflows and configuring settings.

## Installation from Code

Install the CLI to control workflows, along with [built-in extensions](../ext/README.md). To build the code, you need [Go 1.22](https://go.dev/doc/install) or later.

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

After the build is complete, the executable will be available in the `dist` folder. You can use this to run the CLI.

## Configuration

You can configure the environment using the `.uniflow.toml` file or system environment variables.

| TOML Key         | Environment Variable Key | Example                       |
|------------------|--------------------------|-------------------------------|
| `database.url`   | `DATABASE.URL`           | `mem://` or `mongodb://`      |
| `database.name`  | `DATABASE.NAME`          | -                             |

If you use [MongoDB](https://www.mongodb.com/), ensure that [change streams](https://www.mongodb.com/docs/manual/changeStreams/) are enabled so the runtime engine can track changes to node specifications. Use a [replica set](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication) to utilize change streams.

## Commands

The CLI offers various commands to control workflows. For a full list of commands, refer to the help command:

```sh
./dist/uniflow --help
```

### Apply

The `apply` command applies node specifications to a specified namespace. If the node specification already exists in the namespace, it updates the existing one; otherwise, it creates a new one. It outputs the result of the applied specifications. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflow apply --filename examples/ping.yaml
 ID                                    KIND      NAMESPACE  NAME      LINKS                                
 01908c74-8b22-7cbf-a475-6b6bc871b01a  listener  <nil>      listener  map[out:[map[name:router port:in]]]  
 01908c74-8b22-7cc0-ae2b-40504e7c9ff0  router    <nil>      router    map[out[0]:[map[name:pong port:in]]] 
 01908c74-8b22-7cc1-ac48-83b5084a0061  snippet   <nil>      pong      <nil>                                
```

### Delete

The `delete` command removes node specifications from a namespace. This is useful for deleting nodes of a specific workflow. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflow delete --filename examples/ping.yaml
```

This command deletes all node specifications defined in `examples/ping.yaml` from the specified namespace.

### Get

The `get` command retrieves node specifications from a namespace. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflow get
 ID                                    KIND      NAMESPACE  NAME      LINKS                                
 01908c74-8b22-7cbf-a475-6b6bc871b01a  listener  <nil>      listener  map[out:[map[name:router port:in]]]  
 01908c74-8b22-7cc0-ae2b-40504e7c9ff0  router    <nil>      router    map[out[0]:[map[name:pong port:in]]] 
 01908c74-8b22-7cc1-ac48-83b5084a0061  snippet   <nil>      pong      <nil>                                
```

### Start

The `start` command loads node specifications from a specified namespace and runs the runtime. If no namespace is specified, the `default` namespace is used.

```sh
./dist/uniflow start                  
```

If the namespace is empty and there are no nodes to run, you can use the `--filename` flag to provide default node specifications for the namespace.

```sh
./dist/uniflow start --filename examples/ping.yaml
```

## HTTP API

To modify node specifications through the HTTP API, you need to expose the HTTP API via a workflow installed in a namespace using the CLI. Utilize the `syscall` node included in the default extensions for this purpose.

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

Refer to the [workflow example](../examples/crud.yaml) to get started. You can include authentication and authorization processes in this workflow as needed. Typically, workflows related to runtime control are defined in the `system` namespace.
