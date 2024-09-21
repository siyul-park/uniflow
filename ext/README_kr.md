# 🔧 기본 확장 기능

기본 확장 기능을 통해 다양한 작업을 효율적으로 처리하고 시스템 성능을 극대화할 수 있습니다.

## 사용 가능한 확장 기능

### **Control**

데이터 흐름을 정밀하게 제어합니다.

- **[Call 노드](./docs/call_node_kr.md)**: 입력 패킷을 처리하고 결과를 여러 출력 포트로 전달하여 데이터 흐름을 재사용합니다.
- **[Block 노드](./docs/block_node_kr.md)**: 복잡한 데이터 처리 흐름을 체계적으로 관리하며, 여러 하위 노드를 순차적으로 실행합니다.
- **[Fork 노드](./docs/fork_node_kr.md)**: 데이터 흐름을 비동기적으로 분기하여 독립적인 작업을 병렬로 수행합니다.
- **[If 노드](./docs/if_node_kr.md)**: 조건을 평가하여 패킷을 두 경로로 분기합니다.
- **[Loop 노드](./docs/loop_node_kr.md)**: 입력 패킷을 여러 하위 패킷으로 나누어 반복 처리합니다.
- **[Merge 노드](./docs/merge_node_kr.md)**: 여러 입력 패킷을 하나로 통합합니다.
- **[NOP 노드](./docs/nop_node_kr.md)**: 입력 패킷을 처리하지 않고 빈 패킷으로 응답합니다.
- **[Reduce 노드](./docs/reduce_node_kr.md)**: 입력 데이터를 반복적으로 연산하여 하나의 출력 값을 생성합니다. 데이터 집계에 유용합니다.
- **[Session 노드](./docs/session_node_kr.md)**: 프로세스 정보를 저장하고 관리하여 세션을 유지합니다.
- **[Snippet 노드](./docs/snippet_node_kr.md)**: 다양한 프로그래밍 언어로 작성된 코드 스니펫을 실행하여 입력 패킷을 처리합니다.
- **[Split 노드](./docs/split_node_kr.md)**: 입력 패킷을 여러 개로 나누어 처리합니다.
- **[Switch 노드](./docs/switch_node_kr.md)**: 입력 패킷을 조건에 따라 여러 포트 중 하나로 분기합니다.

### **IO**

외부 데이터 소스와의 상호작용을 지원합니다.

- **[Print 노드](./docs/print_node_kr.md)**: 입력된 데이터를 파일에 출력하여 디버깅이나 데이터 흐름 모니터링에 사용됩니다.
- **[Scan 노드](./docs/scan_node_kr.md)**: 다양한 형식의 입력 데이터를 스캔하고 필요한 데이터를 추출하여 처리합니다.
- **[SQL 노드](./docs/sql_node_kr.md)**: 관계형 데이터베이스와 상호작용하여 SQL 쿼리를 실행하고 결과를 패킷으로 반환합니다.

### **Network**

다양한 네트워크 프로토콜을 지원하여 네트워크 작업을 원활하게 수행합니다.

- **[HTTP 노드](./docs/http_node_kr.md)**: HTTP 요청을 처리하고 응답을 반환하여 웹 서비스와 통신합니다.
- **[WebSocket 노드](./docs/websocket_node_kr.md)**: WebSocket 연결을 설정하고 메시지를 송수신합니다.
- **[Gateway 노드](./docs/gateway_node_kr.md)**: HTTP 연결을 WebSocket으로 업그레이드하여 실시간 데이터 통신을 지원합니다.
- **[Listener 노드](./docs/listener_node_kr.md)**: 지정된 프로토콜과 포트에서 네트워크 요청을 수신합니다.
- **[Proxy 노드](./docs/proxy_node_kr.md)**: HTTP 요청을 다른 서버로 프록시하여 응답을 반환합니다.
- **[Router 노드](./docs/router_node_kr.md)**: 입력 패킷을 조건에 따라 여러 출력 포트로 라우팅합니다.

### **System**

시스템 구성 요소를 관리하고 최적화합니다.

- **[Native 노드](./docs/native_node_kr.md)**: 시스템 내부에서 함수 호출을 수행하고 결과를 패킷으로 반환합니다.
