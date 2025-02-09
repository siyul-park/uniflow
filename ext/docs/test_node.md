# Test Node

The **Test Node** provides functionality for executing and validating developed workflows. It allows for both simple success/failure verification and more complex validation scenarios through configurable output ports.

## Specification

No additional configuration parameters are required.

## Ports

There are two ways to define ports: using a single `out` port for simple workflow execution tests, or using an `out[]` array for multiple or detailed test scenarios.

- **out**: Port that receives the result of the workflow under test.
  - Executes the connected workflow
  - Success is determined by the absence of errors from the connected workflow
  - Any error returned from the workflow indicates test failure

- **out[]**: Port that receives results from multiple workflows, where results are passed in [index, value] format according to the array index order.
  - **index**: Represents the current sequence number, starting with -1
  - **value**: Represents the value to compare against the test result
  - If the validation node finds a mismatch between the workflow execution result and the expected value, the test fails and all subsequent tests are considered failed

## Example

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
    - name: assert
      port: in
```

## Behavior

1. **Simple Test**
  - Test succeeds if the connected workflow completes without errors
  - Test fails if the workflow returns any error

2. **Extended Test**
  - The workflow connected to `out[0]` is executed first
  - Results are returned in [index, value] format and passed to the node connected to `out[1]`
  - Test succeeds only if both workflow execution and value validation pass
  - Test fails if either step fails

3. **Error Handling**
  - All errors are propagated through the test node
  - Detailed error information is provided for debugging purposes
  - Test execution is terminated immediately upon encountering any error 