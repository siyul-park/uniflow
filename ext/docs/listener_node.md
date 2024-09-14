# Listener Node

**The Listener Node** provides the functionality to receive and process network requests on a specified protocol and port. It primarily acts as an HTTP server, handling client requests and returning appropriate responses.

## Specification

- **protocol**: Specifies the protocol to handle. Currently, only the `http` protocol is supported.
- **host**: Specifies the host address of the server. (Optional)
- **port**: Sets the port number on which the server will listen.
- **cert**: Sets the TLS certificate for HTTPS use. (Optional)
- **key**: Sets the TLS private key for HTTPS use. (Optional)

## Ports

- **out**: Returns packets received via the HTTP connection.
  - **method**: HTTP request method (e.g., `GET`, `POST`)
  - **scheme**: URL scheme (e.g., `http`, `https`)
  - **host**: Request host
  - **path**: Request path
  - **query**: URL query string parameters
  - **protocol**: HTTP protocol version (e.g., `HTTP/1.1`)
  - **header**: HTTP headers
  - **body**: Request body
  - **status**: HTTP status code

## Example

```yaml
kind: listener
spec:
  protocol: http
  host: "localhost"
  port: 8080
  cert: |
    -----BEGIN CERTIFICATE-----
    [certificate data]
    -----END CERTIFICATE-----
  key: |
    -----BEGIN PRIVATE KEY-----
    [key data]
    -----END PRIVATE KEY-----
```
