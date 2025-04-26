# Split Node

**The Split Node** provides the functionality to divide an input packet into multiple packets for processing. This node
splits the input data according to a specified format and routes each split packet through different paths, allowing for
efficient management of complex data flows.

## Specification

- No additional parameters are required.

## Ports

- **in**: Receives the input packet and splits it into multiple packets. If the input is not in array format, it will be
  output as a single packet.
- **out[*]**: Outputs the split packets through multiple ports.

## Example

```yaml
- kind: split
  ports:
    out[0]:
      - name: next0
        port: in
    out[1]:
      - name: next1
        port: in
```
