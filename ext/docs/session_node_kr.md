# Session 노드

**Session 노드**는 프로세스에서 사용되는 정보를 저장하고 관리하는 노드입니다. 이 노드는 세션 정보를 유지하며, 프로세스가 활성 상태인 동안 세션 상태를 지속적으로 관리합니다. 이를 통해 세션 정보를 조회하고 처리하여, 세션 기반의 데이터 흐름을 효과적으로 관리할 수 있습니다.

## 명세

- 추가적인 인자를 요구하지 않습니다.

## 포트

- **io**: 외부에서 입력된 패킷을 세션 정보로 저장합니다.
- **in**: 입력된 패킷을 저장된 세션 정보와 병합하여 새로운 하위 프로세스를 생성하고 실행합니다. 정보와 입력 패킷을 병합하여 하나의 패킷으로 출력합니다.
- **out**: 저장된 세션 정보와 입력 패킷을 병합하여 출력합니다.

## 예시

```yaml
- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return {
        uid: args.uid,
      };
    }
  ports:
    out:
      - name: session
        port: io

- kind: snippet
  language: javascript
  code: |
    export default function (args) {
      return {
        uid: args.uid,
      };
    }
  ports:
    out:
      - name: session
        port: in

- kind: session
  name: session
  ports:
    out:
      - name: next
        port: out

- kind: if
  name: next
  when: "self[0].uid == self[1].uid"
```
