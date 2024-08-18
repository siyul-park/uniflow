# 🏗️ Architecture

Each node specification declaratively defines the role of each node, and these specifications connect to form workflows. Each workflow is defined within a specific namespace, and each runtime environment executes a single namespace. Namespaces are isolated and cannot reference nodes defined in other namespaces.

```text
   +-------------------------------------------------+
   |                   Workflow A                    |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   |                       \|/                       |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   +-------------------------------------------------+
   |                   Workflow B                    |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   |                       \|/                       |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   +-------------------------------------------------+
```

The engine does not enforce the use of specific nodes. All nodes connect to the engine through extensions and can be freely added or removed according to the service needs.

To effectively execute node specifications, two main processes—compilation and runtime—are employed. These processes help reduce complexity and optimize performance.

## Workflow Modification

The engine does not expose an API to users for modifying node specifications. Instead, the engine focuses solely on loading, compiling, and activating nodes to make them executable.

To modify node specifications, users can update the database using a Command-Line Interface (CLI) or define workflows that provide an HTTP API to modify node specifications. Typically, such workflows are defined in the `system` namespace.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in
    error:
      - name: catch
        port: in

- kind: router
  name: router
  routes:
    - method: POST
      path: /v1/nodes
      port: out[0]
    - method: GET
      path: /v1/nodes
      port: out[1]
    - method: PATCH
      path: /v1/nodes
      port: out[2]
    - method: DELETE
      path: /v1/nodes
      port: out[3]
    - method: POST
      path: /v1/secrets
      port: out[4]
    - method: GET
      path: /v1/secrets
      port: out[5]
    - method: PATCH
      path: /v1/secrets
      port: out[6]
    - method: DELETE
      path: /v1/secrets
      port: out[7]
  ports:
    out[0]:
      - name: nodes_create
        port: in
    out[1]:
      - name: nodes_read
        port: in
    out[2]:
      - name: nodes_update
        port: in
    out[3]:
      - name: nodes_delete
        port: in
    out[4]:
      - name: secrets_create
        port: in
    out[5]:
      - name: secrets_read
        port: in
    out[6]:
      - name: secrets_update
        port: in
    out[7]:
      - name: secrets_delete
        port: in

- kind: block
  name: nodes_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.create

- kind: block
  name: nodes_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.read

- kind: block
  name: nodes_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.update

- kind: block
  name: nodes_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.delete

- kind: block
  name: secrets_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: secrets.create

- kind: block
  name: secrets_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: secrets.read

- kind: block
  name: secrets_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: secrets.update

- kind: block
  name: secrets_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: secrets.delete

- kind: switch
  name: catch
  matches:
    - when: self == 'unsupported type' || self == 'unsupported value'
      port: out[0]
    - when: 'true'
      port: out[1]
  ports:
    out[0]:
      - name: status_400
        port: in
    out[1]:
      - name: status_500
        port: in

- kind: snippet
  name: status_400
  language: javascript
  code: |
    export default function (args) {
      return {
        body: {
          error: args.error()
        },
        status: 400
      };
    }

- kind: snippet
  name: status_500
  language: json
  code: |
    {
      "body": {
        "error": "Internal Server Error"
      },
      "status": 500
    }
```

This approach ensures the runtime environment remains stable while allowing flexible system expansion as needed.

## Compilation Process

The loader tracks real-time changes to node specifications and secrets in the database through a change stream. When additions, modifications, or deletions occur, the loader detects these changes and dynamically reloads them from the database. The specifications are then compiled into executable forms using codecs defined in the schema. This process involves caching and optimization to enhance performance.

The compiled nodes are transformed into symbols and stored in a symbol table. The symbol table connects each symbol's ports based on the port connection information defined in the node specifications.

```text
   +--------------------------+
   |         Database         |
   |  +--------------------+  |
   |  | Node Specification |  |
   |  +--------------------+  |
   |  | Node Specification |  |
   |  +--------------------+  |   +-------------------+
   |  | Node Specification |  |-->|       Loader      |
   |  +--------------------+  |   |  +-------------+  |
   +--------------------------+   |  |    Scheme   |  |
   +--------------------------+   |  |  +-------+  |  |
   |         Database         |   |  |  | Codec |  |  |--+
   |  +--------+  +--------+  |   |  |  +-------+  |  |  |
   |  | Secret |  | Secret |  |-->|  +-------------+  |  |
   |  +--------+  +--------+  |   +-------------------+  |
   |  +--------+  +--------+  |                          |
   |  | Secret |  | Secret |  |                          |
   |  +--------+  +--------+  |                          |
   +--------------------------+                          |
   +-------------------------+                           |
   |      Symbol Table       |                           |
   |  +--------+ +--------+  |                           |
   |  | Symbol | | Symbol |<-----------------------------+
   |  +--------+ +--------+  |
   |           \|/           |
   |  +--------+ +--------+  |
   |  | Symbol | | Symbol |  |
   |  +--------+ +--------+  |
   +-------------------------+
```

Once all nodes within a workflow are loaded into the symbol table and all ports are connected, the workflow executes load hooks sequentially from internal nodes to external nodes to activate them. Conversely, if a specific node is removed from the symbol table and becomes non-executable, unload hooks are executed in reverse order to deactivate all nodes that reference the affected node.

Any changes to node specifications in the database are propagated to all runtime environments through this process.

## Runtime Process

Activated nodes monitor sockets or files and execute workflows. Each node starts by spawning an independent process, isolating its execution flow from other processes. This approach ensures efficient resource management and minimizes impact on other tasks.

Each node opens ports through processes and creates writers to send packets to connected nodes. The payload of these packets is converted into common types used in the runtime for transmission.

Connected nodes monitor whether a new process has opened the port and create readers accordingly. These readers continuously process waiting packets and pass the processed results to the next node or return them.

```text
   +-----------------------+          +-----------------------+
   |        Node A         |          |        Node B         |
   |  +-----------------+  |          |  +-----------------+  |
   |  |    Port Out     |--------------->|    Port In      |  |
   |  |                 |  |          |  |                 |  |
   |  |  +-----------+  |  |  packet  |  |  +-----------+  |  |
   |  |  | Process 1 |======================| Process 1 |  |  |
   |  |  +-----------+  |  |          |  |  +-----------+  |  |
   |  |  +-----------+  |  |  packet  |  |  +-----------+  |  |
   |  |  | Process 2 |======================| Process 2 |  |  |
   |  |  +-----------+  |  |          |  |  +-----------+  |  |
   |  +-----------------+  |          |  +-----------------+  |
   +-----------------------+          +-----------------------+
```

A single reader processes and responds to all packets sequentially, and packets sent by writers must be returned as response packets. This mechanism ensures smooth communication between nodes and maintains data consistency and integrity.

Nodes that execute workflows wait until responses for all sent packets are returned before terminating the process and releasing allocated resources. If an error occurs during packet processing, resulting in an error response, the node logs the error and terminates the process.

Upon process termination, the system verifies normal termination and releases all associated resources, including open file descriptors, allocated memory, and database transactions.

When a parent process terminates, all derived child processes also terminate. Typically, the parent process waits until all child processes have terminated.
