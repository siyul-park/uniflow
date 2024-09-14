# Gateway Node

**The Gateway Node** provides the capability to upgrade network protocols to other protocols, primarily converting HTTP connections to WebSocket connections to support real-time data communication.

## Specification

- **protocol**: Specifies the protocol to be used. Currently, the supported protocol is `websocket`.
- **timeout**: Sets the timeout duration for the HTTP handshake. (Optional)
- **buffer**: Sets the size of the read and write buffers. (Optional)

## Ports

- **io**: Receives HTTP requests and upgrades them to WebSocket connections.
  - **method**: HTTP request method (e.g., `GET`, `POST`)
  - **scheme**: URL scheme (e.g., `http`, `https`)
  - **host**: Request host
  - **path**: Request path
  - **query**: URL query string parameters
  - **protocol**: HTTP protocol version (e.g., `HTTP/1.1`)
  - **header**: HTTP headers
  - **body**: Request body
- **in**: Sends packets through the WebSocket connection.
  - **type**: Type of WebSocket packet
  - **data**: Data of the WebSocket packet, which is analyzed and converted to the appropriate format from the raw bytes.
- **out**: Returns packets received through the WebSocket connection.
  - **type**: Type of WebSocket packet
  - **data**: Data of the WebSocket packet, which is analyzed and converted to the appropriate format from the raw bytes.
- **error**: Returns errors encountered during WebSocket upgrade or data transmission.

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
