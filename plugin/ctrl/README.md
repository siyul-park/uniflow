# 🎮 Ctrl Plugin

The **Ctrl Plugin** provides a set of nodes that allow for precise management of data flow. It enables efficient
processing of complex logic and offers flexible flow control mechanisms.

## Available Nodes

- **[Block Node](./docs/block_node.md)**: Groups multiple sub-nodes to manage complex data flows and executes them
  sequentially.
- **[For Node](./docs/for_node.md)**: Divides the input packet into multiple sub-packets for iterative processing.
- **[Fork Node](./docs/fork_node.md)**: Branches the data flow to perform parallel processing.
- **[If Node](./docs/if_node.md)**: Splits data into two paths based on a condition.
- **[Merge Node](./docs/merge_node.md)**: Combines multiple input packets into a single output.
- **[NOP Node](./docs/nop_node.md)**: Does not process the input packet and returns an empty response.
- **[Pipe Node](./docs/pipe_node.md)**: Routes results to multiple output ports, allowing data flow reuse.
- **[Retry Node](./docs/retry_node.md)**: Retries packet processing in the event of an error, ensuring reliability.
- **[Sleep Node](./docs/sleep_node.md)**: Pauses execution for a specified duration before continuing with the workflow.
- **[Snippet Node](./docs/snippet_node.md)**: Executes user-defined code snippets to process input packets.
- **[Split Node](./docs/split_node.md)**: Splits an input packet into multiple sub-packets for parallel processing.
- **[Step Node](./docs/step_node.md)**: Executes sub-nodes sequentially, managing and controlling the data flow.
- **[Switch Node](./docs/switch_node.md)**: Branches data into multiple paths based on conditions, providing flexible
  flow control.
- **[Throw Node](./docs/throw_node.md)**: Generates an error and returns it as a response for error-handling logic.
- **[Try Node](./docs/try_node.md)**: Handles errors and responds appropriately to exceptions.
