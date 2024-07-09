# Uniflow

[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![go report][go_report_img]][go_report_url]
[![codefactor][repo_codefactor_img]][repo_codefactor_url]
[![release][repo_releases_img]][repo_releases_url]

번역:
  - [English](./README.md)

## 개요

높은 성능과 극도의 유연성을 갖춘 다목적 워크플로 엔진으로, 쉽게 확장할 수 있습니다.

짧은 기간이 걸리는 작업부터 긴 기간이 걸리는 작업까지 다양한 기간이 소요되는 작업을 효율적으로 처리하며, 데이터 처리 흐름을 선언적으로 정의하고 동적으로 수정할 수 있는 환경을 제공합니다.

[내장된 확장 기능](./ext/README_kr.md)을 통해 짧은 기간이 걸리는 작업을 효율적으로 실행하고 다양한 기능을 구현할 수 있습니다. 하지만 엔진은 특정 노드의 사용을 강제하지 않으며, 모든 노드는 서비스에 맞게 자유롭게 추가하거나 제거할 수 있습니다.

엔진을 서비스에 통합하여 사용자에게 개인화된 경험을 제공하고 기능을 풍부하게 확장해보세요.

## 핵심 가치

- **성능:** 다양한 환경에서 최대 처리량과 최소 대기 시간을 달성합니다.
- **유연성:** 명세를 동적으로 수정하고 실시간으로 조정할 수 있습니다.
- **확장성:** 새로운 노드를 자유롭게 지원하여 기능을 확장할 수 있습니다.

## 빠른 시작

아래 명령을 사용하여 [ping 예제](./examples/ping.yaml)를 실행해 보세요:

```shell
uniflow start --filename example/ping.yaml
```

이제 HTTP 엔드포인트가 예상대로 작동하는지 확인하세요:

```shell
curl localhost:8000/ping
pong#
```

## 구성

`.uniflow.toml` 파일 또는 시스템 환경 변수를 사용하여 환경을 구성하세요.

| TOML 키            | 환경 변수 키          | 예시                       |
|--------------------|--------------------|---------------------------|
| `database.url`     | `DATABASE.URL`     | `mem://` 또는 `mongodb://` |
| `database.name`    | `DATABASE.NAME`    | -                         |
| `collection.nodes` | `COLLECTION.NODES` | `nodes`                   |

## 벤치마크

벤치마크는 [Contabo](https://contabo.com/)의 VPS S SSD (4코어, 8GB)에서 수행되었으며, [아파치 웹서버 성능검사 도구](https://httpd.apache.org/docs/2.4/programs/ab.html)를 이용해 측정되었습니다. 워크플로우는 `listener`, `router`, `snippet` 노드로 구성된 [ping 예제](./examples/ping.yaml)를 사용했습니다.

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

## 자세히 알아보기

- [시작하기](./docs/getting_started_kr.md): CLI를 설치하고 워크플로우를 관리하며 엔진을 실행하는 방법을 살펴보세요.
- [핵심 개념](./docs/key_concepts_kr.md): 데이터 처리 객체인 노드와 그들 간의 연결, 포트, 패킷 등 핵심 개념에 대해 확인해보세요.
- [아키텍처](./docs/architecture_kr.md): 노드 명세가 엔진에 어떻게 로드되고 워크플로우가 어떻게 실행되는지 알아보세요.
- [사용자 확장 기능](./docs/user_extension_kr.md): 새로운 기능을 제공하는 노드를 추가하고 기존 서비스에 통합하는 방법을 자세히 알아보세요.

<!-- Go -->

[go_download_url]: https://golang.org/dl/
[go_version_img]: https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go
[go_code_coverage_img]: https://codecov.io/gh/siyul-park/uniflow/graph/badge.svg?token=quEl9AbBcW
[go_code_coverage_url]: https://codecov.io/gh/siyul-park/uniflow
[go_report_img]: https://goreportcard.com/badge/github.com/siyul-park/uniflow
[go_report_url]: https://goreportcard.com/report/github.com/siyul-park/uniflow

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
[repo_codefactor_img]: https://www.codefactor.io/repository/github/siyul-park/uniflow/badge
[repo_codefactor_url]: https://www.codefactor.io/repository/github/siyul-park/uniflow
