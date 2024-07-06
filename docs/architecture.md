# Architecture

Node specifications define the roles of each node declaratively, and these specifications are interconnected to form workflows. Each workflow is defined within a specific namespace, and each runtime executes a single namespace. Namespaces are isolated and cannot reference nodes defined in other namespaces.

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

To effectively execute node specifications, the system undergoes two main processes: compilation and runtime. These processes help reduce complexity and optimize performance.

## Modifying Workflows

The runtime does not directly expose APIs for modifying node specifications. Instead, it focuses on compiling, loading, and activating nodes to make them executable.

If node specifications need to be modified, the Command-Line Interface (CLI) can be used to update the specifications in the database. Alternatively, a workflow can be defined to provide an HTTP API for modifying node specifications. Such workflows are generally defined in the `system` namespace.

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
      - name: '400'
        port: in
    out[1]:
      - name: '500'
        port: in

- kind: snippet
  name: '400'
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
  name: '500'
  language: json
  code: >
    {
      "body": {
        "error": "Internal Server Error"
      },
      "status": 500
    }
```

This approach maintains the stability of the runtime environment while allowing flexible system extensions as needed.

## Compilation Process

During the compilation process, a loader tracks real-time changes to node specifications via the database’s change stream. The loader detects added, modified, or deleted specifications and dynamically reloads them from the database. Then, using codecs defined in the scheme, the specifications are compiled into executable forms. This process includes operations like optimization and caching to improve performance.

Compiled nodes are combined with the specifications and converted into symbols, which are stored in the symbol table. The symbol table connects each symbol’s ports based on the port connection information defined in the node specifications.

```plantext
   +--------------------------+             +-------------------+
   |         Database         |             |       Loader      |
   |  +--------------------+  |             |  +-------------+  |
   |  | Node Specification |  |------------>|  |    Scheme   |  |
   |  +--------------------+  |             |  |  +-------+  |  |
   |  | Node Specification |  |             |  |  | Codec |  |  |--+
   |  +--------------------+  |             |  |  +-------+  |  |  |
   |  | Node Specification |  |             |  +-------------+  |  |
   |  +--------------------+  |             +-------------------+  |
   +--------------------------+                                    |
   +-------------------------+                                     |
   |      Symbol Table       |                                     |
   |  +--------+ +--------+  |                                     |
   |  | Symbol | | Symbol |<---------------------------------------+
   |  +--------+ +--------+  |
   |           \|/           |
   |  +--------+ +--------+  |
   |  | Symbol | | Symbol |  |
   |  +--------+ +--------+  |
   +-------------------------+
```

After all nodes belonging to a workflow are loaded into the symbol table and ports are connected, load hooks are executed sequentially from internal nodes to external nodes to activate them. If a specific node in the symbol table is removed and becomes non-executable, all nodes referencing it are deactivated by executing unload hooks in reverse order.

When node specifications are changed in the database, this process ensures that changes are reflected across all runtime environments.

## Runtime Process

Activated nodes monitor sockets or files and execute workflows. Each node spawns an independent process to begin execution, isolating the execution flow and efficiently managing necessary resources to avoid impacting other tasks.

Each node opens ports through processes, creates writers, and sends packets to connected nodes. The payload of these packets is converted into common types used at runtime before transmission.

Connected nodes monitor whether a new process has opened the port, creating readers accordingly. These readers continuously process pending packets and deliver the processed results to the next node or return them.

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

All packets in a reader must be processed and responded to sequentially, and packets sent by a writer must be returned as response packets. This ensures smooth communication between nodes and maintains data consistency and integrity.

The node that executed the workflow waits for responses to all packets sent, then terminates the process and releases allocated resources. If an error occurs during packet processing, the node logs the error and terminates the process.

Once a process terminates, it checks and releases all resources associated with it, such as open file descriptors, allocated memory, and database transactions.

When a parent process terminates, all derived child processes also terminate. Typically, the parent process waits until all child processes have completed before exiting.
