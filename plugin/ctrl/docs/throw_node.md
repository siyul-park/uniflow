# Throw Node

**The Throw Node** is used to generate errors based on the input packet. The generated errors are returned as a response
through the input port.

## Specification

- No additional parameters are required.

## Ports

- **in**: If the input is a list, each element is converted into an individual error message. If the input is not a
  list, it is processed as a single error message.

## Example

```yaml
- kind: throw
```
