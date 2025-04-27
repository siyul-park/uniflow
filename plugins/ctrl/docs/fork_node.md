# Fork Node

**The Fork Node** provides the capability to asynchronously branch the data processing flow, allowing it to be handled
in separate processes. This enables parallel processing and allows independent tasks to be performed without blocking
the main flow.

## Specification

- No additional arguments are required.

## Ports

- **in**: Passes the input packet to a new process and returns an empty packet.
- **out**: Outputs the results processed asynchronously.
- **error**: Sends any errors encountered during processing to the external environment.

## Example

```yaml
- kind: fork
  ports:
    out:
      - name: next
        port: out
```
