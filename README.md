# Uniflow

[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![go report][go_report_img]][go_report_url]
[![codefactor][repo_codefactor_img]][repo_codefactor_url]
[![release][repo_releases_img]][repo_releases_url]

Translations:
  - [한국어](./README_kr.md)

## Overview

A high-performance, extremely flexible, and easily extensible multipurpose workflow engine.

It efficiently handles tasks of varying durations, from short-term to long-term, providing an environment where data processing flows can be declaratively defined and dynamically modified.

Through [built-in extensions](./ext/README.md), it efficiently executes short-term tasks and implements a wide range of features. However, the engine does not enforce the use of specific nodes; all nodes can be freely added or removed according to the service requirements.

Integrate the engine into your service to offer personalized experiences and extensively expand functionalities.

## Principles

- **Performance:** Achieves maximum throughput and minimum latency across diverse environments.
- **Flexibility:** Allows dynamic modification of specifications and real-time adjustments.
- **Extensibility:** Supports the addition of new nodes freely, enabling extensive feature expansion.

## Quick Start

Run the [ping example](./examples/ping.yaml) using the following command:

```shell
./uniflow start --filename example/ping.yaml
```

Verify that the HTTP endpoint works as expected:

```shell
curl localhost:8000/ping
pong#
```

## Configuration

Configure the environment using the `.uniflow.toml` file or system environment variables.

| TOML Key         | Environment Variable Key | Example                       |
|------------------|--------------------------|-------------------------------|
| `database.url`   | `DATABASE.URL`           | `mem://` or `mongodb://`      |
| `database.name`  | `DATABASE.NAME`          | -                             |

## Benchmark

Benchmarks were conducted on a Contabo VPS S SSD (4 cores, 8GB RAM) using the [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html). The workflow consisted of the `listener`, `router`, and `snippet` nodes from the [ping example](./examples/ping.yaml).

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

## Learn More

- [Getting Started](./docs/getting_started.md): Learn how to install the CLI, manage workflows, and run the engine.
- [Architecture](./docs/architecture.md): Understand how node specifications are loaded into the engine and how workflows are executed.

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
