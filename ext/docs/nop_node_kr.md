# NOP 노드

**NOP 노드**는 입력된 패킷을 처리하지 않고 단순히 빈 패킷으로 응답합니다. 이 노드는 데이터 처리 흐름에서 최종 단계로 사용되며, 추가적인 처리가 필요 없는 경우 불필요한 출력을 제거하는 데 유용합니다.

## 명세

- 추가적인 인자를 요구하지 않습니다.

## 포트

- **in**: 외부에서 입력된 패킷을 수신하여 빈 패킷으로 응답합니다.

## 예시

```yaml
- kind: print
  filename: /dev/stdout
  ports:
    out:
      - name: nop
        port: in

- kind: nop
  name: nop
```