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

It efficiently manages tasks of varying durations, from short-term to long-term, providing a simple environment for defining data processing flows. This ensures optimal performance with low latency and high throughput across various operations.

The built-in extensions are crafted for efficient execution of short-term tasks, offering a wide array of functionalities. Furthermore, they facilitate seamless integration of additional features, allowing for flexible expansion as needed.

Develop a service that integrates user personalization, with the added benefit of easily expanding functionality as needed.

## Principles

- **Performance:** Achieve optimal throughput, minimal latency, and maximum scalability across diverse workloads.
- **Flexibility:** Define complex data processing flows declaratively to adapt seamlessly to changing requirements, enabling dynamic modifications and real-time adjustments.
- **Extensibility:** Utilize the built-in extensions to efficiently execute various tasks, seamlessly integrating or customizing additional functionalities as needed.

## Quick Start

To run the [ping example](/examples/ping.yaml), use this command:

```shell
./uniflow start --filename example/ping.yaml
```

The `--filename` flag automatically installs the node if it doesn't already exist in the namespace.

Check if the instance is providing the expected HTTP endpoint:

```shell
curl localhost:8000/ping
pong#
```

To apply nodes to a running server, use the `apply` command.

For additional details, refer to the command help:

```shell
./dist/uniflow start --help
```

## Configuration

Configure the environment using either `.uniflow.toml` or system environment variables.

| TOML Key         | Env Key          | Example               |
|------------------|------------------|-----------------------|
| `database.url`   | `DATABASE.URL`   | `mem://` or `mongodb://` |
| `database.name`  | `DATABASE.NAME`  | -                     |

## Benchmarks

The benchmarking tests were conducted using a VPS S SSD (4 Core, 8 GB) from [Contabo](https://contabo.com/). Performance was measured with the [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) over the loopback network adapter (127.0.0.1). The test workflow used the [ping example](/examples/ping.yaml), consisting of `listener`, `router`, and `snippet` nodes.

```sh
ab -n 102400 -c 1024 http://127.0.0.1:8000/ping
```

Results:

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

## Links

- [**Documentation**](/docs/README.md)

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
