# WebSocket 노드

**WebSocket 노드**는 WebSocket 클라이언트 연결을 설정하고, WebSocket 프로토콜을 통해 메시지를 송수신하는 기능을 제공합니다. 이 노드는 WebSocket 서버와의 연결을 관리하며, 데이터 전송 및 수신을 처리합니다.

## 명세

- **url**: WebSocket 서버의 URL을 정의합니다. (선택 사항)
- **timeout**: WebSocket 핸드쉐이크 타임아웃을 설정합니다. (선택 사항)

## 포트

- **io**: WebSocket 연결을 설정합니다.
  - **scheme**: URL의 스킴 (예: `ws`, `wss`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
- **in**: WebSocket 연결을 통해 패킷을 송신합니다.
  - **type**: WebSocket 패킷의 타입
  - **data**: WebSocket 패킷의 데이터로, 원본 바이트를 분석하여 적절한 형태로 변환됩니다.
- **out**: WebSocket 연결을 통해 수신된 패킷을 반환합니다.
  - **type**: WebSocket 패킷의 타입
  - **data**: WebSocket 패킷의 데이터로, 원본 바이트를 분석하여 적절한 형태로 변환됩니다.
- **error**: WebSocket 연결 또는 데이터 송수신 중 발생한 오류를 반환합니다.

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