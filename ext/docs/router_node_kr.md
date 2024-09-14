# Router 노드

**Router 노드**는 입력 패킷을 라우팅하여 조건에 맞는 여러 출력 포트로 전달합니다. 이 노드는 HTTP 메서드와 경로에 기반하여 패킷을 적절한 포트로 분기합니다.

## 명세

- **routes**: 패킷의 라우팅 규칙을 정의합니다. 각 규칙은 HTTP 메서드와 경로를 기반으로 패킷을 특정 포트로 전송합니다.
  - **method**: 라우팅할 HTTP 메서드 (예: `GET`, `POST`).
  - **path**: 라우팅할 HTTP 경로.
  - **port**: 규칙이 충족될 때 패킷을 전달할 포트.

## 포트

- **in**: 입력 패킷을 수신합니다.
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **path**: 요청의 경로
- **out[*]**: 라우팅 규칙에 따라 패킷을 전달합니다.
  - **method**: HTTP 메서드 (예: `GET`, `POST`)
  - **path**: 요청의 경로
  - **params**: 경로 변수
- **error**: 라우팅 중 발생한 오류를 반환합니다.

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
      path: /api/data
      port: out[0]
    - method: POST
      path: /api/submit
      port: out[1]
  ports:
    out[0]:
      - name: data
        port: in
    out[1]:
      - name: submit
        port: in
```
