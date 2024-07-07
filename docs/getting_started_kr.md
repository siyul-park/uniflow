# 시작하기

이 가이드는 Command-Line Interface (CLI)를 사용하여 워크플로우를 관리하고 엔진을 실행하는 방법을 설명합니다. CLI 설치부터 워크플로우 제어 및 설정 구성까지의 과정을 살펴보세요.

## 코드에서 설치

[내장된 확장 기능](../ext/README_kr.md)과 함께 워크플로우를 제어할 수 있는 CLI를 설치합니다. 코드를 빌드하려면 [Go 1.22](https://go.dev/doc/install) 이상이 필요합니다.

먼저, 레포지토리에서 코드를 클론하세요.

```sh
git clone https://github.com/siyul-park/uniflow
```

클론한 디렉토리로 이동합니다.

```sh
cd uniflow
```

의존성을 설치하고 프로젝트를 빌드합니다.

```sh
make init
make build
```

빌드가 완료되면 `dist` 폴더에 실행 파일이 생성됩니다. 이를 통해 CLI를 사용할 수 있습니다.

## 구성

환경 설정은 `.uniflow.toml` 파일이나 시스템 환경 변수를 사용하여 구성할 수 있습니다.

| TOML 키         | 환경 변수 키      | 예시                     |
|------------------|------------------|-------------------------|
| `database.url`   | `DATABASE.URL`   | `mem://` 또는 `mongodb://` |
| `database.name`  | `DATABASE.NAME`  | -                       |

[MongoDB](https://www.mongodb.com/)를 사용할 경우 런타임 엔진이 노드 명세의 변경을 추적할 수 있도록 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)이 활성화되어 있어야 합니다. 변경 스트림을 이용하기 위해 [복제본 세트](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication)를 활용하세요.

## 명령어

CLI는 워크플로우를 제어하기 위해 여러 가지 명령어를 제공합니다. 전체 명령어 목록을 확인하려면 다음 도움말 명령어를 참조하세요:

```sh
./dist/uniflow --help
```

### Apply

`apply` 명령어는 노드 명세들을 특정 네임스페이스에 적용합니다. 노드 명세가 이미 네임스페이스에 존재한다면 기존 명세를 업데이트하고, 존재하지 않으면 새로 생성합니다. 명세가 적용된 결과를 출력합니다. 네임스페이스를 명시하지 않으면 기본적으로 `default` 네임스페이스에 적용됩니다.

```sh
./dist/uniflow apply --filename examples/ping.yaml
 ID                                    KIND      NAMESPACE  NAME      LINKS                                
 01908c74-8b22-7cbf-a475-6b6bc871b01a  listener  <nil>      listener  map[out:[map[name:router port:in]]]  
 01908c74-8b22-7cc0-ae2b-40504e7c9ff0  router    <nil>      router    map[out[0]:[map[name:pong port:in]]] 
 01908c74-8b22-7cc1-ac48-83b5084a0061  snippet   <nil>      pong      <nil>                                
```

### Delete

`delete` 명령어는 네임스페이스에 존재하는 노드 명세들을 제거합니다. 이는 특정 워크플로우의 노드를 삭제할 때 유용합니다. 네임스페이스를 명시하지 않으면 `default` 네임스페이스에서 삭제됩니다.

```sh
./dist/uniflow delete --filename examples/ping.yaml
```

위 명령어를 사용하면 `examples/ping.yaml`에 정의된 모든 노드 명세가 해당 네임스페이스에서 삭제됩니다.

### Get

`get` 명령어는 네임스페이스에 존재하는 노드 명세들을 조회합니다. 네임스페이스를 명시하지 않으면 `default` 네임스페이스를 조회합니다.

```sh
./dist/uniflow get
 ID                                    KIND      NAMESPACE  NAME      LINKS                                
 01908c74-8b22-7cbf-a475-6b6bc871b01a  listener  <nil>      listener  map[out:[map[name:router port:in]]]  
 01908c74-8b22-7cc0-ae2b-40504e7c9ff0  router    <nil>      router    map[out[0]:[map[name:pong port:in]]] 
 01908c74-8b22-7cc1-ac48-83b5084a0061  snippet   <nil>      pong      <nil>                                
```

### Start

`start` 명령어는 특정 네임스페이스에 있는 노드 명세를 로드하고 런타임을 실행합니다. 네임스페이스를 명시하지 않으면 `default` 네임스페이스를 실행합니다.

```sh
./dist/uniflow start                  
```

만약 네임스페이스가 비어 있어 실행할 노드가 없을 경우 `--filename` 플래그를 통해 네임스페이스에 기본적으로 설치할 노드 명세를 제공할 수 있습니다.

```sh
./dist/uniflow start --filename examples/ping.yaml
```

## HTTP API

HTTP API를 통해 노드 명세를 수정하려면 CLI를 활용하여 HTTP API를 노출하는 워크플로우를 네임스페이스에 설치해야 합니다. 이를 위해 기본 확장에 포함된 `syscall` 노드를 활용할 수 있습니다.

```yaml
kind: syscall
opcode: nodes.create # nodes.read, nodes.update, nodes.delete
```

[워크플로우 예시](../examples/crud.yaml)를 확인하여 시작해보세요. 필요에 따라 인증과 인가 과정을 이 워크플로우에 포함할 수 있습니다. 일반적으로 이러한 런타임 제어와 관련된 워크플로우는 `system` 네임스페이스에 정의됩니다.