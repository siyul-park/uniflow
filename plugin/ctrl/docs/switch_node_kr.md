# Switch 노드

**Switch 노드**는 입력된 패킷을 조건에 따라 여러 포트 중 하나로 분기하는 기능을 제공합니다. 이 노드는 지정된 조건을 평가하여 입력 데이터에 맞는 포트로 패킷을 전달하여, 복잡한 데이터 흐름 제어와 유연한
분기를 지원합니다.

## 명세

- **matches**: 패킷을 분기하기 위한 조건 목록입니다. 각 조건은 패킷을 특정 포트로 라우팅하는 규칙을 정의합니다.
    - **when**: 조건을 정의하는 표현식입니다. `Common Expression Language (CEL)`을 사용하여 컴파일 및 실행됩니다.
    - **port**: 조건이 충족될 때 패킷을 라우팅할 포트를 지정합니다.

## 포트

- **in**: 입력 패킷을 수신하고 조건에 따라 분기합니다.
- **out[*]**: 조건에 맞는 포트로 분기된 패킷을 출력합니다.
- **error**: 조건 평가 중 발생한 오류를 외부로 전달합니다.

## 예시

```yaml
- kind: switch
  matches:
    - when: "payload['status'] == 'success'"
      port: out[0]
    - when: "payload['status'] == 'error'"
      port: out[1]
  ports:
    out[0]:
      - name: success
        port: in
    out[1]:
      - name: error
        port: in
```
