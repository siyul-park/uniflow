# Session Node

**The Session Node** is responsible for storing and managing information used in a process. It maintains session information and continuously manages the session state while the process is active. This allows for effective management of session-based data flows by querying and processing session information.

## Specification

- Additional arguments are not required.

## Ports

- **io**: Receives input packets to store as session information.
- **in**: Merges the input packet with stored session information to create and execute a new subprocess. The merged information and input packet are output as a single packet.
- **out**: Outputs the combined session information and input packet.

## Example

```yaml
- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return {
        uid: args.uid,
      };
    }
  ports:
    out:
      - name: session
        port: io

- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return {
        uid: args.uid,
      };
    }
  ports:
    out:
      - name: session
        port: in

- kind: session
  name: session
  ports:
    out:
      - name: next
        port: out

- kind: if
  name: next
  when: "self[0].uid == self[1].uid"
```
