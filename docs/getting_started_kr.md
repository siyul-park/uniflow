# 🚀 시작하기

이 안내서는 [명령줄 인터페이스(CLI)](../cmd/README_kr.md)를 설치하고 구성하며 워크플로우를 관리하는 방법을 자세히 설명합니다. 설치에서부터 워크플로우 제어 및 설정까지의 전반적인 과정을 다룹니다.

## 소스에서 설치하기

먼저 [내장된 확장 기능](../ext/README_kr.md)과 함께 제공되는 [CLI](../cmd/README_kr.md)를 설정해야 합니다. 설치를 시작하기 전에 시스템에 [Go 1.22](https://go.dev/doc/install) 이상이 설치되어 있는지 확인하세요.

### 리포지토리 클론

소스 코드를 클론하려면 다음 명령어를 실행합니다:

```sh
git clone https://github.com/siyul-park/uniflow
```

클론한 디렉토리로 이동합니다:

```sh
cd uniflow
```

### 의존성 설치 및 빌드

의존성을 설치하고 프로젝트를 빌드하려면 다음 명령어를 실행합니다:

```sh
make init
make build
```

빌드가 완료되면 `dist` 폴더에 실행 파일이 생성됩니다.

### 설정

`.uniflow.toml` 파일이나 시스템 환경 변수를 통해 설정을 유연하게 변경할 수 있습니다. 주요 구성 옵션은 다음과 같습니다:

| TOML 키              | 환경 변수 키            | 예시                       |
|----------------------|-------------------------|----------------------------|
| `database.url`       | `DATABASE.URL`          | `mem://` 또는 `mongodb://` |
| `database.name`      | `DATABASE.NAME`         | -                          |
| `collection.nodes`   | `COLLECTION.NODES`      | `nodes`                    |
| `collection.secrets` | `COLLECTION.SECRETS`    | `secrets`                  |

[MongoDB](https://www.mongodb.com/)를 사용하는 경우, 엔진이 노드 명세 및 시크릿 변경을 추적할 수 있도록 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 이를 위해 [복제 세트](https://www.mongodb.com/docs/manual/replication/) 설정이 필요합니다.

## Uniflow

`uniflow`는 주로 런타임 환경을 시작하고 관리하는 데 사용됩니다.

### Start

`start` 명령어는 특정 네임스페이스의 노드 명세로 런타임을 시작합니다. 기본 사용법은 다음과 같습니다:

```sh
./dist/uniflow start --namespace default
```

네임스페이스가 비어 있는 경우, `--from-nodes` 플래그를 사용하여 초기 노드 명세를 제공할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-nodes examples/nodes.yaml
```

초기 시크릿을 제공하려면 `--from-secrets` 플래그를 사용할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

이 명령어는 지정된 네임스페이스의 모든 노드 명세를 실행합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스가 사용됩니다.

## Uniflowctl

`uniflowctl`는 네임스페이스 내에서 노드 명세와 시크릿을 관리하는 데 사용됩니다.

### Apply

`apply` 명령어는 네임스페이스에 노드 명세 또는 시크릿을 추가하거나 업데이트합니다. 사용 예시는 다음과 같습니다:

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/nodes.yaml
```

또는

```sh
./dist/uniflowctl apply secrets --namespace default --filename examples/secrets.yaml
```

이 명령어는 지정된 파일의 내용을 네임스페이스에 적용합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

### Delete

`delete` 명령어는 네임스페이스에서 노드 명세 또는 시크릿을 제거합니다. 사용 예시는 다음과 같습니다:

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/nodes.yaml
```

또는

```sh
./dist/uniflowctl delete secrets --namespace default --filename examples/secrets.yaml
```

이 명령어는 지정된 파일에 정의된 모든 노드 명세 또는 시크릿을 제거합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스가 사용됩니다.

### Get

`get` 명령어는 네임스페이스에서 노드 명세 또는 시크릿을 조회합니다. 사용 예시는 다음과 같습니다:

```sh
./dist/uniflowctl get nodes --namespace default
```

또는

```sh
./dist/uniflowctl get secrets --namespace default
```

이 명령어는 지정된 네임스페이스의 모든 노드 명세 또는 시크릿을 표시합니다. 네임스페이스를 지정하지 않으면 `default` 네임스페이스가 사용됩니다.

## HTTP API 통합

HTTP API를 통해 노드 명세를 수정하려면 해당 기능을 노출하는 워크플로우를 설정해야 합니다. 이를 위해 [기본 확장](../ext/README_kr.md)에 포함된 `native` 노드를 활용할 수 있습니다:

```yaml
kind: native
opcode: nodes.create # 또는 nodes.read, nodes.update, nodes.delete
```

시작하려면 [워크플로우 예제](../examples/system.yaml)를 참고하세요. 필요에 따라 이 워크플로우에 인증 및 권한 부여 프로세스를 추가할 수 있습니다. 일반적으로 이러한 런타임 제어 워크플로우는 `system` 네임스페이스에 정의됩니다.
