# Test Node

**Test Node** provides functionality for testing developed workflows. When executed with the `uniflow test` command, the system recognizes and executes this node. supporting both simple success/failure verification and complex validation scenarios through configurable output ports.

## Specification

- No additional configuration parameters are required.

## Ports

Connect the workflow to be tested to the out[0] port and validate its results through out[1]. For simple execution validation, out[1] can be omitted.

- **out[0]**: Executes the workflow and receives its response.
  - If no error occurs in the connected workflow, the test is considered successful
  - If the workflow returns an error, the test is considered failed

- **out[1]**: Results from out[0] execution are passed to out[1] in [payload, index] format.
  - **index**: Indicates where the value is positioned in frames. Starts with -1.
  - **value**: Represents the execution result of the workflow.
  - If out[1] returns an error, the test is considered failed

## Examples

```yaml
- kind: test
  name: test
  ports:
    out:
      - name: snippet
        port: in

- kind: snippet
  name: first
  language: json
  code: 1
```
