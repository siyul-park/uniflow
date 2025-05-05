# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![ci][repo_ci_img]][repo_ci_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**높은 성능과 뛰어난 유연성을 갖춘 확장 가능한 범용 워크플로우 엔진**

## 📝 개요

복잡한 데이터 흐름을 효율적으로 관리하고 실시간으로 조정하세요. [플러그인](plugins/README_kr.md)을 추가하여 워크플로우를 쉽게 확장할 수 있으며, 비즈니스 요구에 맞게 프로세스를 최적화할 수 있습니다. 이를 통해 개인화된 서비스를 제공하고, 시스템의 지속적인 발전을 위한 강력한 기반을 구축해 보세요.

## 🎯 핵심 가치

- **성능:** 다양한 환경에서 최대 처리량과 최소 지연 시간을 제공합니다.
- **유연성:** 워크플로우를 실시간으로 수정하고 조정할 수 있습니다.
- **확장성:** 새로운 컴포넌트를 쉽게 추가하여 기능을 확장할 수 있습니다.

## 🚀 빠른 시작

### 🛠️ 빌드 및 설치

**[Go 1.24 이상](https://go.dev/doc/install)**이 설치된 환경에서 다음 명령어로 소스 코드를 빌드할 수 있습니다:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build-all
```

빌드가 완료되면 `dist` 디렉터리에 실행 파일이 생성됩니다.

### ⚙️ 환경 설정

설정은 환경 변수 또는 `.uniflow.toml` 파일을 통해 관리할 수 있습니다.  
기본 제공 플러그인을 등록하고 설정하는 예시는 다음과 같습니다:

```toml
[runtime]
namespace = "default"
language = "cel"

[database]
url = "memory://"

[collection]
specs = "specs"
values = "values"

[[plugins]]
path = "./dist/cel.so"
config.extensions = ["encoders", "math", "lists", "sets", "strings"]

[[plugins]]
path = "./dist/ecmascript.so"

[[plugins]]
path = "./dist/mongodb.so"

[[plugins]]
path = "./dist/reflect.so"

[[plugins]]
path = "./dist/ctrl.so"

[[plugins]]
path = "./dist/net.so"

[[plugins]]
path = "./dist/sql.so"

[[plugins]]
path = "./dist/testing.so"
```

### ⚡ 예제 실행

HTTP 요청을 처리하는 [ping.yaml](examples/ping.yaml) 예제를 실행하는 방법은 다음과 같습니다:

```yaml
- kind: listener
  name: listener
  protocol: http
  port: '{{ .PORT }}'
  env:
    PORT:
      data: '{{ .PORT }}'
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

다음 명령어를 사용해 워크플로우를 실행할 수 있습니다:

```sh
./dist/uniflow start --from-specs ./examples/ping.yaml --environment PORT=8000
```

정상 작동 여부를 확인하려면 아래 명령어로 HTTP 엔드포인트를 호출하세요:

```sh
curl localhost:8000/ping
pong#
```

## 📊 벤치마크

다음 벤치마크는 **[Contabo](https://contabo.com/)** VPS S SSD (4코어, 8GB) 환경에서 수행되었습니다.  
`listener`, `router`, `snippet` 노드로 구성된 [ping.yaml](examples/ping.yaml) 워크플로우를
대상으로, [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html)을 사용해 테스트했습니다:

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
Time taken for tests:   122.866 seconds
Complete requests:      1024000
Failed requests:        0
Total transferred:      122880000 bytes
HTML transferred:       4096000 bytes
Requests per second:    8334.29 [#/sec] (mean)
Time per request:       122.866 [ms] (mean)
Time per request:       0.120 [ms] (mean, across all concurrent requests)
Transfer rate:          976.67 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   3.8      0      56
Processing:     0  121  53.4    121     593
Waiting:        0  120  53.4    121     592
Total:          0  123  53.3    123     594

Percentage of the requests served within a certain time (ms)
  50%    123
  66%    143
  75%    155
  80%    163
  90%    185
  95%    207
  98%    240
  99%    266
 100%    594 (longest request)
```

## 📚 더 알아보기

- [시작하기](./docs/getting_started_kr.md): CLI 설치 및 워크플로우 관리 방법을 알아보세요.
- [핵심 개념](./docs/key_concepts_kr.md): 노드, 연결, 포트, 패킷 등 주요 개념을 이해하세요.
- [아키텍처](./docs/architecture_kr.md): 명세 로딩과 워크플로우 실행 과정을 살펴보세요.
- [플로우차트](./docs/flowchart_kr.md): 컴파일 및 런타임 프로세스를 단계별로 알아보세요.
- [디버깅](./docs/debugging_kr.md): 문제 해결과 디버깅 방법을 확인하세요.
- [사용자 확장](./docs/user_extensions_kr.md): 새로운 노드 추가 및 서비스 통합 방법을 안내합니다.

## 🌐 커뮤니티 및 지원

- [토론 포럼](https://github.com/siyul-park/uniflow/discussions): 질문하고 피드백을 나눌 수 있습니다.
- [이슈 트래커](https://github.com/siyul-park/uniflow/issues): 버그 신고 및 기능 요청이 가능합니다.

## 📜 라이선스

이 프로젝트는 [MIT 라이선스](./LICENSE)에 따라 배포됩니다. 자유롭게 사용, 수정, 배포할 수 있습니다.

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
[repo_ci_img]: https://github.com/siyul-park/uniflow/actions/workflows/ci.yml/badge.svg
[repo_ci_url]: https://github.com/siyul-park/uniflow/actions/workflows/ci.yml
