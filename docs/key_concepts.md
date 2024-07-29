# 📚 Key Concepts

This guide details the main terms and concepts used in the system.

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
- `kind`: Specifies the type of node. In this example, it is a `listener`.
- `namespace`: Specifies the namespace to which the node belongs, defaulting to `default`.
- `name`: The name of the node, which must be unique within the same namespace.
- `annotations`: Additional metadata about the node, including user-defined key-value pairs for description, version, etc.
- `protocol`: The protocol used by the listener. In this example, it is `http`.
- `ports`: Defines how ports are connected. `out` defines an output port named `proxy` which connects to another node's `in` port.
- `env`: Specifies environment variables required by the node. Here, `PORT` is dynamically set.

## Secret

Secrets securely store sensitive information needed by nodes, such as passwords and API keys. Below is an example of a secret definition:

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: my-secret
annotations:
  purpose: "Database password"
data:
  password: "super-secret-password"
```

- `id`: The unique identifier of the secret.
- `namespace`: The namespace to which the secret belongs.
- `name`: The name of the secret, which must be unique within the same namespace.
- `annotations`: Additional metadata about the secret.
- `data`: Contains the secret data as key-value pairs.

## Node

Nodes are entities that process data, exchanging packets through connected ports to execute workflows. Each node operates independently with its own processing loop, communicating asynchronously with other nodes.

Nodes are classified based on how they handle packets:
- `ZeroToOne`: Nodes that generate initial packets to start a workflow.
- `OneToOne`: Nodes that receive packets from an input port, process them, and send them to an output port.
- `OneToMany`: Nodes that receive packets from an input port and send them to multiple output ports.
- `ManyToOne`: Nodes that receive packets from multiple input ports and send them to a single output port.
- `Other`: Nodes that manage state and interactions beyond simple packet forwarding.

## Port

Ports are connection points for sending and receiving packets between nodes. There are two types of ports: `InPort` and `OutPort`. Connecting these ports allows packet transmission. Packets sent to one port are forwarded to all connected ports.

Common port names include:
- `init`: A special port used to initialize nodes. When a node becomes available, workflows connected to the `init` port are executed.
- `io`: Processes packets and immediately returns the result.
- `in`: Receives packets, processes them, and sends the result to `out` or `error`. If no `out` or `error` port is connected, the result is returned.
- `out`: Sends processed packets. The results can be retransmitted to other `in` ports.
- `error`: Sends packets containing errors that occurred during processing. The results can be retransmitted to the `in` port.

When multiple ports of the same role are needed, they can be represented as `in[0]`, `in[1]`, `out[0]`, `out[1]`, etc.

## Packet

Packets are units of data exchanged between ports. Each packet contains a payload that nodes process and transmit.

Nodes must return response packets in the order of the request packets. When connected to multiple ports, all response packets are aggregated into a single new response packet.

A special `None` packet indicates the absence of a response, merely acknowledging that the packet was accepted.

## Process

Processes are the basic units of execution, managed independently. A process can have a parent process, and if the parent process terminates, all child processes are also terminated.

Processes have their own storage for values that are difficult to transmit via packets. This storage operates on a Copy-On-Write (COW) basis, efficiently sharing data from the parent process.

A new workflow starts by creating a process. When a process terminates, all resources used are released.

Processes can have multiple root packets, but these root packets must originate from the same node. If they originate from different nodes, a new child process must be created to handle them.

## Workflow

A workflow is defined as a directed graph where multiple nodes are connected. In this graph, each node is responsible for data processing, and packets are transmitted between nodes.

Workflows are composed of multiple stages, where data is processed and transmitted according to defined rules. This processing can occur sequentially or in parallel.

For example, given initial data, it is processed by the first node and then passed to the next node. Each node processes the input and sends the processed result to the next stage.

## Namespace

A namespace isolates and manages workflows, providing an independent execution environment. Each namespace can contain multiple workflows, and nodes within a namespace cannot reference nodes in other namespaces. Each namespace independently manages its data and resources.

## Runtime Environment

The runtime environment is the independent space where each namespace executes. The engine loads all nodes in a namespace to build the environment and execute workflows. This isolation prevents conflicts and ensures a stable execution environment.
