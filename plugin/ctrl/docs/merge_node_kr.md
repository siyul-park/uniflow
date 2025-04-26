# Merge 노드

**Merge 노드**는 여러 개의 입력 패킷을 하나로 통합하는 기능을 제공합니다. 이 노드는 다양한 입력 소스에서 데이터를 수집하여 하나의 패킷으로 합치고 처리하거나 전달할 때 유용합니다.

## 명세

- 추가적인 인자를 요구하지 않습니다.

## 포트

- **in[*]**: 여러 입력 패킷을 수신합니다. 각 입력 포트는 별도의 데이터 소스에서 전달되는 패킷을 받으며, 다양한 형식을 지원합니다.
- **out**: 입력된 패킷들을 병합한 결과를 하나의 패킷으로 출력합니다.
- **error**: 병합 과정에서 발생한 오류를 전달합니다.

## 예시

```yaml
- kind: snippet
  language: json
  code: 0
  ports:
    out:
      - name: merge
        port: in[0]

- kind: snippet
  language: json
  code: 1
  ports:
    out:
      - name: merge
        port: in[1]

- kind: merge
  name: merge
  ports:
    out:
      - name: next
        port: out
```
