# **Block Node**

The **Block Node** is designed to manage complex data processing flows by grouping multiple sub-nodes together. Each
sub-node performs a specific task, and they interact to define the overall data flow, either sequentially or in
parallel.

## **Specification**

- **specs**: Defines the list of sub-nodes to be executed. Each sub-node handles a specific data processing task, and
  they interact to form the complete data flow. The sub-nodes can be executed in sequence or concurrently, depending on
  the configuration.
- **inbound**: Specifies the input ports for receiving external data. These ports serve as the entry points for the data
  processing flow.
- **outbound**: Specifies the output ports for sending the processed data to external systems. These ports deliver the
  final result of the data processing flow.

## **Ports**

All ports are determined dynamically at runtime, and they are automatically assigned based on the connections between
sub-nodes.

## **Example**

```yaml
- kind: block
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
      ports:
        out:
          - name: $1
            port: in
    - kind: http
      url: https://api.example.com/data
      ports:
        out:
          - name: $2
            port: in
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return args.body;
        }
  inbounds:
    in:
      - name: $0
        port: in
  outbounds:
    out:
      - name: $2
        port: out
```
