# Architecture

Node specifications define the roles of each node declaratively, and these specifications are interconnected to form workflows. Each workflow is defined within a specific namespace, with each runtime enviroment executing a single namespace. Namespaces are isolated from each other, meaning nodes in one namespace cannot reference nodes in another namespace.

```plantext
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

The engine does not enforce the use of specific nodes. All nodes connect to the engine via extensions and can be added or removed as needed for your service.

To effectively execute node specifications, the process is divided into two main phases: compilation and runtime. This helps reduce complexity and optimize performance.

## Modifying Workflows

The engine does not expose an API to modify node specifications directly. Instead, it focuses on loading, compiling, and activating nodes for execution. 

To modify node specifications, use the Command-Line Interface (CLI) to update the specifications in the database. Alternatively, you can define a workflow that provides an HTTP API for modifying node specifications. These workflows are typically defined in the `system` namespace.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  links:
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
  links:
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
  match:
    - when: self == "invalid argument"
      port: out[0]
    - when: 'true'
      port: out[1]
  links:
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

This approach allows you to maintain a stable runtime environment while flexibly extending the system as needed.

## Compilation Process

The loader tracks changes to node specifications in real-time via database change streams. When specifications are added, modified, or deleted, the loader detects these changes and dynamically reloads them from the database. Using codecs defined in the scheme, these specifications are compiled into executable forms, with caching and optimization performed to enhance performance.

Compiled nodes are converted into symbols and stored in a symbol table, which connects the ports of each symbol based on the port connection information defined in the node specifications.

```plantext
   +--------------------------+   +-------------------+
   |         Database         |   |       Loader      |
   |  +--------------------+  |   |  +-------------+  |
   |  | Node Specification |  |-->|  |    Scheme   |  |
   |  +--------------------+  |   |  |  +-------+  |  |
   |  | Node Specification |  |   |  |  | Codec |  |  |--+
   |  +--------------------+  |   |  |  +-------+  |  |  |
   |  | Node Specification |  |   |  +-------------+  |  |
   |  +--------------------+  |   +-------------------+  |
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

When all nodes in a workflow are loaded into the symbol table and their ports are connected, the workflow activates nodes by running load hooks from internal nodes to external nodes sequentially. If a node becomes inoperative, all nodes that reference it are deactivated by running unload hooks in reverse order.

When node specifications in the database change, this process ensures that all runtime environments reflect the updates.

## Runtime Process

Activated nodes monitor sockets or files and execute workflows. Each node spawns an independent process to start execution, isolating the execution flow from other processes. This efficiently manages resources without impacting other tasks.

Nodes open ports through processes and create writers to send packets to connected nodes. The payload of these packets is converted to common types used during runtime and transmitted.

Connected nodes monitor if a new process has opened a port and create readers. These readers continuously process queued packets and forward the processed results to the next node or return them.

```plantext
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

In a reader, every packet must be processed and responded to sequentially, with packets sent by a writer returned as response packets. This setup ensures seamless node communication and upholds data consistency and integrity.

After executing the workflow, the node waits for responses to all sent packets. Upon receiving all responses, it terminates the process and releases allocated resources. If any errors occur during packet processing, the node logs these errors and subsequently terminates the process.

Once a process terminates, it checks and releases all resources associated with it, such as open file descriptors, allocated memory, and database transactions.

When a parent process terminates, all derived child processes also terminate. Typically, the parent process waits until all child processes have completed before exiting.
