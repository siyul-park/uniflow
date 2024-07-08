# Key Concepts

This guide explains key terms and concepts in detail.

## Node Specification

A node specification defines the behavior and port connections of each node declaratively. The engine compiles this specification into an executable node.

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
links:
  out:
    - name: bar
      port: in
    - id: 01908c74-8b22-7cc0-ae2b-40504e7c9ff0
      port: in
```

- `id`: A UUID, with UUID V7 recommended.
- `kind`: Specifies the type of node, with different specifications depending on the type.
- `namespace`: Specifies the namespace the node belongs to, defaulting to `default`.
- `name`: Specifies the name of the node, which must be unique within the same namespace.
- `annotations`: Contains additional metadata for the node specification. These values are not used by the engine but are useful for providing extended functionalities or integrating with external services. Keys and values are in string format.
- `language`, `code`: Additional fields required for nodes of type `snippet`. Required fields may vary based on the node type.
- `links`: Defines how ports are connected. Each port is identified by the ID or name and port name of another node.

## Node

A node is a data processing object that executes workflows by sending and receiving packets through connected ports. Each node operates in an independent processing loop, communicating asynchronously with other nodes.

Nodes are classified into five types based on packet processing methods:
- `ZeroToOne`: Nodes that generate packets to start the workflow.
- `OneToOne`: Nodes that receive a packet from one input port, process it, and send it to one output port.
- `OneToMany`: Nodes that receive a packet from one input port, process it, and send it to multiple output ports.
- `ManyToOne`: Nodes that receive packets from multiple input ports, aggregate them, and send them to one output port.
- `Other`: Nodes that involve more than simple packet forwarding, including state management and interaction.

## Port

A port is a connection point where nodes send and receive packets. There are two types of ports: `InPort` and `OutPort`, which are connected to transmit packets. A packet sent to one port is delivered to all connected ports.

Commonly used port names include:
- `io`: Processes the packet and returns it immediately.
- `in`: Receives a packet, processes it, and sends the result to either `out` or `error`. If no `out` or `error` ports are connected, it returns the result.
- `out`: Sends the processed packet. The result can be sent back to an `in` port.
- `error`: Sends errors that occur during packet processing. The error handling result can be sent back to an `in` port.

When multiple ports with the same function are needed, they are denoted as `in[0]`, `out[1]`, etc.

## Packet

A packet is the data exchanged between ports, containing a payload.

Nodes must return the corresponding response packet in the order the request packets were sent. If connected to multiple ports, all response packets are aggregated into a single new response packet returned to the node.

There is also a special `None` packet, indicating no response.

Packets are transmitted and processed along the connections between nodes. Forward propagation is when the packet follows the initially connected port, while back propagation occurs when the packet is returned in the opposite direction after processing is complete.

## Process

A process is the basic unit of execution, managed independently. A process can have a parent process, and if the parent process terminates, so does the child process.

Processes have their own storage to hold values that are difficult to transmit via packets, operating with Copy-On-Write (COW) to efficiently share data from the parent process.

A process is created to start a new workflow, and all resources used are released when the process terminates.

A process can have more than one root packet, but root packets must be generated from the same node. If they are generated from different nodes, a new child process is created to handle the flow.

## Workflow

A workflow is defined as a directed graph with multiple connected nodes.

In this graph, each node handles data processing, and packets are transmitted between these nodes.

A workflow consists of a series of steps, where data is processed and transmitted according to defined rules. Data can be processed sequentially or in parallel.

For example, given initial data, it is processed by the first node and then passed to the next node. Each node receives input, processes it, and sends the processed result to the next step.

## Namespace

A namespace manages workflows in isolation, providing an independent execution environment. Each namespace can include multiple workflows, and nodes within a namespace cannot reference nodes in other namespaces. Each namespace independently manages its own data and resources.

## Runtime Environment

The runtime environment is an independent space where each namespace runs. The engine loads all nodes belonging to a namespace, constructs the environment, and executes the workflows. This prevents potential conflicts during workflow execution and ensures a stable execution environment.
