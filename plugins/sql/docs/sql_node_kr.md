# SQL 노드

**SQL 노드**는 관계형 데이터베이스와 상호작용하여 SQL 쿼리를 실행하고 데이터를 처리하는 기능을 제공합니다. 이 노드는 SQL 쿼리를 데이터베이스에서 실행하고, 결과를 패킷으로 반환합니다.

## 명세

- **driver**: 데이터베이스 드라이버의 이름입니다. 예를 들어, `"sqlite3"`, `"postgres"` 등이 될 수 있습니다.
- **source**: 데이터베이스 연결 문자열입니다. 드라이버에 따라 적절한 형식으로 제공되어야 합니다. (선택 사항)
- **isolation**: 트랜잭션의 격리 수준을 설정합니다. 기본값은 `0`입니다. (선택 사항)

## 포트

- **in**: SQL 쿼리와 파라미터를 받아서 데이터베이스에 요청을 보냅니다.
- **out**: 쿼리 실행 결과를 포함하는 패킷을 반환합니다.
- **error**: 쿼리 실행 중 발생한 오류를 반환합니다.

## 예시

```yaml
- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return [
        'INSERT INTO USERS(name) VALUES (?)',
        ["foo", "bar"]
      ];
    }
  ports:
    out:
      - name: sql
        port: in

- kind: sql
  name: sql
  driver: sqlite3
  source: file::memory:?cache=shared
  ports:
    out:
      - name: next
        port: in
```
