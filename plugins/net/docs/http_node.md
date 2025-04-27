# HTTP Node

**The HTTP Node** processes HTTP client requests, generates requests to communicate with web services, and returns
responses.

## Specification

- **url**: Specifies the target URL to send the request to. (Optional)
- **timeout**: Sets the timeout duration for the HTTP request. (Optional)

## Ports

- **in**: Receives HTTP requests.
    - **method**: HTTP method (e.g., `GET`, `POST`)
    - **scheme**: URL scheme (e.g., `http`, `https`)
    - **host**: Request host
    - **path**: Request path
    - **query**: URL query string parameters
    - **protocol**: HTTP protocol version (e.g., `HTTP/1.1`)
    - **header**: HTTP headers
    - **body**: Request body
    - **status**: HTTP status code
- **out**: Returns HTTP responses.
    - **method**: HTTP method (e.g., `GET`, `POST`)
    - **scheme**: URL scheme (e.g., `http`, `https`)
    - **host**: Request host
    - **path**: Request path
    - **query**: URL query string parameters
    - **protocol**: HTTP protocol version (e.g., `HTTP/1.1`)
    - **header**: HTTP headers
    - **body**: Request body
    - **status**: HTTP status code
- **error**: Returns errors encountered during request processing.

## Example

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: proxy
        port: in

- kind: http
  name: proxy
  url: https://example.com/
```
