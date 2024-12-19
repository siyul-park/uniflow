# Step Node

**The Step Node** systematically manages complex data processing flows and executes multiple sub-nodes sequentially.
This allows you to organize data processing tasks clearly and efficiently.

## Specification

- **specs**: Defines the list of sub-nodes to be executed. Each sub-node handles a specific step in the data processing flow and is executed in sequence.

## Ports

- **in**: Passes the input packet to the first sub-node.
- **out**: Outputs the result processed by the last sub-node to the external environment.
- **error**: Sends any errors encountered by the sub-nodes to the external environment.

## Example

```yaml
- kind: step
  specs:
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            payload: {
              method: 'GET',
              body: args
            }
          };
        }
    - kind: http
      url: https://api.example.com/data
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return args.body;
        }
```
