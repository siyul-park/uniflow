# Uniflow

[![go report][go_report_img]][go_report_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![check][repo_check_img]][repo_check_url]
[![release][repo_releases_img]][repo_releases_url]

> Low-Code Engine for Backend Workflows

Uniflow is a low-code engine that enables fast and efficient construction and execution of backend workflows.

## Getting Started
### Installation
1. [Download Go][go_download_url]: Install **Go** (version `1.21` or higher is required).
2. Clone the repository and initialize:
   ```shell
   git clone https://github.com/siyul-park/uniflow
   cd uniflow
   make init
   ```

### Build
1. Build the project:
   ```shell
   make build
   ```
2. Check the build result:
   ```shell
   ls /dist
   uniflow
   ```
3. Run tests:
   ```shell
   make test
   ```

### Start
Uniflow is now ready to be used. Let's start the [ping](/examples/ping.yaml) example.

```shell
./dist/uniflow start --boot example/ping.yaml
```
The `--boot` flag installs initially if the node does not exist in the namespace.

Check if the started Uniflow is providing an HTTP endpoint normally.

```shell
curl localhost:8000/ping
pong#
```

If you wish to apply nodes to a running server, use the `apply` command.

For more information, run the following command:
```shell
./dist/uniflow start --help
```

### Configuration
You can set environment variables before executing any command.

Configuration can be done using `.uniflow.toml` or system environment variables.

| TOML Key | Env Key | Default |
|---|---|---|
| database.url | DATABASE.URL | memdb:// |
| database.name | DATABASE.NAME |  |

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
