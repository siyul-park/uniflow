# Uniflow

[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![go report][go_report_img]][go_report_url]
[![codefactor][repo_codefactor_img]][repo_codefactor_url]
[![release][repo_releases_img]][repo_releases_url]

번역:
  - [English](./README.md)

## 개요

이 엔진은 높은 성능과 극도의 유연성을 갖춘 쉽게 확장 가능한 다목적 워크플로 엔진입니다.

다양한 기간이 걸리는 작업을 효율적으로 처리하며 데이터 처리 흐름을 선언적으로 정의하는 환경을 제공합니다. 이를 통해 다양한 작업에서 최적의 성능, 낮은 대기 시간 및 높은 처리량을 달성할 수 있습니다.

내장된 확장 기능은 짧은 기간이 걸리는 작업을 효율적으로 실행하는 데 중점을 두고 다양한 기능을 제공합니다. 또한 필요에 따라 유연하게 기능을 확장할 수 있도록 설계되었습니다.

사용자에게 개인화된 경험을 제공하는 서비스를 개발하고 필요할 때 쉽게 기능을 확장해보세요.

## 원칙

- **성능:** 다양한 작업 부하에서 최적의 처리량, 최소 대기 시간 및 최대 확장성을 달성합니다.
- **유연성:** 변화하는 요구 사항에 신속하게 적응할 수 있도록 복잡한 데이터 처리 흐름을 선언적으로 정의하며, 동적 수정과 실시간 조정을 가능하게 합니다.
- **확장성:** 내장된 확장 기능을 활용하여 다양한 작업을 효율적으로 실행하고 필요에 따라 기능을 추가하거나 사용자 정의할 수 있습니다.

## 빠른 시작

[ping 예제](/examples/ping.yaml)를 실행하려면 다음 명령을 사용하세요:

```shell
./uniflow start --filename example/ping.yaml
```

`--filename` 플래그는 네임스페이스에 노드가 이미 존재하지 않는 경우 자동으로 설치합니다.

예상대로 HTTP 엔드포인트가 제공되는지 확인하세요:

```shell
curl localhost:8000/ping
pong#
```

실행 중인 서버에 노드를 적용하려면 `apply` 명령을 사용하세요.

추가 세부 정보는 다음 명령 도움말을 참조하세요:

```shell
./dist/uniflow start --help
```

## 구성

`.uniflow.toml` 또는 시스템 환경 변수를 사용하여 환경을 구성하세요.

| TOML 키         | Env 키           | 예시                   |
|------------------|------------------|-----------------------|
| `database.url`   | `DATABASE.URL`   | `mem://` 또는 `mongodb://` |
| `database.name`  | `DATABASE.NAME`  | -                     |

## 벤치마크

벤치마크 테스트는 [Contabo](https://contabo.com/)의 VPS S SSD (4코어, 8GB)에서 수행되었습니다. 성능은 [Apache HTTP 서버 벤치마킹 도구](https://httpd.apache.org/docs/2.4/programs/ab.html)를 이용해 루프백 네트워크 어댑터(127.0.0.1)를 통해 측정되었습니다. 테스트 워크플로는 `listener`, `router`, `snippet` 노드로 구성된 [ping 예제](/examples/ping.yaml)를 사용했습니다.

```sh
ab -n 102400 -c 1024 http://127.0.0.1:8000/ping
```

결과:

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

## 관련 자료

- [**문서**](/docs/README_kr.md)

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
