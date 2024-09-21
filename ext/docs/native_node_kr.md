# Native 노드

**Native 노드**는 시스템 내부에서 함수 호출 작업을 수행하는 노드입니다. 이 노드는 `opcode`를 기반으로 시스템 호출을 처리하며, 입력 패킷을 함수에 전달하여 실행하고, 그 결과를 반환합니다.

## 명세

- **opcode**: 호출할 시스템 작업을 식별하는 문자열입니다. 지정된 함수와 연관되며, 이 값을 통해 노드의 동작을 결정합니다.

## 포트

- **in**: 입력 패킷을 수신하여 지정된 함수 호출에 필요한 인수로 변환합니다. 패킷의 페이로드는 함수의 매개변수와 일치하도록 조정됩니다.
- **out**: 함수 호출의 결과를 반환합니다. 함수의 반환값이 여러 개일 경우, 배열로 패킷을 생성하여 출력합니다.
- **error**: 함수 호출 중 발생한 오류를 반환합니다.

## 예시

```yaml
- kind: snippet
  language: cel
  code: 'has(self.body) ? self.body : null'
  ports:
    out:
      - name: nodes_create
        port: in

- kind: native
  name: nodes_create
  opcode: nodes.create
```
