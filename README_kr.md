# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**높은 성능과 극도의 유연성을 갖춘 쉽게 확장할 수 있는 다목적 워크플로우 엔진.**

## 📝 개요

**Uniflow**는 다양한 작업 시간을 효율적으로 처리하며 데이터 처리 흐름을 선언적으로 정의하고 동적으로 수정할 수 있는 환경을 제공합니다. [내장된 확장 기능](./ext/README_kr.md)을 통해 작업을 효율적으로 실행하며, 필요에 따라 노드를 자유롭게 추가하거나 제거할 수 있습니다.

당신의 서비스에 맞춤형 경험을 제공하고, 기능을 무한히 확장하세요.

## 🎯 핵심 가치

- **성능:** 다양한 환경에서 최대 처리량과 최소 대기 시간을 달성합니다.
- **유연성:** 명세를 동적으로 수정하고 실시간으로 조정할 수 있습니다.
- **확장성:** 새로운 노드를 자유롭게 추가하여 기능을 확장할 수 있습니다.

## 🚀 빠른 시작

### 🛠️ 빌드 및 설치

**[Go 1.22](https://go.dev/doc/install)** 이상이 필요합니다. 코드를 빌드하려면 다음 단계를 따르세요:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build
```

빌드가 완료되면 `dist` 폴더에 실행 파일이 생성됩니다.

### ⚡ 예제 실행

간단한 HTTP 요청 처리 예제인 [ping.yaml](./examples/ping.yaml)를 실행해 보겠습니다:

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  links:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: GET
      path: /ping
      port: out[0]
  links:
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
uniflow start --filename example/ping.yaml
```

예상대로 작동하는지 확인하기 위해 HTTP 엔드포인트를 호출하세요:

```sh
curl localhost:8000/ping
pong#
```

## ⚙️ 구성

환경 설정은 `.uniflow.toml` 파일이나 시스템 환경 변수를 통해 설정할 수 있습니다.

| TOML 키            | 환경 변수 키          | 예시                       |
|--------------------|--------------------|---------------------------|
| `database.url`     | `DATABASE.URL`     | `mem://` 또는 `mongodb://` |
| `database.name`    | `DATABASE.NAME`    | -                         |
| `collection.nodes` | `COLLECTION.NODES` | `nodes`                   |

## 📊 벤치마크

**[Contabo](https://contabo.com/)**의 VPS S SSD (4코어, 8GB)에서 수행된 벤치마크 결과입니다. [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html)를 사용하여 `listener`, `router`, `snippet` 노드로 구성된 [ping.yaml](./examples/ping.yaml) 워크플로우를 측정했습니다.

```sh
ab -n 102400 -c 1024 http://127.0.0.1:8000/ping
```

```
This is ApacheBench, Version 2.3 <$Revision: 1879490 $>
Benchmarking 127.0.0.1 (be patient)
Server Hostname:        127.0.0.1
Server Port:            8000
Document Path:          /ping
Document Length:        0 bytes
Concurrency Level:      1024
Time taken for tests:   14.951 seconds
Complete requests:      102400
Failed requests:        0
Total transferred:      11878400 bytes
Requests per second:    6849.06 [#/sec] (mean)
Time per request:       149.510 [ms] (mean)
Time per request:       0.146 [ms] (mean, across all concurrent requests)
Transfer rate:          775.87 [Kbytes/sec] received
Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   4.7      0      55
Processing:     0  146  62.2    148     528
Waiting:        0  145  62.2    147     528
Total:          0  148  62.3    150     559
Percentage of the requests served within a certain time (ms)
  50%    150
  66%    174
  75%    188
  80%    196
  90%    223
  95%    248
  98%    282
  99%    307
 100%    559 (longest request)
```

## 📚 자세히 알아보기

- [시작하기](./docs/getting_started_kr.md): CLI 설치 및 워크플로우 관리 방법을 살펴보세요.
- [핵심 개념](./docs/key_concepts_kr.md): 노드, 연결, 포트, 패킷 등의 핵심 개념을 이해하세요.
- [아키텍처](./docs/architecture_kr.md): 노드 명세 로딩 및 워크플로우 실행 과정을 자세히 알아보세요.
- [사용자 기능 확장](./docs/user_extensions_kr.md): 새로운 기능 추가 및 기존 서비스 통합 방법을 익히세요.

## 🌐 커뮤니티 & 지원

프로젝트에 대한 질문이나 지원이 필요하신 경우, 다음 채널을 통해 참여해보세요:

- [토론](https://github.com/siyul-park/uniflow/discussions): 질문 및 피드백을 공유하세요.
- [이슈 트래커](https://github.com/siyul-park/uniflow/issues): 버그 보고 및 기능 요청을 남겨주세요.

## 📜 라이센스

이 프로젝트는 [MIT 라이센스](./LICENSE) 하에 배포됩니다. 자유롭게 수정 및 재배포가 가능합니다.

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
