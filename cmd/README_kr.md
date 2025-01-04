# 💻 명령줄 인터페이스 (CLI)

다양한 작업 흐름을 관리하기 위해 설계된 다목적 명령줄 인터페이스 (CLI)를 효과적으로 사용하세요. 이 CLI는 [내장 확장 기능](../ext/README.md)을 포함한 기본 실행 파일로 제공됩니다.

### 설정

설정은 `.uniflow.toml` 파일이나 시스템 환경 변수를 사용해 유연하게 변경할 수 있습니다. 주요 설정 항목은 다음과 같습니다:

| TOML 키              | 환경 변수 키             | 예시                       |
|---------------------|---------------------|--------------------------|
| `database.url`      | `DATABASE_URL`      | `mem://` 또는 `mongodb://` |
| `database.name`     | `DATABASE_NAME`     | -                        |
| `collection.specs`  | `COLLECTION_SPECS`  | `specs`                  |
| `collection.values` | `COLLECTION_VALUES` | `values`                 |

만약 [MongoDB](https://www.mongodb.com/)를 사용한다면, 리소스의 변경 사항을 실시간으로 추적하기 위해 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 이를 위해서는 [복제 세트](https://www.mongodb.com/docs/manual/replication/) 설정이 필요합니다.

## Uniflow 사용하기

`uniflow`는 주로 런타임 환경을 시작하고 관리하는 명령어입니다.

### Start 명령어

`start` 명령어는 지정된 네임스페이스 내의 모든 노드 명세를 실행합니다. 네임스페이스가 지정되지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow start --namespace default
```

네임스페이스가 비어 있을 경우, 초기 노드 명세를 `--from-specs` 플래그로 제공할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-specs examples/specs.yaml
```

초기 변수 파일은 `--from-values` 플래그로 설정할 수 있습니다:
```sh
./dist/uniflow start --namespace default --from-values examples/values.yaml
```

## Uniflowctl 사용하기

`uniflowctl`는 네임스페이스 내에서 리소스를 관리하는 명령어입니다.

### Apply 명령어

`apply` 명령어는 지정된 파일 내용을 네임스페이스에 적용합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/specs.yaml
```

변수을 적용하려면:

```sh
./dist/uniflowctl apply values --namespace default --filename examples/values.yaml
```

### Delete 명령어

`delete` 명령어는 지정된 파일에 정의된 모든 리소스를 삭제합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/specs.yaml
```

변수을 삭제하려면:

```sh
./dist/uniflowctl delete values --namespace default --filename examples/values.yaml
```

### Get 명령어

`get` 명령어는 지정된 네임스페이스 내 모든 리소스를 조회합니다. 네임스페이스가 지정되지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl get nodes --namespace default
```

변수을 조회하려면:

```sh
./dist/uniflowctl get values --namespace default
```
