# 🔧 Built-in Extensions

Built-in extensions enhance system performance and streamline task handling through various functionalities.

## Available Extensions

### **Control**

Provides precise control over data flow.

- **[Call Node](./docs/call_node.md)**: Processes input packets and distributes results to multiple output ports. Enables reuse of data processing flows.
- **[Block Node](./docs/block_node.md)**: Manages complex data processing flows by sequentially executing multiple sub-nodes.
- **[Fork Node](./docs/fork_node.md)**: Splits data flow asynchronously, allowing for parallel execution of independent tasks.
- **[If Node](./docs/if_node.md)**: Evaluates a condition to route packets to one of two paths based on the condition.
- **[Loop Node](./docs/loop_node.md)**: Divides input packets into multiple sub-packets for iterative processing. Suitable for repeated processing tasks.
- **[Merge Node](./docs/merge_node.md)**: Combines multiple input packets into a single output packet. Merges data from various sources.
- **[NOP Node](./docs/nop_node.md)**: Responds to input packets with an empty packet, used as a terminal node in a data processing flow.
- **[Session Node](./docs/session_node.md)**: Stores and manages process-specific information, maintaining session continuity while the process is active.
- **[Snippet Node](./docs/snippet_node.md)**: Executes code snippets in various programming languages to process input packets, enabling flexible application of complex logic.
- **[Split Node](./docs/split_node.md)**: Divides input packets into multiple packets for processing, based on specified formats, and routes them through different paths.
- **[Switch Node](./docs/switch_node.md)**: Routes input packets to one of several output ports based on specified conditions.

### **IO**

Facilitates interaction with external data sources.

- **[Print Node](./docs/print_node.md)**: Outputs input data to a file for debugging or monitoring purposes.
- **[Scan Node](./docs/scan_node.md)**: Scans and extracts data from various input formats.
- **[SQL Node](./docs/sql_node.md)**: Executes SQL queries on relational databases and returns results as packets. Configures database connections and transaction isolation levels.

### **Network**

Supports various network protocols for efficient network-related tasks.

- **[HTTP Node](./docs/http_node.md)**: Processes HTTP requests and returns responses, suitable for web service communication.
- **[WebSocket Node](./docs/websocket_node.md)**: Establishes and manages WebSocket connections, handling message transmission and reception.
- **[Gateway Node](./docs/gateway_node.md)**: Upgrades HTTP connections to WebSocket connections for real-time data communication.
- **[Listener Node](./docs/listener_node.md)**: Receives and processes network requests on specified protocols and ports, functioning as an HTTP server.
- **[Router Node](./docs/router_node.md)**: Routes input packets to multiple output ports based on conditions, including HTTP methods and paths.

### **System**

Manages and optimizes system components.

- **[Native Node](./docs/native_node.md)**: Performs system-level function calls and converts results into packets for return.
