# 📚 Key Concepts

This guide provides a detailed explanation of key terms and concepts used in our system.

## Node Specifications

Node specifications define the behavior and connections of each node in a declarative format. The engine compiles these specifications into executable nodes.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01a
kind: listener
namespace: default
name: example-listener
annotations:
  description: "An example listener node for HTTP protocol"
  version: "1.0"
protocol: http
port: "{{ .PORT }}"
ports:
  out:
    - name: proxy
      port: in
env:
  PORT:
    name: network
    value: "{{ .PORT }}"
```

- `id`: A unique identifier in UUID format. UUID V7 is recommended.
- `kind`: Specifies the type of node, in this case, a `listener`.
- `namespace`: The namespace the node belongs to, defaulting to `default`.
- `name`: The name of the node, which must be unique within the same namespace.
- `annotations`: Additional metadata about the node. This can include descriptions, versions, or other custom key-value pairs.
- `protocol`: Specifies the protocol used by the listener, in this case, `http`.
- `ports`: Defines how ports are connected. `out` specifies the output port named `proxy`, connected to the `in` port of another node.
- `env`: Environment variables for the node. Here, `PORT` is specified to be dynamically set.

## Secrets

Secrets securely store sensitive information such as passwords, API keys, or other confidential data that nodes may require. The following structure illustrates a secret definition:

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: my-secret
annotations:
  purpose: "database password"
data:
  password: "super-secret-password"
```

- `id`: Unique identifier for the secret.
- `namespace`: Namespace the secret belongs to.
- `name`: Name of the secret, which must be unique within the namespace.
- `annotations`: Additional metadata for the secret.
- `data`: Contains the secret data as key-value pairs.

## Nodes

Nodes are data processing objects that execute workflows by exchanging packets through connected ports. Each node operates independently with its own processing loop, communicating asynchronously with other nodes.

Nodes are categorized based on how they process packets:
- `ZeroToOne`: Nodes that generate the initial packet to start the workflow.
- `OneToOne`: Nodes that receive a packet on an input port, process it, and send it to an output port.
- `OneToMany`: Nodes that receive a packet on an input port, process it, and send it to multiple output ports.
- `ManyToOne`: Nodes that receive packets on multiple input ports, aggregate them, and send them to a single output port.
- `Other`: Nodes that manage state and interactions beyond simple packet forwarding.

## Ports

Ports are connection points through which nodes exchange packets. There are two types of ports: `InPort` and `OutPort`. Connecting these ports allows packet transmission. Packets sent to a port are forwarded to all connected ports.

Commonly used port names:
- `init`: Special port used to initialize nodes. The workflow connected to the `init` port is executed when the node becomes available.
- `io`: Processes packets and immediately returns them.
- `in`: Receives and processes packets, sending the results to `out` or `error`. If no `out` or `error` port is connected, the result is returned.
- `out`: Sends processed packets. The result can be sent to another `in` port for further processing.
- `error`: Sends error packets if processing fails. Error handling results can be sent to an `in` port for recovery or logging.

When multiple ports serve the same role, they can be indexed as `in[0]`, `in[1]`, `out[0]`, `out[1]`, etc.

## Packets

Packets are units of data exchanged between ports. Each packet contains a payload that nodes process and transmit.

Nodes must return a response packet corresponding to the received request packet. When connected to multiple ports, nodes aggregate all response packets into a single new response packet.

A special `None` packet exists to indicate the absence of a response, showing that the packet was simply accepted without further processing.

## Process

A process is the basic unit of execution, managed independently. Processes can have parent processes, and child processes terminate when their parent process does.

Processes have their own storage for values that are difficult to transmit as packets. This storage operates in a Copy-On-Write (COW) manner, efficiently sharing data from parent processes.

New workflows start with the creation of a process. When the process ends, all resources used are released.

Processes can have multiple root packets from the same node. If root packets originate from different nodes, a new child process is created to handle the flow.

## Workflow

A workflow is defined as a directed graph where multiple nodes are connected. In this graph, each node is responsible for data processing, and packets are transmitted between nodes.

Workflows consist of a series of steps where data is processed and transmitted according to defined rules. Data can be processed sequentially or in parallel during this process.

For example, given initial data, the first node processes it and then passes it to the next node. Each node receives input, processes it, and sends the processed result to the next step.

## Namespace

Namespaces isolate and manage workflows, providing independent execution environments. Each namespace can include multiple workflows, and nodes within a namespace cannot reference nodes from other namespaces. Each namespace manages its data and resources independently.

## Runtime Environment

A runtime environment is an independent space where each namespace is executed. The engine loads all nodes belonging to a namespace to build the environment and execute the workflow. This prevents conflicts that might occur during workflow execution and ensures a stable execution environment.
