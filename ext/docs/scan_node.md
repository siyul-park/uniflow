# Scan Node

**The Scan Node** reads data from a file and parses it according to a specified format. It uses the format included in the input packet to read and parse the data from the file and returns the parsed data in an output packet.

## Specification

- **filename**: The name of the file to read. If empty, the file name is set dynamically via the input packet. (Optional)

## Ports

- **in**: Receives input packets to read and parse data from the file.
  - If `filename` is provided: The input packet contains the format for reading the data from the file.
  - If `filename` is not provided: The input packet includes both the file name and the format.
- **out**: Returns a packet containing parsed data in the specified format when reading and parsing are successful.
- **error**: Returns errors encountered during file reading or data parsing.

## Example

```yaml
- kind: scan
  filename: /dev/stdin
  ports:
    out:
      - name: next
        port: in
```
