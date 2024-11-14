# Wait Node

**The Wait Node** introduces a delay in packet processing, allowing for timed pauses in workflows. This node is useful for pacing operations or waiting for external conditions before continuing to process data.

## Specification

- **interval**: Defines the duration (in milliseconds or a Go `time.Duration` format) for which the node will delay before passing the input packet to the output.

## Ports

- **in**: Receives the input packet and initiates a delay.
- **out**: Outputs the original input packet after the specified delay.

## Example

```yaml
- kind: wait
  interval: 2000 # Delay of 2 seconds
```
