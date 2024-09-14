# Print Node

**The Print Node** provides functionality for writing input data to a file in a specified format. This node uses a format to save data to a file, and the file name can be either a fixed value or dynamically set through the input packet.

## Specification

- **filename**: The name of the file where data will be recorded. If the filename is an empty string, the file name is dynamically set through the input packet. (Optional)

## Ports

- **in**: Receives input packets to be written to a file.
  - If `filename` is provided: The packet includes the format and optionally arguments for writing to the file.
  - If `filename` is not provided: The packet includes the file name, format, and optionally arguments for writing to the file.
- **out**: Returns a packet containing the number of bytes successfully written to the file.
- **error**: Returns any errors encountered during file writing.

## Example

```yaml
- kind: print
  filename: /dev/stdout
  ports:
    out:
      - name: next
        port: in
```
