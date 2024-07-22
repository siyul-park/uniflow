# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**A high-performance, extremely flexible, and easily extensible universal workflow engine.**

## 📝 Overview

**Uniflow** efficiently handles a wide range of tasks from short-term jobs to long-term processes. It allows for declarative definition and dynamic modification of data flows, leveraging [built-in extension capabilities](./ext/README.md) to easily implement complex workflows. Moreover, it offers flexibility to expand functionality by adding new nodes or removing existing ones as needed.

Provide a personalized experience through your service and consistently expand its capabilities.

## 🎯 Core Values

- **Performance:** Achieve optimal throughput and minimal latency across various environments.
- **Flexibility:** Dynamically modify and adjust workflows in real-time.
- **Extensibility:** Extend system functionality through new components.

## 🚀 Quick Start

### 🛠️ Build and Install

**[Go 1.22](https://go.dev/doc/install)** or higher is required. Follow these steps to build from source:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build
```

After building, the executable will be located in the `dist` directory.

### ⚡ Run an Example

Let's run a basic HTTP request handling example, [ping.yaml](./examples/ping.yaml):

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

To run the workflow, use the following command:

```sh
uniflow start --filename example/ping.yaml
```

To verify it's working, call the HTTP endpoint:

```sh
curl localhost:8000/ping
pong#
```

## ⚙️ Configuration

You can configure settings through the `.uniflow.toml` file or system environment variables.

| TOML Key           | Environment Variable | Example                   |
|--------------------|----------------------|---------------------------|
| `database.url`     | `DATABASE.URL`       | `mem://` or `mongodb://`  |
| `database.name`    | `DATABASE.NAME`      | -                         |
| `collection.nodes` | `COLLECTION.NODES`   | `nodes`                   |

## 📊 Benchmark

Below are benchmark results performed on a **[Contabo](https://contabo.com/)** VPS S SSD (4 cores, 8GB) environment. We used the [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) to measure the [ping.yaml](./examples/ping.yaml) workflow consisting of `listener`, `router`, and `snippet` nodes.

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

## 📚 Learn More

- [Getting Started](./docs/getting_started.md): Introduction to CLI installation and workflow management.
- [Key Concepts](./docs/key_concepts.md): Explanation of basic concepts such as nodes, connections, ports, and packets.
- [Architecture](./docs/architecture.md): Detailed explanation of node specification loading and workflow execution process.
- [User Extensions](./docs/user_extensions.md): Guide on adding new features and integrating existing services.

## 🌐 Community and Support

For questions about the project or if you need support, please use the following channels:

- [Discussion Forum](https://github.com/siyul-park/uniflow/discussions): Share questions and feedback.
- [Issue Tracker](https://github.com/siyul-park/uniflow/issues): Submit bug reports and feature requests.

## 📜 License

This project is distributed under the [MIT License](./LICENSE). You are free to use, modify, and distribute it according to the license terms.

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
