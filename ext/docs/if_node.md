# If Node

**The If Node** provides the capability to branch packets based on a given condition, directing them along one of two paths. This node evaluates the condition and executes different data flows depending on the result.

## Specification

- **when**: An expression that defines the condition. This expression is written in `Common Expression Language (CEL)`, compiled, and executed.

## Ports

- **in**: Receives the input packet and evaluates the condition to determine branching.
- **out[0]**: Passes the packet if the condition is true.
- **out[1]**: Passes the packet if the condition is false.
- **error**: Sends any errors encountered during condition evaluation to the external environment.

## Example

```yaml
- kind: if
  when: "self.count > 10"
  ports:
    out[0]:
      - name: true_path
        port: out
    out[1]:
      - name: false_path
        port: out
```
