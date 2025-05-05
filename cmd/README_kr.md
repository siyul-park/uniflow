# 💻 명령줄 인터페이스 (CLI)

다양한 작업 흐름을 관리하기 위해 설계된 다목적 명령줄 인터페이스 (CLI)를 효과적으로 사용하세요.

## 설정

환경 변수 또는 설정 파일(`.toml`, `.yaml`, `.json`, `.hjson`, `.env`)을 사용하여 설정을 구성할 수 있습니다. 설정 파일의 경로는 `UNIFLOW_CONFIG` 환경 변수로 지정하며, 지정하지 않으면 기본값인 `.uniflow.toml` 파일이 사용됩니다.

```bash
export UNIFLOW_CONFIG=./config/uniflow.toml
```

설정 파일에서 정의할 수 있는 주요 항목은 다음과 같습니다:

```toml
[runtime]
namespace = "default"
language = "cel"

[database]
url = "memory://"

[collection]
specs = "specs"
values = "values"

[[plugins]]
path = "./dist/cel.so"
config.extensions = ["encoders", "math", "lists", "sets", "strings"]

[[plugins]]
path = "./dist/ecmascript.so"

[[plugins]]
path = "./dist/mongodb.so"

[[plugins]]
path = "./dist/reflect.so"

[[plugins]]
path = "./dist/ctrl.so"

[[plugins]]
path = "./dist/net.so"

[[plugins]]
path = "./dist/sql.so"

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

환경 변수는 `--environment` 플래그로 지정할 수 있습니다.

```sh
./dist/uniflow start --namespace default --environment DATABASE_URL=mongodb://localhost:27017 --environment DATABASE_NAME=mydb
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

환경 변수는 `--environment` 플래그로 지정할 수 있습니다.

```sh
./dist/uniflow test --namespace default --environment DATABASE_URL=mongodb://localhost:27017 --environment DATABASE_NAME=mydb
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