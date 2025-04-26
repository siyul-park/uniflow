# **블록 노드**

**블록 노드**는 복잡한 데이터 처리 흐름을 관리하는 노드로, 여러 하위 노드를 묶어 일련의 데이터 처리 작업을 조직화하고 제어합니다. 하위 노드들은 순차적 혹은 병렬로 실행될 수 있으며, 각 하위 노드는 데이터를
변환하거나 외부 시스템과 상호작용하는 등의 작업을 수행합니다.

## **명세**

- **specs**: 실행할 하위 노드들의 목록을 정의합니다. 각 하위 노드는 데이터를 처리하며, 이들이 상호작용하여 전체 데이터 흐름을 구성합니다. 하위 노드들은 지정된 순서대로 또는 병렬로 실행될 수 있습니다.
- **inbound**: 외부에서 들어오는 데이터를 처리하는 입력 포트를 정의합니다. 이 포트는 데이터 처리 흐름의 시작 지점 역할을 합니다.
- **outbound**: 처리된 결과를 외부로 전달하는 출력 포트를 정의합니다. 결과 데이터는 이 포트를 통해 외부 시스템으로 전달됩니다.

## **포트**

모든 포트는 런타임 중에 동적으로 결정되며, 각 포트는 하위 노드와의 연결에 따라 자동으로 할당됩니다.

## **예시**

```yaml
- kind: block
  specs:
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return {
            payload: {
              method: 'GET',
              body: args
            }
          };
        }
      ports:
        out:
          - name: $1
            port: in
    - kind: http
      url: https://api.example.com/data
      ports:
        out:
          - name: $2
            port: in
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return args.body;
        }
  inbounds:
    in:
      - name: $0
        port: in
  outbounds:
    out:
      - name: $2
        port: out
```
