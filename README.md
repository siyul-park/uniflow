# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![ci][repo_ci_img]][repo_ci_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**A high-performance, extremely flexible, and easily extensible universal workflow engine.**

## 📝 Overview

**Uniflow** is designed to manage a wide range of tasks, from short-term jobs to long-term processes. It supports declarative workflow definitions and allows for dynamic changes to data flows. With [built-in extensions](./ext/README.md), you can implement complex workflows and add or remove nodes to expand its functionality as needed.

This system empowers you to deliver customized experiences through your service and continuously enhance its capabilities.

## 🎯 Core Values

- **Performance:** Optimized for maximum throughput and minimal latency.
- **Flexibility:** Adapt workflows on-the-fly to meet evolving needs.
- **Extensibility:** Easily integrate new components and expand your system’s functionality.

## 🚀 Quick Start

### 🛠️ Build and Install

To get started, install **[Go 1.23](https://go.dev/doc/install)** or later. Then, follow these steps:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build
```

The executable will be located in the `dist` directory after building.

### ⚡ Run an Example

Try a basic HTTP request handler using [ping.yaml](./examples/ping.yaml):

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

Start the workflow with:

```sh
./dist/uniflow start --from-specs ./examples/ping.yaml --env PORT=8000
```

Verify by calling the HTTP endpoint:

```sh
curl localhost:8000/ping
pong#
```

## ⚙️ Configuration

Adjust settings through `.uniflow.toml` or environment variables.

| TOML Key            | Environment Variable Key | Example                  |
|---------------------|--------------------------|--------------------------|
| `database.url`      | `DATABASE_URL`           | `mem://` or `mongodb://` |
| `database.name`     | `DATABASE_NAME`          | -                        |
| `collection.specs`  | `COLLECTION_SPECS`       | `specs`                  |
| `collection.values` | `COLLECTION_VALUES`      | `values`                 |

## 📊 Benchmark

The following benchmark was conducted on a **[Contabo](https://contabo.com/)** VPS S SSD (4 cores, 8GB) using the [Apache HTTP benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) to measure the performance of [ping.yaml](./examples/ping.yaml) involving `listener`, `router`, and `snippet` nodes.

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

## 📚 Learn More

- [Getting Started](./docs/getting_started.md): CLI usage and workflow management.
- [Key Concepts](./docs/key_concepts.md): Understanding nodes, ports, and packets.
- [Architecture](./docs/architecture.md): Workflow execution and node specification loading.
- [Debugging](./docs/debugging.md): Debug workflows, set breakpoints, and start sessions.
- [User Extensions](./docs/user_extensions.md): Add features and integrate external services.
- [Kubernetes Operator](https://github.com/siyul-park/uniflow-operator): Detects workflow changes and generates CRD to
  provide serverless services.

## 🌐 Community and Support

- [Discussion Forum](https://github.com/siyul-park/uniflow/discussions): Share questions and feedback.
- [Issue Tracker](https://github.com/siyul-park/uniflow/issues): Submit bugs or request features.

## 📜 License

This project is available under the [MIT License](./LICENSE). You are free to use, modify, and distribute it in accordance with the terms.

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