# 🪐 Uniflow

[![go report][go_report_img]][go_report_url]
[![go doc][go_doc_img]][go_doc_url]
[![release][repo_releases_img]][repo_releases_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]

**A high-performance, extremely flexible, and easily extensible universal workflow engine.**

## 📝 Overview

Efficiently manage complex data flows and adjust them in real-time. With the ability to add [plugins](plugins/README.md), workflows can be easily expanded to optimize processes according to business needs. Build a strong foundation for continuous system evolution and deliver personalized services.

## 🎯 Key Features

- **Performance:** Offers maximum throughput and minimal latency across various environments.
- **Flexibility:** Adjust and modify workflows in real-time.
- **Scalability:** Easily expand functionality by adding new components.

## 🚀 Quick Start

### 🛠️ Build and Installation

To build the source code, ensure you have **[Go 1.24 or above](https://go.dev/doc/install)** installed and run the
following commands:

```sh
git clone https://github.com/siyul-park/uniflow

cd uniflow

make init
make build-all
```

After building, the executable will be created in the `dist` directory.

### ⚙️ Configuration

You can manage configuration using environment variables or a `.uniflow.toml` file. Here's an example of registering and
configuring the built-in plugins:

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

### ⚡ Example Run

To run the [ping.yaml](examples/ping.yaml) example, which processes HTTP requests, use the following configuration:

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

Start the workflow with this command:

```sh
./dist/uniflow start --from-specs ./examples/ping.yaml --environment PORT=8000
```

To verify it's working, use the following command to call the HTTP endpoint:

```sh
curl localhost:8000/ping
pong#
```

## 📊 Benchmark

The following benchmark was run on a **[Contabo](https://contabo.com/)** VPS S SSD (4-core, 8GB) environment.  
The test was conducted using
the [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) on a workflow consisting
of `listener`, `router`, and `snippet` nodes, using the [ping.yaml](examples/ping.yaml) example.

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

- [Getting Started](./docs/getting_started.md): Learn about CLI installation and workflow management.
- [Core Concepts](./docs/key_concepts.md): Understand key concepts such as nodes, connections, ports, and packets.
- [Architecture](./docs/architecture.md): Explore the specification loading and workflow execution process.
- [Flowchart](./docs/flowchart.md): A step-by-step guide to compilation and runtime processes.
- [Debugging](./docs/debugging.md): Find troubleshooting and debugging tips.
- [User Extensions](./docs/user_extensions.md): Learn how to add new nodes and integrate services.

## 🌐 Community & Support

- [Discussion Forum](https://github.com/siyul-park/uniflow/discussions): Ask questions and share feedback.
- [Issue Tracker](https://github.com/siyul-park/uniflow/issues): Report bugs and request features.

## 📜 License

This project is distributed under the [MIT License](./LICENSE). You are free to use, modify, and distribute it.

<!-- Go -->

[go_download_url]: https://golang.org/dl/
[go_version_img]: https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go

[go_code_coverage_img]: https://codecov.io/gh/siyul-park/uniflow/graph/badge.svg?token=HOFm99R9SO
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
