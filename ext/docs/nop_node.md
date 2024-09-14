# NOP Node

**The NOP Node** does not process input packets and simply responds with an empty packet. This node is used as a final stage in data processing flows and is useful for eliminating unnecessary outputs when no further processing is required.

## Specification

- No additional parameters are required.

## Ports

- **in**: Receives input packets from external sources and responds with an empty packet.

## Example

```yaml
- kind: print
  filename: /dev/stdout
  ports:
    out:
      - name: nop
        port: in

- kind: nop
  name: nop
```
