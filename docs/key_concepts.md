# 📚 Key Concepts

This guide provides detailed explanations of the key terms and concepts used in the system.

## Node Specification

A node specification declaratively defines the behavior and connections of each node. The engine compiles this specification into executable nodes.

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
ports:
  out:
    - name: proxy
      port: in
env:
  PORT:
    - name: network
      value: "{{ .PORT }}"
```

- `id`: A unique identifier in UUID format. UUID V7 is recommended.
- `kind`: Specifies the type of node. This example is a `listener`. Additional fields may vary based on the node type.
- `namespace`: The namespace to which the node belongs, default is `default`.
- `name`: The name of the node, which must be unique within the namespace.
- `annotations`: Additional metadata about the node. It can include user-defined key-value pairs like description and version.
- `protocol`: Specifies the protocol used by the listener. This field is required for `listener` nodes.
- `port`: Specifies the port used by the listener. This field is required for `listener` nodes.
- `ports`: Defines the connection scheme of ports. `out` defines an output port named `proxy`, which connects to the `in` port of another node.
- `env`: Specifies environment variables needed by the node. Here, `PORT` is dynamically set from a secret.

## Secret

A secret securely stores sensitive information needed by nodes, such as passwords and API keys.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: database
annotations:
  description: "Database information"
data:
  password: "super-secret-password"
```

- `id`: A unique identifier in UUID format. UUID V7 is recommended.
- `namespace`: The namespace to which the secret belongs, default is `default`.
- `name`: The name of the secret, which must be unique within the namespace.
- `annotations`: Additional metadata about the secret. It can include user-defined key-value pairs like description and version.
- `data`: Contains the secret data in key-value pairs.

## Node

A node is an entity that processes data, exchanging packets through connected ports to execute workflows. Each node has an independent processing loop and communicates asynchronously with other nodes.

Nodes are classified based on their packet processing methods:
- `ZeroToOne`: Generates initial packets to start a workflow.
- `OneToOne`: Receives a packet from an input port, processes it, and sends it to an output port.
- `OneToMany`: Receives a packet from an input port and sends it to multiple output ports.
- `ManyToOne`: Receives packets from multiple input ports and sends a single packet to an output port.
- `Other`: Includes nodes that manage state and interactions beyond simple packet forwarding.

## Port

Ports are connection points for exchanging packets between nodes. There are two types of ports: `InPort` and `OutPort`. Packets sent to a port are delivered to all connected ports.

Common port names include:
- `init`: A special port used to initialize a node. When the node becomes available, workflows connected to the `init` port are executed.
- `io`: Processes and immediately returns packets.
- `in`: Receives packets for processing and sends results to `out` or `error`. If `out` or `error` ports are not connected, the result is returned directly.
- `out`: Sends processed packets. The results can be sent back to another `in` port.
- `error`: Sends packets containing error information. Error handling results can be sent back to an `in` port.

When multiple ports with the same role are needed, they can be expressed as `in[0]`, `in[1]`, `out[0]`, `out[1]`, etc.

## Packet

A packet is a unit of data exchanged between ports. Each packet contains a payload, which nodes process and transmit.

Nodes must return response packets in the order of received request packets. When connected to multiple ports, all response packets are collected and returned as a new response packet.

A special `None` packet indicates no response, simply acknowledging the packet's acceptance.

## Process

A process is the fundamental unit of execution, managed independently. Processes can have parent processes, and when a parent process terminates, its child processes also terminate.

Processes have their own storage to retain values that are difficult to transmit via packets. This storage operates using a Copy-On-Write (COW) mechanism to efficiently share data from parent processes.

A new workflow begins by creating a process. When a process ends, all resources used by it are released.

Processes can have multiple root packets, but root packets must originate from the same node. If they come from different nodes, a new child process must be created to handle them.

## Workflow

A workflow is defined as a directed graph with multiple interconnected nodes. Each node processes data, and packets flow between nodes.

Workflows consist of multiple stages, with data processed and transmitted according to defined rules at each stage. Data can be processed sequentially or in parallel during this process.

For example, given initial data, it is processed by the first node and then forwarded to the next node. Each node receives input, processes it, and sends the processed result to the next stage.

## Namespace

A namespace manages workflows in isolation, providing an independent execution environment. Each namespace can contain multiple workflows, and nodes within a namespace cannot reference nodes from another namespace. Each namespace independently manages its data and resources.

## Runtime Environment

A runtime environment is an independent space where each namespace is executed. The engine loads all nodes within the namespace to build the environment and execute workflows. This prevents conflicts during workflow execution and ensures a stable execution environment.
