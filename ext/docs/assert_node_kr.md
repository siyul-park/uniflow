# Assert 노드

**Assert 노드**는 워크플로우 실행 중 특정 조건을 검증하고, 조건이 충족되지 않으면 오류를 발생시키는 노드입니다. 일반적으로 Test 노드와 함께 작동하여 테스트가 정상적으로 수행되었는지 검증하며, 다음과 같은 구조로 연결됩니다:
```
Test 노드 -> 대상 노드 -> Assert 노드
```
검증에 성공할 경우 특별한 추가 동작 없이 실행을 완료하고, 실패할 경우 오류를 발생시키게 됩니다.

## 명세

- **expect**: 예상되는 결과 값을 지정합니다. 이 값과 실제 결과가 비교됩니다.
- **mode**: 비교 방식을 지정합니다. 다음 중 하나를 선택할 수 있습니다:
  - `exact`: 정확한 값 비교 (기본값)
  - `type`: 값의 타입만 비교
  - `exists`: 값이 존재하는지만 확인
- **msg**: (선택 사항) 검증 실패 시 표시할 오류 메시지를 지정합니다. 지정하지 않으면 기본 오류 메시지가 표시됩니다.

## 포트

- **in**: 검증할 입력 데이터를 받습니다. 일반적으로 Test 노드를 통해 실행된 노드의 출력이 이 포트로 전달됩니다.

## 예시

다음은 간단한 숫자 연산 결과를 검증하는 예시입니다:

```yaml
- kind: test
  name: test-multiply
  ports:
    out[0]:
      - name: multiply
        port: in
    out[1]:
      - name: assert-result
        port: in

- kind: snippet
  name: multiply
  spec:
    language: javascript
    code: export default function(x) { return x * 2; }
  ports:
    out:
      - name: test-multiply
        port: in

- kind: assert
  name: assert-result
  ports:
    in:
      - name: test-multiply
        port: out
  spec:
    expect: 10
    msg: "곱셈 결과는 10이어야 합니다"
```
