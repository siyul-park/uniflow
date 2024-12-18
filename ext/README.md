# 🔧 Built-in Extensions

Built-in extensions enable efficient processing of various tasks and maximize system performance.

## Available Extensions

### **Control**

Precisely manage data flow.

- **[Fork Node](./docs/fork_node.md)**: Asynchronously branches data flows to perform independent tasks in parallel.
- **[If Node](./docs/if_node.md)**: Evaluates conditions to split packets into two paths.
- **[Loop Node](./docs/loop_node.md)**: Divides input packets into multiple sub-packets for repeated processing.
- **[Merge Node](./docs/merge_node.md)**: Combines multiple input packets into one.
- **[NOP Node](./docs/nop_node.md)**: Responds to input packets with an empty packet, without any processing.
- **[Pipe Node](./docs/pipe_node.md)**: Processes input packets and distributes results to multiple output ports, allowing for reusable data flows.
- **[Reduce Node](./docs/reduce_node.md)**: Repeatedly processes input data to produce a single output value, useful for data aggregation.
- **[Retry Node](./docs/retry_node.md)**: Retries packet processing a specified number of times upon encountering errors.
- **[Sequential Node](./docs/sequential_node.md)**: Systematically manages complex data processing flows and executes multiple sub-nodes sequentially.
- **[Session Node](./docs/session_node.md)**: Stores and manages process information, maintaining session continuity.
- **[Snippet Node](./docs/snippet_node.md)**: Executes code snippets written in various programming languages to process input packets.
- **[Split Node](./docs/split_node.md)**: Splits input packets into multiple parts for processing.
- **[Switch Node](./docs/switch_node.md)**: Routes input packets to one of several ports based on conditions.
- **[Wait Node](./docs/wait_node.md)**: Introduces a specified delay in processing to pace workflows or await external conditions.

### **IO**

Supports interaction with external data sources.

- **[Print Node](./docs/print_node.md)**: Outputs input data to a file for debugging or monitoring data flow.
- **[Scan Node](./docs/scan_node.md)**: Scans various input data formats to extract and process required data.
- **[SQL Node](./docs/sql_node.md)**: Interacts with relational databases to execute SQL queries and return results as packets.

### **Network**

Facilitates smooth execution of network-related tasks across various protocols.

- **[HTTP Node](./docs/http_node.md)**: Processes HTTP requests and returns responses, suitable for web service communication.
- **[WebSocket Node](./docs/websocket_node.md)**: Establishes WebSocket connections and handles message sending and receiving.
- **[Gateway Node](./docs/gateway_node.md)**: Upgrades HTTP connections to WebSocket for real-time data communication.
- **[Listener Node](./docs/listener_node.md)**: Receives network requests on specified protocols and ports.
- **[Proxy Node](./docs/proxy_node.md)**: Proxies HTTP requests to other servers and returns their responses.
- **[Router Node](./docs/router_node.md)**: Routes input packets to multiple output ports based on conditions.

### **System**

Manages and optimizes system components.

- **[Native Node](./docs/native_node.md)**: Performs function calls within the system and returns results as packets.