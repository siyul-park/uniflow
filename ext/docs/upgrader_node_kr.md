# Upgrade 노드

**Upgrade 노드**는 네트워크 프로토콜을 다른 프로토콜로 업그레이드하는 기능을 제공하며, 주로 HTTP 연결을 WebSocket 연결로 변환하여 실시간 데이터 통신을 지원합니다.

## 명세

- **protocol**: 사용할 프로토콜을 지정합니다. 현재 지원되는 프로토콜은 `websocket`입니다.
- **timeout**: HTTP 핸드셰이크의 타임아웃 기간을 설정합니다. (선택 사항)
- **buffer**: 읽기 및 쓰기 버퍼의 크기를 설정합니다. (선택 사항)

## 포트

- **io**: HTTP 요청을 수신하여 WebSocket 연결로 업그레이드합니다.
  - **method**: HTTP 요청 메서드 (예: `GET`, `POST`)
  - **scheme**: URL의 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
- **in**: WebSocket 연결을 통해 패킷을 송신합니다.
  - **type**: WebSocket 패킷의 타입
  - **data**: WebSocket 패킷의 데이터로, 원본 바이트를 분석하여 적절한 형태로 변환됩니다.
- **out**: WebSocket 연결을 통해 수신된 패킷을 반환합니다.
  - **type**: WebSocket 패킷의 타입
  - **data**: WebSocket 패킷의 데이터로, 원본 바이트를 분석하여 적절한 형태로 변환됩니다.
- **error**: WebSocket 업그레이드 또는 데이터 송수신 중 발생한 오류를 반환합니다.

## 예시

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
      - name: upgrader
        port: io
      - name: proxy
        port: io

- kind: upgrader
  name: upgrader
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
      - name: upgrader
        port: in
```
