# 🏗️ Architecture

Node specifications define the role of each node declaratively, and these specifications are connected to form workflows. Each workflow is defined within a specific namespace, and each runtime environment executes a single namespace. Namespaces are isolated and cannot reference nodes defined in other namespaces.

```plaintext
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

The engine does not enforce the use of specific nodes. All nodes are connected to the engine through extensions and can be freely added or removed to suit the service.

To effectively execute node specifications, two main processes are involved: compilation and runtime. These processes help reduce complexity and optimize performance.

## Workflow Modification

The engine does not expose an API for users to modify node specifications directly. Instead, it focuses on loading, compiling, and activating nodes to make them executable.

To modify node specifications, users can update the database using a Command-Line Interface (CLI) or define workflows that provide HTTP APIs for modifying node specifications. These workflows are typically defined in the `system` namespace.

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

- kind: block
  name: nodes_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: syscall
      opcode: nodes.create

- kind: block
  name: nodes_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: syscall
      opcode: nodes.read

- kind: block
  name: nodes_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: syscall
      opcode: nodes.update

- kind: block
  name: nodes_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: syscall
      opcode: nodes.delete

- kind: switch
  name: catch
  matches:
    - when: self == "unsupported type" || self == "unsupported value"
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
  code: >
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
  code: >
    {
      "body": {
        "error": "Internal Server Error"
      },
      "status": 500
    }
```

This approach ensures the runtime environment remains stable while allowing flexible system expansion as needed.

## Compilation Process

The loader tracks real-time changes to node specifications and secrets in the database through a change stream. When additions, modifications, or deletions occur, the loader detects these changes and dynamically reloads from the database. It then compiles the specifications into executable form using codecs defined in the schema. Caching and optimization are performed during this process to improve performance.

The compiled nodes are transformed into symbols and stored in a symbol table. The symbol table links each symbol’s ports based on the port connection information defined in the node specifications.

```plaintext
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

Once all nodes within a workflow are loaded into the symbol table and all ports are connected, load hooks are executed sequentially from internal nodes to external nodes to activate the nodes. However, if a specific node is removed from the symbol table and cannot be executed, unload hooks are executed in reverse order to deactivate all nodes referencing the removed node.

Any changes to node specifications in the database are reflected across all runtime environments through this process.

## Runtime Process

Activated nodes monitor sockets or files and execute the workflow. Nodes initiate independent processes for execution, isolating the execution flow from other processes. This ensures efficient resource management without affecting other tasks.

Each node opens ports through processes and creates writers to send packets to connected nodes. The payload of the packets is converted into a common type used at runtime.

Connected nodes monitor whether a new process has opened the respective port and create readers accordingly. The readers continuously process pending packets and pass the processed results to the next node or return them.

```plaintext
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

Each reader must sequentially process and respond to all packets, and packets sent by a writer must be returned as response packets. This ensures smooth communication between nodes and maintains data consistency and integrity.

Nodes executing the workflow wait until all response packets are returned before terminating the process and releasing allocated resources. If an error occurs during packet processing, resulting in an error response, the node logs the error and terminates the process.

Upon process termination, all open file descriptors, allocated memory, and database transactions are

 released after verifying normal termination.

When a parent process terminates, all derived child processes also terminate. The parent process typically waits until all child processes have terminated.