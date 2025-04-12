# Test Node

**Test Node** provides functionality for executing and validating developed workflows. It allows for both simple success/failure verification and more complex validation scenarios through configurable output ports.

## Specification

- No additional configuration parameters are required.

## Ports

Connect the workflow to be tested to the out[0] port and validate its results through out[1]. For simple execution validation, out[1] can be omitted.

- **out[0]**: Executes the workflow and receives its response.
  - Executes the connected workflow
  - Success is determined by the absence of errors from the connected workflow
  - Any error returned from the workflow indicates test failure

- **out[1]**: Results from out[0] execution are passed to out[1] in [value, index] format.
  - **index**: Indicates where the value is positioned in frames. Starts with -1.
  - **value**: Represents the execution result of the workflow.
  - Test fails if out[1] returns an error

## Examples

### Simple Test Configuration
```yaml
kind: test
ports:
  out:
    - name: sub
      port: in
```

### Extended Test Configuration
```yaml
kind: test
ports:
  out[0]:
    - name: sub
      port: in
  out[1]:
    - name: require
      port: in
```