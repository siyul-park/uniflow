# WebSocket Node

**The WebSocket Node** provides functionality for setting up WebSocket client connections and handling message transmission and reception using the WebSocket protocol. This node manages connections to a WebSocket server and processes data exchanges.

## Specification

- **url**: Defines the URL of the WebSocket server. (Optional)
- **timeout**: Sets the WebSocket handshake timeout period. (Optional)

## Ports

- **io**: Sets up the WebSocket connection.
  - **scheme**: Scheme of the URL (e.g., `ws`, `wss`)
  - **host**: Host of the request
  - **path**: Path of the request
- **in**: Sends packets through the WebSocket connection.
  - **type**: Type of the WebSocket packet
  - **data**: Data of the WebSocket packet, analyzed and converted from raw bytes to an appropriate format
- **out**: Returns packets received through the WebSocket connection.
  - **type**: Type of the WebSocket packet
  - **data**: Data of the WebSocket packet, analyzed and converted from raw bytes to an appropriate format
- **error**: Returns errors encountered during WebSocket connection or data transmission and reception

## Example

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: GET
      path: /ws
      port: out[0]
  ports:
    out[0]:
      - name: gateway
        port: io
      - name: proxy
        port: io

- kind: gateway
  name: gateway
  protocol: websocket
  ports:
    out:
      - name: proxy
        port: in

- kind: websocket
  name: proxy
  url: wss://echo.websocket.org/
  ports:
    out:
      - name: gateway
        port: in
```
