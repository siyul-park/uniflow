# 🏗️ 아키텍처

작업을 처리하는 최소 단위인 노드를 기반으로 하여, 노드 명세는 각 노드가 동작할 역할을 정의하고, 이 노드들이 서로 연결되어 워크플로우를 형성합니다. 각 워크플로우는 하나의 런타임 안에서 사전에 정의된 네임스페이스에 연결되어 동작하며, 각 런타임 환경은 하나의 네임스페이스를 실행합니다.

네임스페이스는 다른 네임스페이스에 정의된 노드를 임의적으로 참조할 수 없으며, 각각 격리되어 관리됩니다.

```text
   +-------------------------------------------------+
   |                   Workflow A                    |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   |                       \|/                       |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   +-------------------------------------------------+
   |                   Workflow B                    |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   |                       \|/                       |
   |  +--------------------+ +--------------------+  |
   |  | Node Specification | | Node Specification |  |
   |  +--------------------+ +--------------------+  |
   +-------------------------------------------------+
```

엔진은 특정 노드를 강제하지 않으며, 모든 노드는 서비스에 맞게 자유롭게 추가되고 제거될 수 있습니다.

### 워크플로우 수정

엔진은 사용자에게 노드 명세 변경을 위한 API를 제공하지 않으며, 노드를 로드하고 컴파일하여 실행하는 데 집중합니다. 노드 명세를 수정할 필요가 있을 때는 Command-Line Interface(CLI)나 직접 정의한 워크플로우를 통해 HTTP API로 명세를 업데이트할 수 있습니다. 이러한 워크플로우는 보통 `system` 네임스페이스에 정의됩니다.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: '{{ .PORT }}'
  ports:
    out:
      - name: router
        port: in
    error:
      - name: catch
        port: in
  env:
    PORT:
      data: '{{ .PORT }}'

- kind: router
  name: router
  routes:
    - method: POST
      path: /v1/nodes
      port: out[0]
    - method: GET
      path: /v1/nodes
      port: out[1]
    - method: PATCH
      path: /v1/nodes
      port: out[2]
    - method: DELETE
      path: /v1/nodes
      port: out[3]
  ports:
    out[0]:
      - name: nodes_create
        port: in
    out[1]:
      - name: nodes_read
        port: in
    out[2]:
      - name: nodes_update
        port: in
    out[3]:
      - name: nodes_delete
        port: in

- kind: block
  name: nodes_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.create
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 201
          };
        }

- kind: block
  name: nodes_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.read
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: nodes_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.update
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            body: args,
            status: 200
          };
        }

- kind: block
  name: nodes_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.delete
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            status: 204
          };
        }

- kind: switch
  name: catch
  matches:
    - when: self == "unsupported type" || self == "unsupported value"
      port: out[0]
    - when: 'true'
      port: out[1]
  ports:
    out[0]:
      - name: status_400
        port: in
    out[1]:
      - name: status_500
        port: in

- kind: snippet
  name: status_400
  language: javascript
  code: |
    export default function (args) {
      return {
        body: {
          error: args.error()
        },
        status: 400
      };
    }

- kind: snippet
  name: status_500
  language: json
  code: |
    {
      "body": {
        "error": "Internal Server Error"
      },
      "status": 500
    }
```

이 접근 방식은 안정적인 런타임 환경을 유지하면서 시스템을 유연하게 확장할 수 있도록 합니다.

### 컴파일 과정

로더는 데이터베이스의 변경 스트림을 통해 실시간으로 노드 명세와 시크릿의 변경 사항을 추적합니다. 추가, 수정, 삭제가 발생하면 로더는 해당 명세를 다시 로드하고, 스키마에 정의된 코덱을 활용해 실행 가능한 형태로 컴파일합니다. 캐싱과 최적화 과정도 함께 수행되어 성능을 개선합니다.

컴파일된 노드는 명세와 결합하여 심볼로 변환되며, 심볼 테이블에 저장됩니다. 심볼 테이블은 각 심볼의 포트를 노드 명세에 정의된 포트 연결 정보를 기반으로 연결합니다.
로더는 데이터베이스의 변경 스트림을 통해 실시간으로 노드 명세와 시크릿의 변경 사항을 추적합니다. 추가, 수정, 삭제가 발생하면 로더는 해당 명세를 다시 로드하고, 스키마에 정의된 코덱을 활용해 실행 가능한 형태로 컴파일합니다. 캐싱과 최적화 과정도 함께 수행되어 성능을 개선합니다.

```text
```text
   +--------------------------+
   |         Database         |
   |  +--------------------+  |
   |  | Node Specification |  |
   |  +--------------------+  |
   |  | Node Specification |  |
   |  +--------------------+  |   +-------------------+
   |  | Node Specification |  |-->|       Loader      |
   |  +--------------------+  |   |  +-------------+  |
   +--------------------------+   |  |    Scheme   |  |
   +--------------------------+   |  |  +-------+  |  |
   |         Database         |   |  |  | Codec |  |  |--+
   |  +--------+  +--------+  |   |  |  +-------+  |  |  |
   |  | Secret |  | Secret |  |-->|  +-------------+  |  |
   |  +--------+  +--------+  |   +-------------------+  |
   |  +--------+  +--------+  |                          |
   |  | Secret |  | Secret |  |                          |
   |  +--------+  +--------+  |                          |
   +--------------------------+                          |
   +-------------------------+                           |
   |      Symbol Table       |                           |
   |  +--------+ +--------+  |                           |
   |  | Symbol | | Symbol |<-----------------------------+
   |  +--------+ +--------+  |
   |           \|/           |
   |  +--------+ +--------+  |
   |  | Symbol | | Symbol |  |
   |  +--------+ +--------+  |
   +-------------------------+
```

컴파일된 노드는 심볼 테이블에 저장되어, 각 심볼이 정의된 포트에 따라 연결됩니다. 워크플로우의 모든 노드가 심볼 테이블에 로드되면, 노드를 활성화하기 위한 순차적 작업이 실행됩니다. 노드가 제거되면 비활성화 작업도 순차적으로 실행됩니다.

### 런타임 과정

활성화된 노드는 워크플로우를 실행하며, 독립적인 프로세스를 통해 리소스를 관리하고 다른 작업에 영향을 주지 않도록 합니다. 각 노드는 프로세스 간 통신을 통해 패킷을 주고받으며, 페이로드는 공용 타입으로 변환되어 전송됩니다.

```text
```text
   +-----------------------+          +-----------------------+
   |        Node A         |          |        Node B         |
   |  +-----------------+  |          |  +-----------------+  |
   |  |    Port Out     |--------------->|    Port In      |  |
   |  |                 |  |          |  |                 |  |
   |  |  +-----------+  |  |  packet  |  |  +-----------+  |  |
   |  |  | Process 1 |======================| Process 1 |  |  |
   |  |  +-----------+  |  |          |  |  +-----------+  |  |
   |  |  +-----------+  |  |  packet  |  |  +-----------+  |  |
   |  |  | Process 2 |======================| Process 2 |  |  |
   |  |  +-----------+  |  |          |  |  +-----------+  |  |
   |  +-----------------+  |          |  +-----------------+  |
   +-----------------------+          +-----------------------+
```

하나의 리더는 모든 패킷을 순차적으로 처리하고, 반드시 라이터로 전송된 패킷에 대한 응답 패킷을 반환해야 합니다. 이를 통해 노드 간 통신이 원활히 이루어지며, 데이터의 일관성과 무결성이 보장됩니다.

워크플로우를 실행한 노드는 전송한 모든 패킷의 응답을 받을 때까지 기다렸다가, 프로세스를 종료하고 할당된 리소스를 해제합니다. 패킷 처리 중 오류가 발생해 오류 응답이 반환될 경우, 노드는 해당 오류를 기록한 뒤 프로세스를 종료합니다.

프로세스가 종료되면, 정상 종료 여부를 확인하고 열린 파일 디스크립터, 할당된 메모리, 데이터베이스 트랜잭션 등을 해제합니다.

부모 프로세스가 종료되면, 그에 의해 파생된 모든 자식 프로세스도 함께 종료됩니다. 이때 부모 프로세스는 자식 프로세스가 모두 종료될 때까지 대기합니다.
