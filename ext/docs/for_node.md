# For Node

**The For Node** provides the functionality to split an incoming packet into multiple sub-packets for repeated
processing. This is useful in data processing flows where repeated tasks are needed, and it integrates the results of
each sub-packet before returning them.

## Specification

- No additional arguments are required.

## Ports

- **in**: Receives packets from external sources and initiates the repeat operation. If the input is an array, each element is split into sub-packets and processed individually. If the input is not an array, it will be processed only once.
- **out[0]**: Passes the split sub-packets to the first output port.
- **out[1]**: Aggregates the results of all sub-packet processing and passes them to the second output port.
- **error**: Sends any errors encountered during processing to the external environment.

## Example

```yaml
- kind: for
  ports:
    out[0]:
      - name: next
        port: out
    out[1]:
      - name: done
        port: out
```
