# 🏗️ Architecture

Based on nodes as the minimal unit of processing work, node specifications define the role each node will perform, and these nodes connect to each other to form workflows. Each workflow operates within a predefined namespace in a single runtime, and each runtime environment executes one namespace.

> 💡 For detailed information on the initialization and execution process of workflows, please refer to the [Runtime Documentation](./runtime.md).
> 
> 💡 For detailed explanations of core terms and concepts used in the system, please refer to the [Key Concepts Documentation](./key_concepts.md).

Namespaces are isolated and managed separately, and cannot arbitrarily reference nodes defined in other namespaces.

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

The engine does not enforce specific nodes, and all nodes can be freely added or removed according to service requirements.

### Workflow Modification

The engine does not provide an API for users to change node specifications, focusing instead on loading, compiling, and executing nodes. When node specifications need to be modified, they can be updated through a Command-Line Interface (CLI) or an HTTP API using a directly defined [workflow](../examples/system.yaml). Such workflows are typically defined in the `system` namespace.

This approach allows for flexible system expansion while maintaining a stable runtime environment.

### Compilation Process

The loader tracks changes to node specifications and variables in real-time through the database's change stream. When additions, modifications, or deletions occur, the loader reloads the specifications and compiles them into executable forms using codecs defined in the scheme. Caching and optimization processes are also performed to improve performance.

Compiled nodes are combined with their specifications to form symbols, which are then stored in the symbol table. The symbol table connects each symbol's ports based on the port connection information defined in the node specifications.

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
   |  | Value  |  | Value  |  |-->|  +-------------+  |  |
   |  +--------+  +--------+  |   +-------------------+  |
   |  +--------+  +--------+  |                          |
   |  | Value  |  | Value  |  |                          |
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

Compiled nodes are stored in the symbol table, and symbols are connected according to their defined ports. Once all nodes in a workflow are loaded into the symbol table, sequential operations to activate the nodes are executed. When nodes are removed, deactivation operations are also executed sequentially.

### Runtime Process

Activated nodes execute workflows, managing resources through independent processes to avoid affecting other operations. Each node exchanges packets through inter-process communication, and payloads are converted to common types for transmission.

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

A single reader processes all packets sequentially and must return response packets for packets sent to the writer. This ensures smooth communication between nodes and guarantees data consistency and integrity.

The node that executes a workflow waits until it receives responses for all sent packets, then terminates the process and releases allocated resources. If an error occurs during packet processing and an error response is returned, the node logs the error and terminates the process.

When a process terminates, it checks for normal termination and releases open file descriptors, allocated memory, database transactions, and other resources.

When a parent process terminates, all child processes derived from it are also terminated. The parent process waits until all child processes have terminated.

This architecture ensures efficient node communication, data integrity, and stable execution across workflows.
