# Signal 노드

**Signal 노드**는 신호 채널을 수신하여 받은 신호를 패킷으로 전달하는 노드입니다. 이 노드는 실시간 이벤트 또는 시스템 수준의 신호를 워크플로우에서 처리하는 데 유용합니다.

## 명세

- **topic**: 청취할 시스템 작업을 식별하는 문자열입니다. 지정된 함수와 연관되며 노드의 동작을 결정합니다.
-

## 포트

- **out**: 수신된 신호 데이터를 포함한 패킷을 전송합니다.

## 예시

```yaml
- kind: signal
  topic: specs
  ports:
    out:
      - name: next
        port: in
```