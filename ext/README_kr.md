# 🔧 기본 확장 기능

기본 확장 기능을 통해 다양한 작업을 효율적으로 처리하고 시스템 성능을 극대화할 수 있습니다.

## 사용 가능한 확장 기능

### **Control**

데이터 흐름을 정밀하게 제어합니다.

- **[Block 노드](./docs/block_node_kr.md)**: 여러 하위 노드를 묶어 복잡한 데이터 처리 흐름을 관리하며, 각 하위 노드는 특정 작업을 수행합니다.
- **[Cache 노드](./docs/cache_node_kr.md)**: LRU(Least Recently Used) 캐시를 사용하여 데이터를 저장하고 조회합니다.
- **[For 노드](./docs/for_node_kr.md)**: 입력 패킷을 여러 하위 패킷으로 나누어 반복 처리합니다.
- **[Fork 노드](./docs/fork_node_kr.md)**: 데이터 흐름을 비동기적으로 분기하여 독립적인 작업을 병렬로 수행합니다.
- **[If 노드](./docs/if_node_kr.md)**: 조건을 평가하여 패킷을 두 경로로 분기합니다.
- **[Merge 노드](./docs/merge_node_kr.md)**: 여러 입력 패킷을 하나로 통합합니다.
- **[NOP 노드](./docs/nop_node_kr.md)**: 입력 패킷을 처리하지 않고 빈 패킷으로 응답합니다.
- **[Pipe 노드](./docs/pipe_node_kr.md)**: 입력 패킷을 처리하고 결과를 여러 출력 포트로 전달하여 데이터 흐름을 재사용합니다.
- **[Retry 노드](./docs/retry_node_kr.md)**: 오류가 발생하면 지정된 횟수만큼 패킷 처리를 재시도합니다.
- **[Session 노드](./docs/session_node_kr.md)**: 프로세스 정보를 저장하고 관리하여 세션을 유지합니다.
- **[Sleep 노드](./docs/sleep_node_kr.md)**: 지정된 지연 시간을 추가하여 워크플로우를 조정하거나 외부 조건을 기다립니다.
- **[Snippet 노드](./docs/snippet_node_kr.md)**: 다양한 프로그래밍 언어로 작성된 코드 스니펫을 실행하여 입력 패킷을 처리합니다.
- **[Split 노드](./docs/split_node_kr.md)**: 입력 패킷을 여러 개로 나누어 처리합니다.
- **[Step 노드](./docs/step_node_kr.md)**: 복잡한 데이터 처리 흐름을 체계적으로 관리하며, 여러 하위 노드를 순차적으로 실행합니다.
- **[Switch 노드](./docs/switch_node_kr.md)**: 입력 패킷을 조건에 따라 여러 포트 중 하나로 분기합니다.
- **[Throw 노드](./docs/throw_node_kr.md)**: 입력 패킷을 기반으로 에러를 발생시키고, 이를 응답으로 반환합니다.
- **[Try 노드](./docs/try_node_kr.md)**: 패킷 처리 중 발생할 수 있는 오류를 오류 포트를 통해 적절히 처리합니다.

### **IO**

외부 데이터 소스와의 상호작용을 지원합니다.

- **[Print 노드](./docs/print_node_kr.md)**: 입력된 데이터를 파일에 출력하여 디버깅이나 데이터 흐름 모니터링에 사용됩니다.
- **[Scan 노드](./docs/scan_node_kr.md)**: 다양한 형식의 입력 데이터를 스캔하고 필요한 데이터를 추출하여 처리합니다.
- **[SQL 노드](./docs/sql_node_kr.md)**: 관계형 데이터베이스와 상호작용하여 SQL 쿼리를 실행하고 결과를 패킷으로 반환합니다.

### **Network**

다양한 네트워크 프로토콜을 지원하여 네트워크 작업을 원활하게 수행합니다.

- **[HTTP 노드](./docs/http_node_kr.md)**: HTTP 요청을 처리하고 응답을 반환하여 웹 서비스와 통신합니다.
- **[WebSocket 노드](./docs/websocket_node_kr.md)**: WebSocket 연결을 설정하고 메시지를 송수신합니다.
- **[Upgrade 노드](./docs/upgrade_node_kr.md)**: HTTP 연결을 WebSocket으로 업그레이드하여 실시간 데이터 통신을 지원합니다.
- **[Listener 노드](./docs/listener_node_kr.md)**: 지정된 프로토콜과 포트에서 네트워크 요청을 수신합니다.
- **[Router 노드](./docs/router_node_kr.md)**: 입력 패킷을 조건에 따라 여러 출력 포트로 라우팅합니다.
