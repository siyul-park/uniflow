# Switch Node

**The Switch Node** provides the functionality to route input packets to one of several ports based on specified conditions. This node evaluates given conditions and directs packets to the appropriate port, supporting complex data flow control and flexible routing.

## Specification

- **matches**: A list of conditions for routing packets. Each condition defines a rule for routing packets to a specific port.
  - **when**: An expression that defines the condition. It is compiled and executed using Common Expression Language (CEL).
  - **port**: Specifies the port to route packets to when the condition is met.

## Ports

- **in**: Receives input packets and routes them based on conditions.
- **out[*]**: Outputs packets routed to the corresponding port based on conditions.
- **error**: Passes errors that occur during condition evaluation to the external system.

## Example

```yaml
- kind: switch
  matches:
    - when: "payload['status'] == 'success'"
      port: out[0]
    - when: "payload['status'] == 'error'"
      port: out[1]
  ports:
    out[0]:
      - name: success
        port: in
    out[1]:
      - name: error
        port: in
```
