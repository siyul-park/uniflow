# Fork 노드

**Fork 노드**는 데이터 처리 흐름을 비동기적으로 분기하여 별도의 프로세스에서 처리할 수 있는 기능을 제공합니다. 이를 통해 병렬 처리가 가능해지며, 메인 흐름을 차단하지 않고 독립적인 작업을 수행할 수
있습니다.

## 명세

- 추가 인자를 요구하지 않습니다.

## 포트

- **in**: 입력된 패킷을 새로운 프로세스로 전달하고, 빈 패킷을 반환합니다.
- **out**: 비동기적으로 처리된 결과를 출력합니다.
- **error**: 처리 과정에서 발생한 오류를 외부로 전달합니다.

## 예시

```yaml
- kind: fork
  ports:
    out:
      - name: next
        port: out
```
