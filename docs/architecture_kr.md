# 🏗️ 아키텍처

각 노드 명세는 각 노드의 역할을 선언적으로 정의하며, 이러한 명세들이 서로 연결되어 워크플로우를 형성합니다. 각 워크플로우는 특정 네임스페이스에 정의되며, 각 런타임 환경은 하나의 네임스페이스를 실행합니다. 네임스페이스는 다른 네임스페이스에 정의된 노드를 참조할 수 없으며, 각각 격리되어 관리됩니다.

```plantext
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

엔진은 특정 노드의 사용을 강요하지 않습니다. 모든 노드는 확장을 통해 엔진에 연결되며, 서비스에 맞게 자유롭게 추가되고 제거될 수 있습니다.

노드 명세를 효과적으로 실행하기 위해 컴파일과 런타임이라는 두 가지 주요 과정을 거칩니다. 이를 통해 복잡성을 줄이고 성능을 최적화할 수 있습니다.

## 워크플로우 수정

엔진은 노드 명세를 변경할 수 있는 API를 사용자에게 노출하지 않습니다. 대신, 엔진은 오직 노드를 로드하고 컴파일하여 실행 가능하게 활성화하는 데 집중합니다.

노드 명세를 수정해야 할 경우, Command-Line Interface(CLI)을 사용하여 데이터베이스에 명세를 업데이트하거나, 노드 명세를 수정할 수 있는 워크플로우를 직접 정의하여 HTTP API를 제공할 수 있습니다. 일반적으로 이러한 워크플로우는 `system` 네임스페이스에 정의됩니다.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in
    error:
      - name: catch
        port: in

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
    - method: POST
      path: /v1/secrets
      port: out[4]
    - method: GET
      path: /v1/secrets
      port: out[5]
    - method: PATCH
      path: /v1/secrets
      port: out[6]
    - method: DELETE
      path: /v1/secrets
      port: out[7]
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
    out[4]:
      - name: secrets_create
        port: in
    out[5]:
      - name: secrets_read
        port: in
    out[6]:
      - name: secrets_update
        port: in
    out[7]:
      - name: secrets_delete
        port: in

- kind: block
  name: nodes_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.create

- kind: block
  name: nodes_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.read

- kind: block
  name: nodes_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: nodes.update

- kind: block
  name: nodes_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: nodes.delete

- kind: block
  name: secrets_create
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: secrets.create

- kind: block
  name: secrets_read
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: secrets.read

- kind: block
  name: secrets_update
  specs:
    - kind: snippet
      language: cel
      code: 'has(self.body) ? self.body : null'
    - kind: native
      opcode: secrets.update

- kind: block
  name: secrets_delete
  specs:
    - kind: snippet
      language: json
      code: 'null'
    - kind: native
      opcode: secrets.delete

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

이러한 접근 방식은 런타임 환경을 안정적으로 유지하면서도 필요에 따라 유연하게 시스템을 확장할 수 있도록 합니다.

## 컴파일 과정

로더는 데이터베이스의 변경 스트림을 통해 실시간으로 노드 명세와 시크릿의 변경 사항을 추적합니다. 추가, 수정, 또는 삭제가 이루어지면 로더에 의해 감지되어 데이터베이스에서 동적으로 다시 로드됩니다. 그런 다음, 스키마에 정의된 코덱을 활용하여 실행 가능한 형태로 컴파일됩니다. 이 과정에서 캐싱과 최적화가 수행되어 성능이 개선됩니다.

컴파일된 노드는 명세와 결합하여 심볼로 변환되며, 심볼 테이블에 저장됩니다. 심볼 테이블은 각 심볼의 포트를 노드 명세에 정의된 포트 연결 정보를 기반으로 연결합니다.

```plantext
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

워크플로우에 속한 모든 노드가 심볼 테이블에 로드되고 노드의 모든 포트가 연결되면, 워크플로우의 내부 노드부터 외부 노드까지 순차적으로 로드 훅을 실행하여 노드를 활성화합니다. 그러나 심볼 테이블에서 특정 노드가 제거되어 실행할 수 없는 상태가 되면, 해당 노드를 참조하는 모든 노드를 활성화의 역순으로 언로드 훅을 실행하여 비활성화합니다.

데이터베이스에서 노드 명세가 변경되면, 이 과정을 통해 변경 사항이 모든 런타임 환경에 반영됩니다.

## 런타임 과정

활성화된 노드는 소켓이나 파일을 감시하며, 워크플로우를 실행합니다. 노드는 독립적인 프로세스를 생성하여 실행을 시작하고, 실행 흐름을 다른 프로세스로부터 격리합니다. 이를 통해 필요한 리소스를 효율적으로 관리하여 다른 작업에 영향을 미치지 않도록 합니다.

각 노드는 프로세스를 통해 포트를 열고, 라이터를 생성하여 연결된 다른 노드에게 패킷을 전송합니다. 패킷의 페이로드는 런타임에서 사용되는 공용 타입으로 변환되어 전송됩니다.

연결된 노드는 새로운 프로세스가 해당 포트를 열었는지를 감시하며, 리더를 생성합니다. 생성된 리더는 대기 중인 패킷을 지속적으로 처리하고, 처리된 결과를 다음 노드에 전달하거나 반환합니다.

```plantext
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

하나의 리더는 모든 패킷을 순차적으로 처리하고 응답해야 하며, 라이터로 전송된 패킷은 반드시 응답 패킷으로 반환되어야 합니다. 이를 통해 노드 간 통신이 원활하게 이루어지고, 데이터의 일관성과 무결성이 유지됩니다.

워크플로우를 실행한 노드는 전송한 모든 패킷의 응답이 반환될 때까지 기다린 후, 프로세스를 종료하고 할당된 리소스를 해제합니다. 패킷 처리 중 오류가 발생하여 오류 응답이 반환된 경우, 노드는 해당 오류를 기록하고 프로세스를 종료합니다.

프로세스가 종료되면, 정상 종료 여부를 확인한 후 열린 파일 디스크립터, 할당된 메모리, 데이터베이스 트랜잭션 등을 모두 해제합니다.

부모 프로세스가 종료되면, 파생된 모든 자식 프로세스도 종료됩니다. 이때 부모 프로세스는 일반적으로 자식 프로세스가 모두 종료될 때까지 대기합니다.
