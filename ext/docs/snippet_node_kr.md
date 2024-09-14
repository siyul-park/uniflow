# Snippet 노드

**Snippet 노드**는 다양한 프로그래밍 언어로 작성된 코드 스니펫을 실행하여 입력된 패킷을 처리하고 그 결과를 출력합니다. 이 노드는 복잡한 데이터 처리 로직을 유연하게 적용할 수 있으며, 동적으로 코드 실행을 통해 데이터 처리의 다양성을 제공합니다.

## 명세

- **language**: 코드 스니펫이 작성된 프로그래밍 언어를 지정합니다. (예: `text`, `json`, `yaml`, `cel`, `javascript`, `typescript`)
- **code**: 실행할 코드 스니펫을 입력합니다.

## 포트

- **in**: 입력된 패킷을 코드로 전달하여 실행합니다.
- **out**: 코드 실행 결과를 출력합니다.
- **error**: 코드 실행 중 발생한 오류를 전달합니다.

## 예시

```yaml
- kind: snippet
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
```