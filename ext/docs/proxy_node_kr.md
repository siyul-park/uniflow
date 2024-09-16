# Proxy 노드

**Proxy 노드**는 HTTP 요청을 다른 서버로 프록시하여 중계하고, 그 응답을 반환하는 기능을 제공합니다.

## 명세

- **urls**: 프록시할 대상 서버의 URL 목록을 지정합니다. 요청은 이 목록에서 라운드 로빈 방식으로 선택된 서버로 전달됩니다.

## 포트

- **in**: HTTP 요청을 수신하는 포트입니다. 다음 필드를 포함합니다:
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **scheme**: URL의 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
  - **status**: HTTP 상태 코드

- **out**: 프록시된 서버의 응답을 반환하는 포트입니다. 다음 필드를 포함합니다:
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **scheme**: URL의 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
  - **status**: HTTP 상태 코드

- **error**: 오류가 발생했을 때 에러를 반환하는 포트입니다. (예: 네트워크 장애, 잘못된 URL)

## 예시

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
