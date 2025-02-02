# For 노드

**For 노드**는 입력된 패킷을 여러 하위 패킷으로 분리하여 반복적으로 처리할 수 있는 기능을 제공합니다. 반복 작업이 필요한 데이터 처리 흐름에서 유용하며, 각각의 하위 패킷을 처리한 후 그 결과를 통합하여
반환합니다.

## 명세

- 추가 인자는 요구되지 않습니다.

## 포트

- **in**: 외부에서 입력된 패킷을 수신하여 반복 작업을 시작합니다. 입력이 배열일 경우 각 요소가 하위 패킷으로 분리되어 개별적으로 처리됩니다. 배열이 아닌 경우 한 번만 반복됩니다.
- **out[0]**: 분리된 하위 패킷을 첫 번째 출력 포트로 전달합니다.
- **out[1]**: 모든 하위 패킷의 처리 결과를 모아 두 번째 출력 포트로 전달합니다.
- **error**: 처리 중 발생한 오류를 외부로 전달합니다.

## 예시

```yaml
- kind: for
  ports:
    out[0]:
      - name: next
        port: out
    out[1]:
      - name: done
        port: out
```
