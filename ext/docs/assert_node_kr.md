# Assert 노드

**Assert 노드**는 워크플로우 테스트 성공 조건과 실제 실행 결과를 비교하여, 동일하면 성공을, 다르다면 오류를 발생시켜 실패하도록 만듭니다. 일반적으로 Test 노드와 함께 구성하여 테스트가 정상적으로 수행되었는지 검증할 때 사용되며, 필요하다면 포트 연결을 추가로 설정하여 복잡한 테스트 검증을 할 수 있습니다.

## 명세

- **expect**: 예상되는 결과 값을 정의합니다. `Common Expression Language (CEL)`로 작성하며, 실제 결과와 비교되어 예상과 일치하는지 검사합니다.
- **target**: 검증할 대상을 지정합니다. 
    - **name**: 대상 노드의 이름
    - **port**: 대상 노드의 출력 포트
    - 주의: 해당 필드가 존재하지 않을 경우 직후에 전달받은 프레임을 사용하며, 존재하면 조건에 맞는 프레임을 검색하여 사용합니다. 이 때, 해당 프레임을 찾을 수 없다면 **오류로 판단하고 테스트를 중단합니다.**

## 포트

- **in**: [payload, index] 형식으로 검증할 데이터를 전달합니다.
- **out**: 검증 성공 시 현재 프레임과 페이로드와 인덱스를 [payload, index] 형식으로 다음 노드에 전달합니다.

## 예시

```yaml
- kind: test
  name: non_target_test
  ports:
    out[0]:
      - name: snippet
        port: in
    out[1]:
      - name: assert
        port: in

- kind: snippet
  name: snippet
  language: json
  code: 42

- kind: assert
  name: assert
  expect: self == 42
```

```yaml
- kind: test
  name: target_test
  ports:
    out[0]:
      - name: first
        port: in
    out[1]:
      - name: assert
        port: in

- kind: snippet
  name: first
  language: json
  code: 1
  ports:
    out:
      - name: second
        port: in

- kind: snippet
  name: second
  language: json
  code: 2

- kind: assert
  name: assert
  expect: self == 1
  target:
    name: first
    port: out
```
