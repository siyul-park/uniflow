# Call Node

**The Call Node** processes the input packet and delivers the result to multiple output ports. This node allows for the reuse of data processing flows and is useful for modularizing complex tasks.

## Specification

- No additional arguments are required.

## Ports

- **in**: Processes the input packet.
- **out[0]**: Passes the input packet to the first processing node and outputs the result externally.
- **out[1]**: Passes the result of the first processing node to the second processing node.
- **error**: Sends any errors encountered during processing to the external environment.

## Example

```yaml
- kind: call
  ports:
    out[0]:
      - name: origin
        port: in
    out[1]:
      - name: next
        port: in
```
