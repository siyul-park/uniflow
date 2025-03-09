# 📚 Key Concepts

This guide provides a detailed explanation of the key terms and concepts used in the system.

## Summary of Key Terms

| Term | Description |
|------|-------------|
| **Specification (Spec)** | Information that contains the definition of a node. It is expressed in JSON/YAML format and includes ID, namespace, name, kind, port information, etc. |
| **Node** | The basic execution unit of a workflow. It receives input, processes it, and generates output. |
| **Symbol** | An object that wraps a node to manage it in the runtime. It contains both the node instance and specification information. |
| **Port** | Connection point between nodes. It is divided into input port (InPort) and output port (OutPort). |
| **Packet** | The basic unit of data transferred between nodes. There are request packets and response packets. |
| **Process** | Independent unit of workflow execution. It has a unique ID and state, and provides context for packet processing. |
| **Runtime** | Environment that manages and executes workflows. |
| **Loader** | Loads specifications, converts them to nodes, and registers them in the symbol table. |
| **Symbol Table** | Manages symbols by ID and name, and establishes connections between symbols. |
| **Scheme** | Registers codecs for each node type and provides rules for converting specifications to nodes. |
| **Hook** | Handles symbol lifecycle and port activation events. Examples include LoadHook, UnloadHook, OpenHook, etc. |
| **Namespace** | Isolates and manages workflows, providing an independent execution environment. |
| **Variable** | Securely stores sensitive information required by nodes, such as passwords and API keys. |

## Node Specification

The node specification declaratively defines how each node operates and connects. The engine compiles this specification into an executable node.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01a
kind: listener
namespace: default
name: example-listener
annotations:
  description: "Example listener node using HTTP protocol"
  version: "1.0"
protocol: http
port: "{{ .PORT }}"
env:
  PORT:
    name: network
    data: "{{ .PORT }}"
ports:
  out:
    - name: proxy
      port: in
```

- `id`: Unique identifier in UUID format. UUID V7 is recommended.
- `kind`: Specifies the type of the node. Additional fields may vary depending on the node type.
- `namespace`: Specifies the namespace to which the node belongs; the default value is `default`.
- `name`: Specifies the name of the node, which must be unique within the same namespace.
- `annotations`: Additional metadata for the node. It can include user-defined key-value pairs such as description, version, etc.
- `protocol`: Specifies the protocol used by the listener. This is a required field for nodes of the `listener` kind.
- `port`: Specifies the port used by the listener. This is a required field for nodes of the `listener` kind.
- `ports`: Defines how ports are connected. `out` defines an output port named `proxy` that connects to the `in` port of another node.
- `env`: Specifies environment variables needed by the node. Here, `PORT` is dynamically set from a variable.

## Variables

Variables securely store sensitive information needed by nodes, such as passwords and API keys.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: database
annotations:
  description: "Database information"
data:
  password: "super-value-password"
```

- `id`: Unique identifier in UUID format. UUID V7 is recommended.
- `namespace`: Specifies the namespace to which the variable belongs; the default value is `default`.
- `name`: Specifies the name of the variable, which must be unique within the same namespace.
- `annotations`: Additional metadata for the variable. It can include user-defined key-value pairs such as description, version, etc.
- `data`: Contains variable data consisting of key-value pairs.

## Nodes

Nodes are objects that process data, exchanging packets through interconnected ports to execute workflows. Each node has an independent processing loop and communicates asynchronously with other nodes.

Nodes are classified according to their packet processing method:
- `ZeroToOne`: Nodes that generate initial packets to start the workflow.
- `OneToOne`: Nodes that receive packets from the input port, process them, and transmit them to the output port.
- `OneToMany`: Nodes that receive packets from the input port and transmit them to multiple output ports.
- `ManyToOne`: Nodes that receive packets from multiple input ports and transmit them to one output port.
- `Other`: Nodes that include state management and interactions beyond simple packet transfer.

## Ports

Ports are connection points for exchanging packets between nodes. There are two types of ports, `InPort` and `OutPort`, which are connected to transmit packets. Packets sent to a port are forwarded to all connected ports.

Commonly used port names include:
- **`init`**: A special port used to initialize a node. When a node is activated, the workflow connected to the `init` port is executed.
- **`term`**: A special port used to terminate a node. When a node is deactivated, the workflow connected to the `term` port is executed.
- **`io`**: Processes packets and returns them immediately.
- **`in`**: Receives and processes packets, and sends the processing results to `out` or `error`. If there is no connected `out` or `error` port, the result is returned immediately.
- **`out`**: Transmits processed packets. The transmitted results can be forwarded to another `in` port.
- **`error`**: Transmits errors that occur during packet processing. Error handling results can be forwarded back to the `in` port.

When multiple ports with the same role are needed, they are denoted as `in[0]`, `in[1]`, `out[0]`, `out[1]`, etc.

## Packets

A packet is a unit of data exchanged between ports. Each packet contains a payload, which nodes process and transmit.

Nodes must return response packets in the order of request packets. When connected to multiple ports, all response packets are collected and returned as a single new response packet.

The special `None` packet indicates that there is no response, simply indicating that the packet has been accepted.

## Processes

A process is the basic unit of execution and is managed independently. Processes can have parent processes, and when a parent process terminates, child processes are also terminated.

Processes have their own storage for storing values that are difficult to transmit as packets. This storage operates in Copy-On-Write (COW) mode, efficiently sharing data from the parent process.

A new workflow starts by creating a process. When a process terminates, all resources used are released.

A process can have two or more root packets, but root packets must be generated from the same node. If generated from different nodes, a new child process must be created for processing.

## Workflow

A workflow is defined as a directed graph, a structure where multiple nodes are connected. In this graph, each node is responsible for data processing, and packets are transmitted between nodes.

Workflows consist of multiple stages, and at each stage, data is processed and transferred according to defined rules. In this process, data can be processed sequentially or in parallel.

For example, when initial data is given, it is processed by the first node and then passed to the next node. Each node receives input, processes it, and sends the processed result to the next stage.

## Namespace

A namespace isolates and manages workflows, providing an independent execution environment. Each namespace can contain multiple workflows, and nodes within a namespace cannot reference nodes belonging to other namespaces. Each namespace independently manages its own data and resources.

## Runtime Environment

The runtime environment is an independent space where each namespace runs. The engine loads all nodes belonging to a namespace to build the environment and execute workflows. This prevents conflicts that may occur during workflow execution and provides a stable execution environment.
