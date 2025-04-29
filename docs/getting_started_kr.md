# 🚀 시작하기

이 가이드는 [명령줄 인터페이스(CLI)](../cmd/README_kr.md)의 설치, 설정, 그리고 워크플로우 관리 방법을 쉽게 따라 할 수 있도록 설명합니다. 설치 과정부터 워크플로우의 제어 및 설정 방법까지, 필요한 모든 단계를 다룹니다.

## 소스에서 설치하기

먼저 [기본 확장 기능](../ext/README_kr.md)과 함께 제공되는 [CLI](../cmd/README_kr.md)를 설정해야 합니다. 시작하기 전에, 시스템에 [Go 1.23](https://go.dev/doc/install) 이상의 버전이 설치되어 있는지 확인하세요.

### 리포지토리 클론

소스 코드를 다운로드하려면, 터미널에서 아래 명령어를 입력하세요:

```sh
git clone https://github.com/siyul-park/uniflow
```

다운로드한 폴더로 이동합니다:

```sh
cd uniflow
```

### 의존성 설치 및 빌드

필요한 의존성을 설치하고 프로젝트를 빌드하려면, 아래 명령어를 실행하세요:

```sh
make init
make build
```

빌드가 완료되면 `dist` 폴더에 실행 파일이 생성됩니다.

### 설정

환경 변수 또는 설정 파일(`.toml`, `.yaml`, `.json`, `.hjson`, `.env`)을 사용하여 설정을 구성할 수 있습니다. 설정 파일의 경로는 `UNIFLOW_CONFIG` 환경 변수로 지정하며, 지정하지 않으면 기본값인 `.uniflow.toml` 파일이 사용됩니다.

```bash
export UNIFLOW_CONFIG=./config/uniflow.toml
```

설정 파일에서 정의할 수 있는 주요 항목은 다음과 같습니다:

```toml
[database]
url = "memory://"

[collection]
specs = "specs"
values = "values"

[language]
default = "cel"

[[plugins]]
path = "./dist/cel.so"
config.extensions = ["encoders", "math", "lists", "sets", "strings"]

[[plugins]]
path = "./dist/ecmascript.so"

[[plugins]]
path = "./dist/mongodb.so"

[[plugins]]
path = "./dist/ctrl.so"

[[plugins]]
path = "./dist/net.so"

[[plugins]]
path = "./dist/testing.so"
```

환경 변수도 자동으로 로드되며, 환경 변수는 `UNIFLOW_` 접두어를 사용합니다. 예를 들어, 다음과 같이 설정할 수 있습니다:

```env
UNIFLOW_DATABASE_URL=memory://
UNIFLOW_COLLECTION_SPECS=specs
UNIFLOW_COLLECTION_VALUES=values
UNIFLOW_LANGUAGE_DEFAULT=cel
```

만약 [MongoDB](https://www.mongodb.com/)를 사용하는 경우, 리소스의 변경 사항을 실시간으로 추적하려면 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 이를 위해서는 [복제 세트](https://www.mongodb.com/docs/manual/replication/) 구성이 필요합니다.

## 예제 실행

다음은 HTTP 요청 처리 예제인 [ping.yaml](./examples/ping.yaml)을 실행하는 방법입니다:

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: GET
      path: /ping
      port: out[0]
  ports:
    out[0]:
      - name: pong
        port: in

- kind: snippet
  name: pong
  language: text
  code: pong
```

다음 명령어로 워크플로우를 실행합니다:

```sh
uniflow start --from-specs example/ping.yaml
```

정상 작동 여부를 확인하려면 HTTP 엔드포인트를 호출하세요:

```sh
curl localhost:8000/ping
pong#
```

## 지원하는 명령어

`uniflow`는 워크플로우의 런타임 환경을 시작하고 관리하는 데 사용되는 다양한 명령어를 제공합니다.

### Start 명령어

`start` 명령어는 지정된 네임스페이스 내의 모든 노드 명세를 실행합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow start --namespace default
```

네임스페이스가 비어 있을 경우, `--from-specs` 플래그를 사용해 초기 노드 명세를 제공할 수 있습니다.

```sh
./dist/uniflow start --namespace default --from-specs examples/specs.yaml
```

초기 변수 파일은 `--from-values` 플래그로 설정할 수 있습니다.

```sh
./dist/uniflow start --namespace default --from-values examples/values.yaml
```

환경 변수는 `--env` 플래그로 지정할 수 있습니다.

```sh
./dist/uniflow start --namespace default --env DATABASE_URL=mongodb://localhost:27017 --env DATABASE_NAME=mydb
```

### Test 명령어

`test` 명령어는 지정된 네임스페이스에서 워크플로우 테스트를 실행합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow test --namespace default
```

특정 테스트만 실행하려면 정규식을 사용하여 필터링할 수 있습니다.

```sh
./dist/uniflow test ".*/my_test" --namespace default
```

네임스페이스가 비어 있을 경우, 초기 명세와 변수를 적용할 수도 있습니다.

```sh
./dist/uniflow test --namespace default --from-specs examples/specs.yaml --from-values examples/values.yaml
```

환경 변수는 `--env` 플래그로 지정할 수 있습니다.

```sh
./dist/uniflow test --namespace default --env DATABASE_URL=mongodb://localhost:27017 --env DATABASE_NAME=mydb
```

### Apply 명령어

`apply` 명령어는 지정된 파일 내용을 네임스페이스에 적용합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow apply nodes --namespace default --filename examples/specs.yaml
```

변수을 적용하려면:

```sh
./dist/uniflow apply values --namespace default --filename examples/values.yaml
```

### Delete 명령어

`delete` 명령어는 지정된 파일에 정의된 모든 리소스를 삭제합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow delete nodes --namespace default --filename examples/specs.yaml
```

변수을 삭제하려면:

```sh
./dist/uniflow delete values --namespace default --filename examples/values.yaml
```

### Get 명령어

`get` 명령어는 지정된 네임스페이스 내 모든 리소스를 조회합니다. 네임스페이스가 지정되지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow get nodes --namespace default
```

변수을 조회하려면:

```sh
./dist/uniflow get values --namespace default
```
