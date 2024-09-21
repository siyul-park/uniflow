# Reduce Node

**The Reduce Node** provides functionality to iteratively compute an output value from input data. It is useful for data aggregation or transformation tasks.

## Specification

- **action**: Defines an operation that takes two input values and returns one output value. This operation is written in `Common Expression Language (CEL)` and processes data cumulatively. (Required)
- **init**: Sets the initial value (optional).

## Ports

- **in**: The port that receives input data for the reduction operation. The accumulated value is passed as the first argument, and the current value as the second.
- **out**: The port that outputs the result of the reduction operation.
- **error**: Returns any errors encountered during the execution of the operation.

## Example

```yaml
- kind: reduce
  action: "self[0] + self[1]"
  init: 0
  ports:
    out:
      - name: result
        port: out
``` 
