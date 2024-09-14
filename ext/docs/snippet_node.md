# Snippet Node

**The Snippet Node** executes code snippets written in various programming languages to process input packets and produce output. This node allows for flexible application of complex data processing logic and provides diverse data handling capabilities through dynamic code execution.

## Specification

- **language**: Specifies the programming language in which the code snippet is written. (e.g., `text`, `json`, `yaml`, `cel`, `javascript`, `typescript`)
- **code**: Provides the code snippet to be executed.

## Ports

- **in**: Receives the input packet and executes it using the provided code.
- **out**: Outputs the result of the code execution.
- **error**: Returns any errors encountered during code execution.

## Example

```yaml
- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return {
        body: {
          error: args.error()
        },
        status: 400
      };
    }
```
