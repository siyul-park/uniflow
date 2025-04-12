# Assert Node

The **Assert Node** compares the expected test conditions with the actual execution results in a workflow. If they match, the test succeeds; if they differ, it triggers an error and fails. It is typically used in conjunction with Test nodes to verify whether tests are executed correctly. Additional port connections can be configured when needed to perform more complex test validations.

## Specification

- **expect**: Defines the expected result value. Written in `Common Expression Language (CEL)`, it is compared with the actual result to check if it matches the expectation.
- **target**: Specifies the target to validate.
    - **name**: Name of the target node
    - **port**: Output port of the target node
    - Note: If this field does not exist, it uses the frame received immediately after. If it exists, it searches for a frame matching the conditions and uses it. In this case, if the frame cannot be found, it **considers it an error and stops the test**.

## Ports

- **in**: Receives data to validate in the format [value, index]
- **out**: When validation succeeds, passes the current frame and index to the next node in the format [value, index]

## Examples

```yaml
- kind: test
  name: non_target_test
  ports:
    out[0]:
      - name: snippet
        port: in
    out[1]:
      - name: assert
        port: in

- kind: snippet
  name: snippet
  language: json
  code: 42

- kind: assert
  name: assert
  expect: self == 42
```

```yaml
- kind: test
  name: target_test
  ports:
    out[0]:
      - name: first
        port: in
    out[1]:
      - name: assert
        port: in

- kind: snippet
  name: first
  language: json
  code: 1
  ports:
    out:
      - name: second
        port: in

- kind: snippet
  name: second
  language: json
  code: 2

- kind: assert
  name: assert
  expect: self == 1
  target:
    name: first
    port: out
```
