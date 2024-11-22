# Native Node

**The Native Node** performs function call operations within the system. This node processes system calls based on the `opcode`, passing input packets to the function for execution and returning the result.

## Specification

- **opcode**: A string identifying the system operation to be invoked. It is associated with the specified function and determines the node's behavior.

## Ports

- **in**: Receives input packets and converts them into arguments for the specified function call. The payload of the packet is adjusted to match the function's parameters.
- **out**: Returns the result of the function call. If the function returns multiple values, the result is output as an array.
- **error**: Returns any errors encountered during the function call.

## Example

```yaml
- kind: snippet
  language: cel
  code: 'has(self.body) ? self.body : null'
  ports:
    out:
      - name: specs_create
        port: in

- kind: native
  name: specs_create
  opcode: specs.create
```
