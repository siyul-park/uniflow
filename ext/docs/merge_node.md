# Merge Node

**The Merge Node** provides the functionality to combine multiple input packets into a single output packet. It is useful for aggregating data from various input sources into one packet for further processing or transmission.

## Specification

- No additional arguments are required.

## Ports

- **in[*]**: Receives multiple input packets. Each input port accepts packets from different data sources and supports various formats.
- **out**: Outputs the result of merging the input packets into a single packet.
- **error**: Passes any errors encountered during the merging process.

## Example

```yaml
- kind: snippet
  language: json
  code: 0
  ports:
    out:
      - name: merge
        port: in[0]

- kind: snippet
  language: json
  code: 1
  ports:
    out:
      - name: merge
        port: in[1]

- kind: merge
  name: merge
  ports:
    out:
      - name: next
        port: out
```
