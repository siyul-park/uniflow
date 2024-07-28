# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**고성능과 유연성을 겸비한 확장 가능한 범용 워크플로우 엔진.**

## 📝 개요

**Uniflow**는 단기 작업부터 장기 프로세스까지 다양한 작업을 효율적으로 처리합니다. 데이터 흐름을 선언적으로 정의하고 동적으로 수정할 수 있으며, [내장 확장 기능](./ext/README_kr.md)을 활용하여 복잡한 워크플로우도 쉽게 구현할 수 있습니다. 게다가 필요에 따라 새로운 노드를 추가하거나 기존 노드를 제거하여 기능을 유연하게 확장할 수 있습니다.

여러분의 서비스에 개인화된 경험을 제공하고, 지속적으로 기능을 확장해 나가세요.

## 🎯 핵심 가치

- **성능:** 다양한 환경에서 최적의 처리량과 최소 지연 시간을 실현합니다.
- **유연성:** 워크플로우를 동적으로 수정하고 실시간으로 조정할 수 있습니다.
- **확장성:** 새로운 컴포넌트를 통해 시스템 기능을 확장할 수 있습니다.

## 🚀 빠른 시작

### 🛠️ 빌드 및 설치

**[Go 1.22](https://go.dev/doc/install)** 이상이 필요합니다. 소스 코드를 빌드하려면 다음 단계를 따르세요:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build
```

빌드가 완료되면 `dist` 디렉토리에 실행 파일이 생성됩니다.

### ⚡ 예제 실행

기본적인 HTTP 요청 처리 예제인 [ping.yaml](./examples/ping.yaml)을 실행해 보겠습니다:

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: GET
      path: /ping
      port: out[0]
  ports:
    out[0]:
      - name: pong
        port: in

- kind: snippet
  name: pong
  language: text
  code: pong
```

워크플로우를 실행하려면 다음 명령어를 사용하세요:

```sh
uniflow start --from-nodes example/ping.yaml
```

정상 작동 여부를 확인하기 위해 HTTP 엔드포인트를 호출해 보세요:

```sh
curl localhost:8000/ping
pong#
```

## ⚙️ 구성

환경 설정은 `.uniflow.toml` 파일 또는 시스템 환경 변수를 통해 구성할 수 있습니다.

| TOML 키              | 환경 변수 키            | 예시                       |
|----------------------|--------------------------|---------------------------|
| `database.url`       | `DATABASE.URL`           | `mem://` 또는 `mongodb://` |
| `database.name`      | `DATABASE.NAME`          | -                         |
| `collection.nodes`   | `COLLECTION.NODES`       | `nodes`                   |
| `collection.secrets` | `COLLECTION.SECRETS`     | `secrets`                 |

## 📊 벤치마크

**[Contabo](https://contabo.com/)** VPS S SSD (4코어, 8GB) 환경에서 수행된 벤치마크 결과입니다. [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html)을 사용하여 `listener`, `router`, `snippet` 노드로 구성된 [ping.yaml](./examples/ping.yaml) 워크플로우를 측정했습니다.

```sh
ab -n 102400 -c 1024 http://127.0.0.1:8000/ping
```

```
This is ApacheBench, Version 2.3 <$Revision: 1879490 $>
Benchmarking 127.0.0.1 (be patient)
Server Hostname:        127.0.0.1
Server Port:            8000
Document Path:          /ping
Document Length:        4 bytes
Concurrency Level:      1024
Time taken for tests:   13.760 seconds
Complete requests:      102400
Failed requests:        0
Total transferred:      12288000 bytes
HTML transferred:       409600 bytes
Requests per second:    7441.92 [#/sec] (mean)
Time per request:       137.599 [ms] (mean)
Time per request:       0.134 [ms] (mean, across all concurrent requests)
Transfer rate:          872.10 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   3.1      0      34
Processing:     0  136  58.2    137     550
Waiting:        0  135  58.2    137     550
Total:          0  137  58.0    139     553

Percentage of the requests served within a certain time (ms)
  50%    139
  66%    162
  75%    174
  80%    181
  90%    202
  95%    223
  98%    264
  99%    295
 100%    553 (longest request)
```

## 📚 자세히 알아보기

- [시작하기](./docs/getting_started_kr.md): CLI 설치 및 워크플로우 관리 방법을 소개합니다.
- [핵심 개념](./docs/key_concepts_kr.md): 노드, 연결, 포트, 패킷 등의 기본 개념을 설명합니다.
- [아키텍처](./docs/architecture_kr.md): 노드 명세 로딩 및 워크플로우 실행 과정을 상세히 설명합니다.
- [사용자 확장](./docs/user_extensions_kr.md): 새로운 기능 추가 및 기존 서비스 통합 방법을 안내합니다.

## 🌐 커뮤니티 및 지원

프로젝트에 관한 질문이나 지원이 필요한 경우, 다음 채널을 이용해 주세요:

- [토론 포럼](https://github.com/siyul-park/uniflow/discussions): 질문 및 피드백을 공유할 수 있습니다.
- [이슈 트래커](https://github.com/siyul-park/uniflow/issues): 버그 보고 및 기능 요청을 제출할 수 있습니다.

## 📜 라이센스

이 프로젝트는 [MIT 라이센스](./LICENSE)에 따라 배포됩니다. 라이센스 조건에 따라 자유롭게 사용, 수정 및 배포가 가능합니다.

<!-- Go -->

[go_download_url]: https://golang.org/dl/
[go_version_img]: https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go
[go_code_coverage_img]: https://codecov.io/gh/siyul-park/uniflow/graph/badge.svg?token=quEl9AbBcW
[go_code_coverage_url]: https://codecov.io/gh/siyul-park/uniflow
[go_report_img]: https://goreportcard.com/badge/github.com/siyul-park/uniflow
[go_report_url]: https://goreportcard.com/report/github.com/siyul-park/uniflow
[go_doc_img]: https://godoc.org/github.com/siyul-park/uniflow?status.svg
[go_doc_url]: https://godoc.org/github.com/siyul-park/uniflow

<!-- Repository -->

[repo_url]: https://github.com/siyul-park/uniflow
[repo_issues_url]: https://github.com/siyul-park/uniflow/issues
[repo_pull_request_url]: https://github.com/siyul-park/uniflow/pulls
[repo_discussions_url]: https://github.com/siyul-park/uniflow/discussions
[repo_releases_img]: https://img.shields.io/github/release/siyul-park/uniflow.svg
[repo_releases_url]: https://github.com/siyul-park/uniflow/releases
[repo_wiki_url]: https://github.com/siyul-park/uniflow/wiki
[repo_wiki_img]: https://img.shields.io/badge/docs-wiki_page-blue?style=for-the-badge&logo=none
[repo_wiki_faq_url]: https://github.com/siyul-park/uniflow/wiki/FAQ
[repo_check_img]: https://github.com/siyul-park/uniflow/actions/workflows/check.yml/badge.svg
[repo_check_url]: https://github.com/siyul-park/uniflow/actions/workflows/check.yml
```
