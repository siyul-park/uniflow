# Signal Node

**The Signal Node** listens to a signal channel and forwards received signals as packets. This node is useful for
processing real-time events or system-level signals within the workflow.

## Specification

- **opcode**: A string identifying the system operation to be listened. It is associated with the specified function and
  determines the node's behavior.

## Ports

- **out**: Sends packets containing signal data when a signal is received.

## Example

```yaml
- kind: signal
  opcode: specs.watch
  ports:
    out:
      - name: next
        port: in
```
