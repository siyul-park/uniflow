# Reduce 노드

**Reduce 노드**는 입력 데이터를 반복적으로 연산하여 하나의 출력 값을 생성하는 기능을 제공합니다. 이 노드는 데이터 집계나 변환 작업에 유용합니다.

## 명세

- **action**: 두 개의 입력 값을 받아 하나의 출력 값을 반환하는 연산을 정의합니다. 이 연산은 `Common Expression Language (CEL)`로 작성되며, 데이터를 누적하여 처리합니다. (필수)
- **init**: 초기 값을 설정합니다 (선택 사항).

## 포트

- **in**: 축소 연산을 위한 입력 데이터를 받는 포트입니다. 누적값이 첫 번째 인자로, 현재 값이 두 번째 인자로 전달됩니다.
- **out**: 축소 연산의 결과를 출력하는 포트입니다.
- **error**: 연산 실행 중 발생한 오류를 반환합니다.

## 예시

```yaml
- kind: reduce
  action: "self[0] + self[1]"
  init: 0
  ports:
    out:
      - name: result
        port: out
``` 