# 🚀 시작하기

이 종합 안내서는 [명령줄 인터페이스(CLI)](../cmd/README_kr.md)를 사용하여 워크플로우를 관리하고 엔진을 실행하는 방법을 다룹니다. 설치부터 워크플로우 제어 및 구성 설정까지의 모든 과정을 포함하고 있습니다.

## 소스에서 설치하기

먼저 [내장된 확장 기능](../ext/README_kr.md)과 함께 워크플로우를 제어할 수 있는 [CLI](../cmd/README_kr.md)를 설정합니다. 시작하기 전에 시스템에 [Go 1.22](https://go.dev/doc/install) 이상이 설치되어 있는지 확인하세요.

리포지토리를 클론하는 것으로 시작합니다:

```sh
git clone https://github.com/siyul-park/uniflow
```

클론한 디렉토리로 이동합니다:

```sh
cd uniflow
```

의존성을 설치하고 프로젝트를 빌드합니다:

```sh
make init
make build
```

빌드 과정이 완료되면 `dist` 폴더에 실행 파일이 생성되어 사용할 준비가 됩니다.

## 구성

Uniflow는 `.uniflow.toml` 파일이나 시스템 환경 변수를 통해 유연한 구성 옵션을 제공합니다:

| TOML 키              | 환경 변수 키            | 예시                       |
|----------------------|-------------------------|----------------------------|
| `database.url`       | `DATABASE.URL`          | `mem://` 또는 `mongodb://` |
| `database.name`      | `DATABASE.NAME`         | -                          |
| `collection.nodes`   | `COLLECTION.NODES`      | `nodes`                    |
| `collection.secrets` | `COLLECTION.SECRETS`    | `secrets`                  |

[MongoDB](https://www.mongodb.com/)를 사용할 경우, 엔진이 노드 명세 및 시크릿 변경을 추적할 수 있도록 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 이를 위해서는 [복제 세트](https://www.mongodb.com/docs/manual/replication/) 설정이 필요합니다.

## CLI 명령어

CLI는 워크플로우 제어를 위한 다양한 명령어를 제공합니다. 사용 가능한 모든 명령어를 보려면 다음을 실행하세요:

```sh
./dist/uniflowctl --help
```
```sh
./dist/uniflow --help
```

### Apply

`apply` 명령어는 네임스페이스에 노드 명세 또는 시크릿을 추가하거나 업데이트합니다. 명령어는 다음과 같이 사용할 수 있습니다:

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/nodes.yaml
```

또는

```sh
./dist/uniflowctl apply secrets --namespace default --filename examples/secrets.yaml
```

이 명령어는 결과를 출력하며, 네임스페이스를 지정하지 않으면 `default` 네임스페이스를 사용합니다.

### Delete

`delete` 명령어는 네임스페이스에서 노드 명세 또는 시크릿을 제거합니다:

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/nodes.yaml
```

또는

```sh
./dist/uniflowctl delete secrets --namespace default --filename examples/secrets.yaml
```

이 명령어는 `examples/nodes.yaml` 또는 `examples/secrets.yaml`에 정의된 모든 노드 명세 또는 시크릿을 지정된 네임스페이스에서 제거합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스를 사용합니다.

### Get

`get` 명령어는 네임스페이스에서 노드 명세 또는 시크릿을 조회합니다:

```sh
./dist/uniflowctl get nodes --namespace default
```

또는

```sh
./dist/uniflowctl get secrets --namespace default
```

이 명령어는 지정된 네임스페이스의 모든 노드 명세 또는 시크릿을 표시합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스를 사용합니다.

### Start

`start` 명령어는 특정 네임스페이스의 노드 명세로 런타임을 시작합니다:

```sh
./dist/uniflow start --namespace default
```

네임스페이스가 비어 있을 경우, `--from-nodes` 플래그를 사용하여 초기 노드 명세를 제공할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.yaml
```

또한, 초기 시크릿을 제공하려면 `--from-secrets` 플래그를 사용할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

이 명령어는 지정된 네임스페이스의 모든 노드 명세를 실행합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스를 사용합니다.

## HTTP API 통합

HTTP API를 통해 노드 명세를 수정하려면 해당 기능을 노출하는 워크플로우를 설정해야 합니다. [기본 확장](../ext/README_kr.md)에 포함된 `syscall` 노드를 활용할 수 있습니다:

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

시작하려면 [워크플로우 예제](../examples/system.yaml)를 참고하세요. 필요에 따라 이 워크플로우에 인증 및 권한 부여 프로세스를 추가할 수 있습니다. 일반적으로 이러한 런타임 제어 워크플로우는 `system` 네임스페이스에 정의됩니다.
