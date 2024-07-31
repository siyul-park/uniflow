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
  description: "An example listener node using the HTTP protocol"
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
- `kind`: Specifies the type of node. In this example, it is a `listener`. Additional fields may vary depending on the node type.
- `namespace`: Specifies the namespace to which the node belongs, with a default value of `default`.
- `name`: Specifies the name of the node, which must be unique within the same namespace.
- `annotations`: Additional metadata about the node, such as a description, version, and other custom key-value pairs.
- `protocol`: Specifies the protocol used by the listener. In this example, it is `http`. This is an additional field required by nodes of the `listener` type.
- `ports`: Defines the connection scheme of the ports. `out` defines an output port named `proxy` that connects to the `in` port of another node.
- `env`: Specifies the environment variables needed by the node. Here, `PORT` is dynamically set.

## Secret

Secrets securely store sensitive information required by nodes, such as passwords or API keys. Below is an example of a secret definition:

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
- `namespace`: Specifies the namespace to which the secret belongs, with a default value of `default`.
- `name`: Specifies the name of the secret, which must be unique within the same namespace.
- `annotations`: Additional metadata about the secret, such as a description, version, and other custom key-value pairs.
- `data`: Contains the secret data as key-value pairs.

## Node

Nodes are objects that process data and execute workflows by sending and receiving packets through connected ports. Each node has an independent processing loop and communicates asynchronously with other nodes.

Nodes are categorized based on how they handle packets:
- `ZeroToOne`: Nodes that generate initial packets to start workflows.
- `OneToOne`: Nodes that receive packets from an input port, process them, and send them to an output port.
- `OneToMany`: Nodes that receive packets from an input port and send them to multiple output ports.
- `ManyToOne`: Nodes that receive packets from multiple input ports and send them to a single output port.
- `Other`: Nodes that involve state management and interactions beyond simple packet forwarding.

## Port

Ports are connection points for sending and receiving packets between nodes. There are two types of ports: `InPort` and `OutPort`, which are connected to transmit packets. A packet sent to one port is delivered to all connected ports.

Commonly used port names include:
- `init`: A special port used to initialize the node. When the node becomes available, workflows connected to the `init` port are executed.
- `io`: Processes packets and immediately returns them.
- `in`: Receives packets for processing and sends the results to `out` or `error`. If no `out` or `error` port is connected, it returns the results.
- `out`: Sends processed packets, which can be received by another `in` port.
- `error`: Sends error packets resulting from processing. The error handling results can be sent back to the `in` port.

When multiple ports serving the same role are needed, they are expressed as `in[0]`, `in[1]`, `out[0]`, `out[1]`, and so on.

## Packet

A packet is a unit of data exchanged between ports. Each packet contains a payload, which nodes process and transmit.

Nodes must return response packets in the order of the request packets. When connected to multiple ports, all response packets are combined into a new response packet.

A special `None` packet indicates no response, simply acknowledging that the packet was accepted.

## Process

A process is the basic unit of execution, managed independently. Processes can have parent processes, and when a parent process terminates, its child processes are also terminated.

Processes have their own storage to hold values that are difficult to transmit as packets. This storage operates using a Copy-On-Write (COW) mechanism to efficiently share data with parent processes.

A new workflow begins by creating a process. When the process terminates, all resources used are released.

Processes can have more than one root packet, but root packets must be created by the same node. If they are created by different nodes, new child processes must be created to handle them.

## Workflow

A workflow is defined as a directed graph comprising multiple interconnected nodes. In this graph, each node handles data processing, and packets are transmitted between nodes.

Workflows consist of several stages, where data is processed and transmitted according to defined rules. This processing can be sequential or parallel.

For example, given initial data, the first node processes it and then passes it to the next node. Each node receives input, processes it, and sends the result to the next stage.

## Namespace

Namespaces isolate and manage workflows, providing an independent execution environment. Each namespace can contain multiple workflows, and nodes within a namespace cannot reference nodes in another namespace. Each namespace manages its own data and resources independently.

## Runtime Environment

A runtime environment is the independent space where each namespace is executed. The engine loads all nodes within a namespace to build the environment and execute the workflows. This prevents conflicts and ensures a stable execution environment during workflow execution.
