# Retry Node

**The Retry Node** retries packet processing multiple times in case of errors. This node is useful for tasks prone to temporary failures, providing multiple attempts to improve the chances of success before ultimately sending the packet to an error output if the retries are exhausted.

## Specification

- **limit**: Specifies the maximum number of retry attempts for processing a packet in case of failure. Once the retry limit is exceeded, the packet is routed to the error output port.

## Ports

- **in**: Receives the input packet and initiates processing. The packet will be retried until the `limit` is reached if processing fails.
- **out**: Outputs the packet if processing is successful within the retry limit.
- **error**: Routes the packet to the error output if it exceeds the retry limit without success.

## Example

```yaml
- kind: retry
  limit: 3 # Retry up to 3 times
```
