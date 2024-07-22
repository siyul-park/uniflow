# 📚 Key Concepts

This guide explains key terms and concepts in detail.

## Node Specification

A node specification declaratively defines how each node operates and connects to ports. This specification is compiled into executable nodes by the engine.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01a
kind: snippet
namespace: default
name: foo
annotations:
  key1: value1
  key2: value2
language: text
code: foo
ports:
  out:
    - name: bar
      port: in
    - id: 01908c74-8b22-7cc0-ae2b-40504e7c9ff0
      port: in
```

- `id`: A UUID, preferably in UUID V7 format.
- `kind`: Specifies the type of node, with different specifications depending on the kind.
- `namespace`: Specifies the namespace the node belongs to, with `default` as the default.
- `name`: Specifies the name of the node, which must be unique within the same namespace.
- `annotations`: Includes additional metadata for the node specification. These values are not used by the engine but are useful for providing extensions or integrating with external services. Both keys and values are strings.
- `language`, `code`: Additional fields required for `snippet` type nodes. Required fields may vary depending on the node type.
- `ports`: Defines how ports are connected. Each port is identified by the ID or name and port name of another node.

## Node

Nodes are data processing objects that execute workflows by sending and receiving packets through interconnected ports. Each node has an independent processing loop, allowing asynchronous communication with other nodes.

Nodes are classified into five types based on packet processing:
- `ZeroToOne`: Generates the first packet to start the workflow.
- `OneToOne`: Receives a packet from one input port, processes it, and sends it to one output port.
- `OneToMany`: Receives a packet from one input port, processes it, and sends it to multiple output ports.
- `ManyToOne`: Receives packets from multiple input ports, aggregates them, processes them, and sends them to one output port.
- `Other`: Includes nodes that manage state and interactions beyond simple packet forwarding.

## Port

Ports are connection points where nodes send and receive packets. There are two types of ports, `InPort` and `OutPort`, which are connected to transmit packets. Packets sent to one port are delivered to all connected ports.

Common port names include:
- `io`: Processes and immediately returns packets.
- `in`: Receives packets, processes them, and sends the result to `out` or `error`. If no `out` or `error` ports are connected, it returns the result.
- `out`: Sends processed packets. The results sent may be returned to `in`.
- `error`: Sends errors that occur during packet processing. Error handling results may be returned to `in`.

When multiple ports with the same role are needed, they are expressed as `in[0]` or `out[1]`.

## Packet

A packet is the data exchanged between ports, and each packet includes a payload.

Nodes must return the corresponding response packet according to the transmission order of the request packets. When connected to multiple ports, all response packets are aggregated into a single new response packet and returned to the node.

A special `None` packet indicates that the packet was simply accepted, showing that there is no response.

Packets are transmitted along node connections for processing. Forward propagation occurs when packets follow the connected ports, and backpropagation occurs when responses are returned after all processing is complete.

## Process

A process is the basic unit of execution and is managed independently. Processes can have parent processes, and child processes terminate when the parent process terminates.

Each process has its own storage to save values that are difficult to transmit via packets, and operates using a Copy-On-Write (COW) method to efficiently share the parent process's data.

A new process is created to start a new workflow, and all resources used are released when the process terminates.

Processes can have more than one root packet, but root packets must be generated from the same node. If they are generated from different nodes, a new child process must be created to handle the flow.

## Workflow

A workflow is defined as a directed graph where multiple nodes are connected. In this graph, each node is responsible for data processing, and packets are transmitted between nodes.

Workflows consist of a series of steps where data is processed and transmitted according to defined rules. Data can be processed sequentially or in parallel during this process.

For example, given initial data, the first node processes it and then passes it to the next node. Each node receives input, processes it, and sends the processed result to the next step.

## Namespace

Namespaces isolate and manage workflows, providing independent execution environments. Each namespace can include multiple workflows, and nodes within a namespace cannot reference nodes from other namespaces. Each namespace manages its data and resources independently.

## Runtime Environment

A runtime environment is an independent space where each namespace is executed. The engine loads all nodes belonging to a namespace to build the environment and execute the workflow. This prevents conflicts that might occur during workflow execution and ensures a stable execution environment.
