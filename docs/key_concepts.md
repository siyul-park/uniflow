# 📚 Key Concepts

This guide provides a detailed explanation of the key terms and concepts used within the system.

## Node Specification

A node specification declaratively define how each node operates and connects. The engine compiles these specifications into executable nodes.

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

- `id`: A unique identifier in UUID format. UUID V7 is recommended.
- `kind`: Specifies the type of the node. Additional fields may vary based on the node type.
- `namespace`: Specifies the namespace to which the node belongs, defaulting to `default`.
- `name`: Specifies the name of the node, which must be unique within the same namespace.
- `annotations`: Additional metadata for the node, including user-defined key-value pairs such as description and version.
- `protocol`: Specifies the protocol used by the listener. This is an additional required field for nodes of the `listener` type.
- `port`: Specifies the port used by the listener. This is an additional required field for nodes of the `listener` type.
- `ports`: Defines how the ports are connected. `out` defines an output port named `proxy`, which connects to the `in` port of another node.
- `env`: Specifies the environment variables required by the node. In this case, `PORT` is dynamically set from a value.

## Value

A value securely store sensitive information needed by nodes, such as passwords and API keys.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: database
annotations:
  description: "Database information"
data:
  password: "super-value-password"
```

- `id`: A unique identifier in UUID format. UUID V7 is recommended.
- `namespace`: Specifies the namespace to which the value belongs, defaulting to `default`.
- `name`: Specifies the name of the value, which must be unique within the same namespace.
- `annotations`: Additional metadata for the value, including user-defined key-value pairs such as description and
  version.
- `data`: Contains the value data structured as key-value pairs.

## Node

A node is an object that processes data, executing workflows by sending and receiving packets through connected ports. Each node has its own independent processing loop and communicates asynchronously with other nodes.

Nodes are classified based on how they process packets:
- `ZeroToOne`: A node that generates an initial packet to start the workflow.
- `OneToOne`: A node that receives packets from an input port, processes them, and sends them to an output port.
- `OneToMany`: A node that receives packets from an input port and sends them to multiple output ports.
- `ManyToOne`: A node that receives packets from multiple input ports and sends them to a single output port.
- `Other`: A node that includes state management and interaction beyond simple packet forwarding.

## Port

Ports are connection points for sending and receiving packets between nodes. There are two types of ports: `InPort` and `OutPort`, and packets are transmitted by connecting them. A packet sent to one port is forwarded to all connected ports.

Commonly used port names include:
- **`init`**: A special port used to initialize nodes. When a node becomes available, the workflow connected to the `init` port is executed.
- **`term`**: A special port used to terminate nodes. The workflow connected to the `term` port is executed when the node is being deactivated.
- **`io`**: Processes packets and immediately returns the results.
- **`in`**: Receives packets for processing and sends the results to the `out` or `error` ports. If there are no connected `out` or `error` ports, the result is returned directly.
- **`out`**: Sends processed packets. The transmitted results can be sent to other `in` ports.
- **`error`**: Sends errors encountered during packet processing. The results of error handling can be sent back to an `in` port.

When multiple ports with the same role are needed, they are expressed as `in[0]`, `in[1]`, `out[0]`, `out[1]`, etc.

## Packet

A packet is a unit of data exchanged between ports. Each packet includes a payload, which the node processes before transmission.

Nodes must return response packets in the order of the request packets. If connected to multiple ports, all response packets are gathered and returned as a single new response packet.

A special `None` packet indicates that there is no response, merely indicating that the packet was accepted.

## Process

A process is the basic unit of execution, managed independently. A process may have a parent process, and when the parent terminates, the child process also terminates.

Processes maintain their own storage to store values that are difficult to transmit as packets. This storage operates on a Copy-On-Write (COW) basis, efficiently sharing data from the parent process.

A new workflow is initiated by creating a process. When the process terminates, all resources used are released.

Processes can have two or more root packets, but the root packets must be generated by the same node. If generated by different nodes, a new child process must be created to handle them.

## Workflow

A workflow is defined as a directed graph, a structure where multiple nodes are interconnected. In this graph, each node is responsible for data processing, and packets are transmitted between nodes.

Workflows consist of multiple stages, where data is processed and passed according to defined rules. In this process, data can be processed sequentially or in parallel.

For example, given initial data, it is processed by the first node before being passed to the next node. Each node receives input, processes it, and sends the result to the next stage.

## Namespace

A namespace serves to isolate and manage workflows, offering independent execution environments. It can house multiple workflows, ensuring that nodes within a namespace cannot reference nodes from other namespaces. Each namespace independently manages its own data and resources, allowing for streamlined operations and clear boundaries between workflows.

## Runtime Environment

The runtime environment provides an independent execution context for each namespace, ensuring that workflows within namespaces operate without interference. This environment includes the necessary resources and configurations for the execution of nodes and processes, facilitating the smooth operation of workflows within the system.
