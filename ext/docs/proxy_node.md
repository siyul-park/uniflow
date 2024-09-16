# Proxy Node

**The Proxy Node** acts as an HTTP proxy, forwarding requests to other servers and returning their responses. It can be used for load balancing, API gateway functionality, or caching.

## Specification

- **urls**: A list of target server URLs to proxy requests to. Requests are forwarded to these URLs using a round-robin approach. (Required)

## Ports

- **in**: The port that receives HTTP requests. The following fields are included:
  - **method**: The HTTP method (e.g., `GET`, `POST`)
  - **scheme**: The URL scheme (e.g., `http`, `https`)
  - **host**: The request's host (e.g., `example.com`)
  - **path**: The request's path (e.g., `/api/v1/resource`)
  - **query**: URL query string parameters (e.g., `?key=value`)
  - **protocol**: The HTTP protocol version (e.g., `HTTP/1.1`)
  - **header**: HTTP headers (e.g., `Content-Type: application/json`)
  - **body**: The request body (e.g., JSON, XML, text)
  - **status**: The HTTP status code

- **out**: The port that returns the response from the proxied server. The following fields are included:
  - **method**: The HTTP method (e.g., `GET`, `POST`)
  - **scheme**: The URL scheme (e.g., `http`, `https`)
  - **host**: The request's host
  - **path**: The request's path
  - **query**: URL query string parameters
  - **protocol**: The HTTP protocol version (e.g., `HTTP/1.1`)
  - **header**: The HTTP headers of the response
  - **body**: The response body
  - **status**: The HTTP status code of the response

- **error**: The port that returns any errors encountered during the request (e.g., network failure, invalid URL).

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

- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```
