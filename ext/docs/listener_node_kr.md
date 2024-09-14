# Listener 노드

**Listener 노드**는 지정된 프로토콜과 포트에서 네트워크 요청을 수신하고 처리하는 기능을 제공합니다. 주로 HTTP 서버 역할을 하며, 클라이언트 요청을 처리하고 적절한 응답을 반환합니다.

## 명세

- **protocol**: 처리할 프로토콜을 지정합니다. 현재는 `http` 프로토콜을 지원합니다.
- **host**: 서버의 호스트 주소를 지정합니다. (선택 사항)
- **port**: 서버가 리슨할 포트 번호를 설정합니다.
- **cert**: HTTPS를 사용할 때 TLS 인증서를 설정합니다. (선택 사항)
- **key**: HTTPS를 사용할 때 TLS 비밀 키를 설정합니다. (선택 사항)

## 포트

- **out**: HTTP 연결을 통해 수신된 패킷을 반환합니다.
  - **method**: HTTP 요청 메서드 (예: `GET`, `POST`)
  - **scheme**: URL 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
  - **status**: HTTP 상태 코드

## 예시

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
