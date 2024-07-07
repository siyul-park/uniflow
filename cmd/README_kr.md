# 명령줄 인터페이스 (CLI)

다양한 작업 흐름 관리를 위해 설계된 다목적 명령줄 인터페이스 (CLI)를 효과적으로 관리하세요. 이 CLI는 [내장 확장 기능](../ext/README.md)을 포함한 기본 실행 파일로 제공됩니다.

## 구성

명령을 실행하기 전에 환경 변수를 사용하여 시스템을 구성하세요. `.uniflow.toml` 파일이나 시스템 환경 변수를 활용할 수 있습니다.

| 키             | 예시                  |
|-----------------|-----------------------|
| `database.url`  | `mem://` 또는 `mongodb://` |
| `database.name` | -                     |

[MongoDB](https://www.mongodb.com/)를 사용할 경우 런타임 엔진이 노드 명세의 변경을 추적할 수 있도록 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)이 활성화되어 있어야 합니다. 변경 스트림을 이용하기 위해 [복제본 세트](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication)를 활용하세요.
