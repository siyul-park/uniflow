# HTTP 노드

**HTTP 노드**는 HTTP 클라이언트 요청을 처리하고, 요청을 생성하여 웹 서비스와 통신한 후 응답을 반환하는 기능을 제공합니다.

## 명세

- **url**: 요청을 보낼 대상 URL을 지정합니다. (선택 사항)
- **timeout**: HTTP 요청의 타임아웃 기간을 설정합니다. (선택 사항)

## 포트

- **in**: HTTP 요청을 수신합니다.
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **scheme**: URL의 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
  - **status**: HTTP 상태 코드
- **out**: HTTP 응답을 반환합니다.
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **scheme**: URL의 스킴 (예: `http`, `https`)
  - **host**: 요청의 호스트
  - **path**: 요청의 경로
  - **query**: URL 쿼리 문자열 파라미터
  - **protocol**: HTTP 프로토콜 버전 (예: `HTTP/1.1`)
  - **header**: HTTP 헤더
  - **body**: 요청 본문
  - **status**: HTTP 상태 코드
- **error**: 요청 처리 중 발생한 오류를 반환합니다.

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

- kind: http
  name: proxy
  url: https://example.com/
```
