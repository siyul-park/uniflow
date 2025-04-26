# Step 노드

**Step 노드**는 복잡한 데이터 처리 흐름을 체계적으로 관리하며, 여러 하위 노드를 순차적으로 실행하는 기능을 제공합니다. 이를 통해 데이터 처리 작업을 명확하고 효율적으로 구성할 수 있습니다.

## 명세

- **specs**: 실행할 하위 노드의 목록을 정의합니다. 각 하위 노드는 데이터 처리 흐름에서 특정 단계를 담당하며, 순차적으로 실행됩니다.

## 포트

- **in**: 입력된 패킷을 첫 번째 하위 노드로 전달합니다.
- **out**: 마지막 하위 노드에서 처리된 결과를 외부로 출력합니다.
- **error**: 하위 노드에서 발생한 오류를 외부로 전달합니다.

## 예시

```yaml
- kind: step
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
    - kind: http
      url: https://api.example.com/data
    - kind: snippet
      language: javascript
      code: |
        export default function (args) {
          return args.body;
        }
```
