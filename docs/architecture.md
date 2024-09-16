# 🏗️ Architecture

Each node specification declaratively defines the role of each node, and these specifications connect to form workflows. Each workflow is defined within a specific namespace, and each runtime environment executes a single namespace. Namespaces are isolated and cannot reference nodes from other namespaces.

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

The engine does not enforce specific nodes. Nodes connect to the engine through extensions and can be freely added or removed based on service requirements.

To optimize execution, two key processes—compilation and runtime—are employed. These processes reduce complexity and improve performance.

## Workflow Modification

The engine does not expose an API for directly modifying node specifications. Instead, it focuses on loading, compiling, and activating nodes to make them executable.

Users can update node specifications by using a Command-Line Interface (CLI) or defining workflows that provide an HTTP API for modifications. Typically, these workflows are defined in the `system` namespace.

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
      code: "has(self.body) ? self.body : null"
    - kind: native
      opcode: nodes.create
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 201
          };
        }

- kind: block
  name: nodes_read
  specs:
    - kind: snippet
      language: json
      code: "null"
    - kind: native
      opcode: nodes.read
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: nodes_update
  specs:
    - kind: snippet
      language: cel
      code: "has(self.body) ? self.body : null"
    - kind: native
      opcode: nodes.update
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: nodes_delete
  specs:
    - kind: snippet
      language: json
      code: "null"
    - kind: native
      opcode: nodes.delete
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            status: 204
          };
        }

- kind: block
  name: secrets_create
  specs:
    - kind: snippet
      language: cel
      code: "has(self.body) ? self.body : null"
    - kind: native
      opcode: secrets.create
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 201
          };
        }

- kind: block
  name: secrets_read
  specs:
    - kind: snippet
      language: json
      code: "null"
    - kind: native
      opcode: secrets.read
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: secrets_update
  specs:
    - kind: snippet
      language: cel
      code: "has(self.body) ? self.body : null"
    - kind: native
      opcode: secrets.update
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: secrets_delete
  specs:
    - kind: snippet
      language: json
      code: "null"
    - kind: native
      opcode: secrets.delete
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            status: 204
          };
        }

- kind: switch
  name: catch
  matches:
    - when: self == 'unsupported type' || self == 'unsupported value'
      port: out[0]
    - when: "true"
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

This approach ensures runtime stability while allowing flexible system expansion.

## Compilation Process

The loader tracks changes to node specifications and secrets in real-time through a change stream. When additions, modifications, or deletions occur, the loader dynamically reloads the specifications from the database. These specifications are compiled into executable forms using codecs defined in the schema, with caching and optimization to enhance performance.

Compiled nodes are transformed into symbols and stored in a symbol table. The symbol table connects each symbol's ports based on the port connection information in the node specifications.

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

Once all nodes in a workflow are loaded into the symbol table and ports are connected, load hooks are executed to activate nodes. If a node is removed, unload hooks deactivate dependent nodes.

Changes in node specifications propagate to all runtime environments.

## Runtime Process

Activated nodes monitor sockets or files and execute workflows. Each node spawns an independent process, isolating its execution flow from other nodes to optimize resource management.

Nodes open ports through these processes and create writers to send packets to connected nodes. The payload is converted into common types for transmission. Connected nodes create readers to process waiting packets and pass the results to the next node or return them.

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

Each reader processes packets sequentially and responds. Writers must return packets as responses to ensure smooth node communication and data consistency.

Nodes wait for responses to all sent packets before terminating the process and releasing allocated resources. If an error occurs, the node logs the error and terminates.

Upon process termination, the system releases all associated resources, including open file descriptors, memory, and database transactions. When a parent process terminates, all child processes terminate as well. The parent process typically waits until all child processes have completed.

This architecture ensures efficient node communication, data integrity, and stable execution across workflows.
