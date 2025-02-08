# 워크플로우 테스트

워크플로우 테스트를 위한 테스트 패키지입니다.

## 시작하기

현재 테스트 패키지는 test runner, 테스트를 실행하고 결과를 확인할 수 있는 기능만 제공합니다. 후에 좀 더 섬세한 테스트 기능을 지원할 예정입니다.

### 테스트 노드 정의

테스트는 YAML 파일에서 정의합니다. kind에 `test`를 지정하고, 테스트 이름을 지정합니다. 여기에 테스트를 할 노드를 구성합니다.

```yaml
kind: test
name: "ping test"
request:
  method: "GET"
  url: "http://localhost:8080/ping"
  headers:
    Content-Type: "application/json"
  body:
    key: "value"
expect:
  status: 200
  body:
    status: "ok"
```

### CLI 사용

```bash
# 기본 실행
uniflow test

# 네임스페이스 지정
uniflow test --namespace=test

# 환경 변수 설정 (개별 설정)
uniflow test --env PORT=8000 --env HOST=localhost

# 스펙 파일에서 로드
uniflow test --from-specs=tests.yaml

# 값 파일에서 로드
uniflow test --from-values=values.yaml

# 특정 테스트만 실행 (정규식 지원)
uniflow test "ping.*"

# 디버그 모드로 실행
uniflow test --debug
```

### 환경 변수

환경 변수는 `--env` 플래그를 사용하여 설정할 수 있습니다. 설정된 환경 변수는 테스트 실행 시 템플릿 처리에 사용됩니다.

예시:
```yaml
kind: test
name: "port test"
request:
  method: "GET"
  url: "http://localhost:{{ .PORT }}/ping"
expect:
  status: 200
```

```bash
uniflow test --from-specs=test.yaml --env PORT=8000
```

### 디버그 모드

`--debug` 플래그를 사용하면 다음과 같은 추가 정보를 확인할 수 있습니다:
- 스펙 파일 로딩 과정
- 테스트 노드 컴파일 과정
- 환경 변수 처리 상태
- 상세한 에러 메시지

## 테스트 결과

테스트 결과는 다음과 같이 출력됩니다:

```
✓ ping test
  Duration: 123.45ms

✗ auth test
  Error: validation failed: status code mismatch: expected 200, got 401
  Duration: 234.56ms
```

## 테스트 노드 스펙

### Request 설정

- `method`: HTTP 메서드 (GET, POST, PUT, DELETE 등)
- `url`: 요청 URL
- `headers`: HTTP 헤더 (선택)
- `body`: 요청 본문 (선택)

### Expect 설정

- `status`: 예상되는 HTTP 상태 코드
- `body`: 예상되는 응답 본문 (JSON)

## 리포터 사용하기

```go
runner := testing.NewRunner()
runner.AddReporter(testing.NewTextReporter(os.Stdout))
```

## 모범 사례

1. **테스트 구조화**
   - 관련된 테스트를 TestCases로 그룹화
   - 의미 있는 테스트 이름 사용
   - 각 테스트는 하나의 동작만 검증

2. **리소스 관리**
   - Cleanup 함수로 리소스 정리
   - BeforeAll/AfterAll 훅으로 공유 리소스 관리

3. **assertion 사용**
   - 명확한 에러 메시지를 위해 적절한 assertion 메서드 선택
   - 복잡한 검증은 여러 assertion으로 분리

## 향후 계획

- 병렬 테스트 실행 지원
- 테스트 필터링 기능 강화
- HTML 리포트 생성
- 테스트 커버리지 리포트 