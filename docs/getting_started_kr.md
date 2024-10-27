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

설정은 `.uniflow.toml` 파일이나 시스템 환경 변수를 사용해 유연하게 변경할 수 있습니다. 주요 설정 항목은 다음과 같습니다:

| TOML 키              | 환경 변수 키            | 예시                       |
|----------------------|-------------------------|----------------------------|
| `database.url`       | `DATABASE.URL`          | `mem://` 또는 `mongodb://` |
| `database.name`      | `DATABASE.NAME`         | -                          |
| `collection.charts`  | `COLLECTION.CHARTS`     | `charts`                   |
| `collection.nodes`   | `COLLECTION.NODES`      | `nodes`                    |
| `collection.secrets` | `COLLECTION.SECRETS`    | `secrets`                  |

만약 [MongoDB](https://www.mongodb.com/)를 사용한다면, 리소스의 변경 사항을 실시간으로 추적하기 위해 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 이를 위해서는 [복제 세트](https://www.mongodb.com/docs/manual/replication/) 설정이 필요합니다.

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


## Uniflow 사용하기

`uniflow`는 주로 런타임 환경을 시작하고 관리하는 명령어입니다.

### Start 명령어

`start` 명령어는 지정된 네임스페이스 내의 모든 노드 명세를 실행합니다. 네임스페이스가 지정되지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflow start --namespace default
```

네임스페이스가 비어 있을 경우, 초기 노드 명세를 `--from-specs` 플래그로 제공할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-specs examples/nodes.yaml
```

초기 시크릿 파일은 `--from-secrets` 플래그로 설정할 수 있습니다:
```sh
./dist/uniflow start --namespace default --from-secrets examples/secrets.yaml
```

초기 차트 파일은 `--from-charts` 플래그로 제공할 수 있습니다:

```sh
./dist/uniflow start --namespace default --from-charts examples/charts.yaml
```

## Uniflowctl 사용하기

`uniflowctl`는 네임스페이스 내에서 리소스를 관리하는 명령어입니다.

### Apply 명령어

`apply` 명령어는 지정된 파일 내용을 네임스페이스에 적용합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl apply nodes --namespace default --filename examples/nodes.yaml
```

시크릿을 적용하려면:

```sh
./dist/uniflowctl apply secrets --namespace default --filename examples/secrets.yaml
```

차트를 적용하려면:

```sh
./dist/uniflowctl apply charts --namespace default --filename examples/charts.yaml
```

### Delete 명령어

`delete` 명령어는 지정된 파일에 정의된 모든 리소스를 삭제합니다. 네임스페이스를 지정하지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl delete nodes --namespace default --filename examples/nodes.yaml
```

시크릿을 삭제하려면:

```sh
./dist/uniflowctl delete secrets --namespace default --filename examples/secrets.yaml
```

차트를 삭제하려면:

```sh
./dist/uniflowctl delete charts --namespace default --filename examples/charts.yaml
```

### Get 명령어

`get` 명령어는 지정된 네임스페이스 내 모든 리소스를 조회합니다. 네임스페이스가 지정되지 않으면 기본적으로 `default` 네임스페이스가 사용됩니다.

```sh
./dist/uniflowctl get nodes --namespace default
```

시크릿을 조회하려면:

```sh
./dist/uniflowctl get secrets --namespace default
```

차트를 조회하려면:

```sh
./dist/uniflowctl get charts --namespace default
```

## HTTP API 통합

HTTP API를 통해 노드 명세를 수정하려면, 관련 워크플로우를 설정해야 합니다. 이를 위해 [기본 확장](../ext/README_kr.md)에 포함된 `native` 노드를 사용할 수 있습니다:

```yaml
kind: native
opcode: nodes.create # 또는 nodes.read, nodes.update, nodes.delete
```

시작하려면 [워크플로우 예제](../examples/system.yaml)를 참고하세요. 필요한 경우, 인증 및 권한 관리 프로세스를 추가할 수 있습니다. 이러한 런타임 제어 워크플로우는 보통 `system` 네임스페이스에 정의됩니다.
