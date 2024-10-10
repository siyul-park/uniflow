# 💻 명령줄 인터페이스 (CLI)

다양한 작업 흐름을 관리하기 위해 설계된 다목적 명령줄 인터페이스 (CLI)를 효과적으로 사용하세요. 이 CLI는 [내장 확장 기능](../ext/README.md)을 포함한 기본 실행 파일로 제공됩니다.

## 구성

명령을 실행하기 전에 환경 변수를 사용하여 시스템을 구성해야 합니다. `.uniflow.toml` 파일이나 시스템 환경 변수를 활용할 수 있습니다.

| TOML 키              | 환경 변수 키            | 예시                       |
|----------------------|--------------------------|----------------------------|
| `database.url`       | `DATABASE.URL`           | `mem://` 또는 `mongodb://` |
| `database.name`      | `DATABASE.NAME`          | -                          |
| `collection.charts`  | `COLLECTION.CHARTS`      | `charts`                   |
| `collection.nodes`   | `COLLECTION.NODES`       | `nodes`                    |
| `collection.secrets` | `COLLECTION.SECRETS`     | `secrets`                  |

[MongoDB](https://www.mongodb.com/)를 사용할 경우, 엔진이 리소스의 변경을 추적할 수 있도록 [변경 스트림](https://www.mongodb.com/docs/manual/changeStreams/)을 활성화해야 합니다. 변경 스트림을 이용하려면 [복제본 세트](https://www.mongodb.com/ko-kr/docs/manual/replication/#std-label-replication)를 설정하세요.
