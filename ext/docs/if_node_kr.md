# If 노드

**If 노드**는 주어진 조건에 따라 패킷을 분기하여 두 개의 경로 중 하나로 전달하는 기능을 제공합니다. 이 노드는 조건을 평가하고, 그 결과에 따라 다른 데이터 흐름을 실행할 수 있습니다.

## 명세

- **when**: 조건을 정의하는 표현식입니다. 이 표현식은 `Common Expression Language (CEL)`로 작성되어 컴파일되고 실행됩니다.

## 포트

- **in**: 입력 패킷을 수신하고 조건을 평가하여 분기합니다.
- **out[0]**: 조건이 참일 때 패킷을 전달합니다.
- **out[1]**: 조건이 거짓일 때 패킷을 전달합니다.
- **error**: 조건 평가 중 발생한 오류를 외부로 전달합니다.

## 예시

```yaml
- kind: if
  when: "self.count > 10"
  ports:
    out[0]:
      - name: true_path
        port: out
    out[1]:
      - name: false_path
        port: out
```
