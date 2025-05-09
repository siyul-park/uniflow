# 📚 핵심 개념

이 가이드는 시스템에서 사용되는 주요 용어와 개념을 자세히 설명합니다.

## 핵심 용어 요약

| 용어 | 설명 |
|------|------|
| **명세(Spec)** | 노드의 정의를 담고 있는 정보입니다. JSON/YAML 형태로 표현되며 ID, 네임스페이스, 이름, 종류, 포트 정보 등을 포함합니다. |
| **노드(Node)** | 워크플로우의 기본 실행 단위입니다. 입력을 받아 처리하고 출력을 생성합니다. |
| **심볼(Symbol)** | 노드를 런타임에서 관리할 수 있도록 래핑한 객체입니다. 노드 인스턴스와 명세 정보를 함께 보유합니다. |
| **포트(Port)** | 노드 간 연결 지점입니다. 입력 포트(InPort)와 출력 포트(OutPort)로 구분됩니다. |
| **패킷(Packet)** | 노드 간 전달되는 데이터의 기본 단위입니다. 요청 패킷과 응답 패킷이 있습니다. |
| **프로세스(Process)** | 워크플로우 실행의 독립적인 단위입니다. 고유 ID와 상태를 가지며 패킷 처리의 컨텍스트를 제공합니다. |
| **런타임(Runtime)** | 워크플로우를 관리하고 실행하는 환경입니다. |
| **로더(Loader)** | 명세를 로드하고 노드로 변환하여 심볼 테이블에 등록합니다. |
| **심볼 테이블(Symbol Table)** | 심볼을 ID와 이름으로 관리하고 심볼 간 연결을 설정합니다. |
| **스키마(Scheme)** | 노드 유형별 코덱을 등록하고 명세를 노드로 변환하는 규칙을 제공합니다. |
| **훅(Hook)** | 심볼의 생명주기 및 포트 활성화 이벤트를 처리합니다. LoadHook, UnloadHook, OpenHook 등이 있습니다. |
| **네임스페이스(Namespace)** | 워크플로우를 격리하여 관리하며, 독립된 실행 환경을 제공합니다. |
| **변수(Variable)** | 비밀번호, API 키 등 노드에서 필요로 하는 민감한 정보를 안전하게 저장합니다. |

## 노드 명세

노드 명세는 각 노드의 동작 방식과 연결을 선언적으로 정의합니다. 엔진은 이 명세를 실행 가능한 노드로 컴파일합니다.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01a
kind: listener
namespace: default
name: example-listener
annotations:
  description: "HTTP 프로토콜을 사용하는 예제 리스너 노드"
  version: "1.0"
protocol: http
port: "{{ .PORT }}"
env:
  PORT:
    name: network
    data: "{{ .PORT }}"
ports:
  out:
    - name: proxy
      port: in
```

- `id`: UUID 형식의 고유 식별자입니다. UUID V7을 권장합니다.
- `kind`: 노드의 종류를 지정합니다. 노드 종류에 따라 추가 필드가 달라질 수 있습니다.
- `namespace`: 노드가 속한 네임스페이스를 지정하며, 기본값은 `default`입니다.
- `name`: 노드의 이름을 지정하며, 동일한 네임스페이스 내에서 고유해야 합니다.
- `annotations`: 노드에 대한 추가 메타데이터입니다. 설명, 버전 등 사용자 정의 키-값 쌍을 포함할 수 있습니다.
- `protocol`: 리스너에서 사용하는 프로토콜을 지정합니다. `listener` 종류의 노드에서 추가로 요구하는 필드입니다.
- `port`: 리스너에서 사용하는 포트를 지정합니다. `listener` 종류의 노드에서 추가로 요구하는 필드입니다.
- `ports`: 포트의 연결 방식을 정의합니다. `out`은 `proxy`라는 이름의 출력 포트를 정의하며, 다른 노드의 `in` 포트에 연결됩니다.
- `env`: 노드에 필요한 환경 변수를 지정합니다. 여기서는 `PORT`가 변수으로부터 동적으로 설정됩니다.

## 변수

변수은 비밀번호, API 키 등 노드에서 필요로 하는 민감한 정보를 안전하게 저장합니다.

```yaml
id: 01908c74-8b22-7cbf-a475-6b6bc871b01b
namespace: default
name: database
annotations:
  description: "데이터베이스 정보"
data:
  password: "super-value-password"
```

- `id`: UUID 형식의 고유 식별자입니다. UUID V7을 권장합니다.
- `namespace`: 변수이 속한 네임스페이스를 지정하며, 기본값은 `default`입니다.
- `name`: 변수의 이름을 지정하며, 동일한 네임스페이스 내에서 고유해야 합니다.
- `annotations`: 변수에 대한 추가 메타데이터입니다. 설명, 버전 등 사용자 정의 키-값 쌍을 포함할 수 있습니다.
- `data`: 키-값 쌍으로 구성된 변수 데이터를 포함합니다.

## 노드

노드는 데이터를 처리하는 객체로, 서로 연결된 포트를 통해 패킷을 주고받으며 워크플로우를 실행합니다. 각 노드는 독립적인 처리 루프를 가지며, 비동기적으로 다른 노드와 통신합니다.

노드는 패킷 처리 방식에 따라 다음과 같이 분류됩니다:
- `ZeroToOne`: 초기 패킷을 생성하여 워크플로우를 시작하는 노드입니다.
- `OneToOne`: 입력 포트에서 패킷을 받아 처리하고 출력 포트로 전송하는 노드입니다.
- `OneToMany`: 입력 포트에서 패킷을 받아 여러 출력 포트로 전송하는 노드입니다.
- `ManyToOne`: 여러 입력 포트에서 패킷을 받아 하나의 출력 포트로 전송하는 노드입니다.
- `Other`: 단순 패킷 전달 이외의 상태 관리 및 상호작용을 포함하는 노드입니다.

## 포트

포트는 노드 간에 패킷을 주고받는 연결 지점입니다. 포트에는 `InPort`와 `OutPort` 두 가지 종류가 있으며, 이들을 연결하여 패킷을 전송합니다. 한 포트로 전송된 패킷은 모든 연결된 포트로 전달됩니다.

일반적으로 사용되는 포트 이름은 다음과 같습니다:
- **`init`**: 노드를 초기화하는 데 사용되는 특수 포트입니다. 노드가 활성화될 때 `init` 포트에 연결된 워크플로우가 실행됩니다.
- **`term`**: 노드를 종료하는 데 사용되는 특수 포트입니다. 노드가 비활성화될 때 `term` 포트에 연결된 워크플로우가 실행됩니다.
- **`io`**: 패킷을 처리하고 즉시 반환합니다.
- **`in`**: 패킷을 입력받아 처리하며, 처리 결과를 `out` 또는 `error`로 전송합니다. 연결된 `out`이나 `error` 포트가 없을 경우, 결과를 즉시 반환합니다.
- **`out`**: 처리된 패킷을 전송합니다. 전송된 결과는 다른 `in` 포트로 전달될 수 있습니다.
- **`error`**: 패킷 처리 중 발생한 오류를 전송합니다. 오류 처리 결과는 `in` 포트로 다시 전달될 수 있습니다.

동일한 역할을 하는 여러 포트가 필요할 경우, `in[0]`, `in[1]`, `out[0]`, `out[1]`과 같이 표기합니다.

## 패킷

패킷은 포트 간에 교환되는 데이터 단위입니다. 각 패킷은 페이로드를 포함하며, 노드는 이를 처리하여 전송합니다.

노드는 요청 패킷의 순서에 따라 응답 패킷을 반환해야 합니다. 여러 포트에 연결된 경우, 모든 응답 패킷을 모아 하나의 새로운 응답 패킷으로 반환합니다.

특수한 `None` 패킷은 응답이 없음을 나타내며, 단순히 패킷이 수락되었음을 표시합니다.

## 프로세스

프로세스는 실행의 기본 단위로, 독립적으로 관리됩니다. 프로세스는 부모 프로세스를 가질 수 있으며, 부모 프로세스가 종료되면 자식 프로세스도 함께 종료됩니다.

프로세스는 패킷으로 전송하기 어려운 값을 저장하기 위해 각자의 저장소를 가지고 있습니다. 이 저장소는 Copy-On-Write (COW) 방식으로 동작하여 부모 프로세스의 데이터를 효율적으로 공유합니다.

새로운 워크플로우는 프로세스를 생성하여 시작됩니다. 프로세스가 종료되면 사용된 모든 자원이 해제됩니다.

프로세스는 두 개 이상의 루트 패킷을 가질 수 있지만, 루트 패킷은 같은 노드에서 생성되어야 하며, 다른 노드에서 생성된 경우 새로운 자식 프로세스를 생성하여 처리해야 합니다.

## 워크플로우

워크플로우는 방향성을 가진 그래프로 정의되며, 여러 노드가 연결된 구조입니다. 이 그래프에서 각 노드는 데이터 처리를 담당하며, 패킷은 노드들 사이를 통해 전송됩니다.

워크플로우는 여러 단계로 구성되며, 각 단계에서 데이터는 정의된 규칙에 따라 처리되고 전달됩니다. 이 과정에서 데이터는 순차적으로 또는 병렬적으로 처리될 수 있습니다.

예를 들어, 초기 데이터가 주어지면 첫 번째 노드에서 처리된 후 다음 노드로 전달됩니다. 각 노드는 입력을 받아 처리하고, 처리된 결과를 다음 단계로 전송합니다.

## 네임스페이스

네임스페이스는 워크플로우를 격리하여 관리하며, 독립된 실행 환경을 제공합니다. 각 네임스페이스는 여러 워크플로우를 포함할 수 있으며, 네임스페이스 내의 노드는 다른 네임스페이스에 속한 노드를 참조할 수 없습니다. 각 네임스페이스는 자체 데이터 및 리소스를 독립적으로 관리합니다.

## 런타임 환경

런타임 환경은 각 네임스페이스가 실행되는 독립적인 공간입니다. 엔진은 네임스페이스에 속한 모든 노드를 로드하여 환경을 구축하고 워크플로우를 실행합니다. 이를 통해 워크플로우 실행 중 발생할 수 있는 충돌을 방지하고 안정적인 실행 환경을 제공합니다.
