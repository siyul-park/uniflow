# Uniflow

[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![go report][go_report_img]][go_report_url]
[![codefactor][repo_codefactor_img]][repo_codefactor_url]
[![release][repo_releases_img]][repo_releases_url]

An ultra-fast, highly flexible, and easily customizable multipurpose workflow engine.

Efficiently manage tasks across all lifespans—from short-term to long-term—using this engine. It provides a declarative environment for defining data processing flows, ensuring high performance and low latency across diverse operations.

The built-in extensions are designed to prioritize short-term processing tasks while also supporting a diverse range of functionalities. Additionally, you can seamlessly integrate and customize additional features into the engine as needed.

## Principles

- **High Performance:** Achieve optimal throughput and minimal latency, scaling seamlessly across diverse workloads.

- **Flexibility:** Define complex data processing flows declaratively to adapt effortlessly to changing requirements, dynamically modifying and reflecting flow adjustments.

- **Extensibility:** Utilize built-in extensions for efficient execution of diverse tasks and seamlessly integrate or customize additional functionalities.

## Getting Started

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

| TOML Key         | Env Key          | Example          |
|------------------|------------------|------------------|
| `database.url`   | `DATABASE.URL`   | `mem://` or `mongodb://` |
| `database.name`  | `DATABASE.NAME`  | -                |

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
