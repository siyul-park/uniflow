# uniflow

[![go report][go_report_img]][go_report_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![check][repo_check_img]][repo_check_url]
[![release][repo_releases_img]][repo_releases_url]

> Create your uniflow and integrate it anywhere!
  
Uniflow is a low-code engine for the backend. You can connect the nodes to create a flow and run it.

## Getting Started
### Installation
First, [download][go_download_url] and install **Go**. Version `1.21` or higher is required.
  
Clone the repository by using the `git clone` command:
```shell
git clone https://github.com/siyul-park/uniflow
```

And then init the project:
```shell
cd uniflow
make init
```
  
### Build

Build the project using the following command:
```shell
make build
```

The build result is created in the `/dist`.
```shell
ls /dist
uniflow
```

If you want to test the project. then run the following command:
```shell
make test
```

### Configuration
Before use any command. You can configure environment variables.

You can set environment variables to use `.uniflow.toml` or system environment variables.

| TOML Key | Env Key | Default |
|---|---|---|
| database.url | DATABASE.URL | memdb:// |
| database.name | DATABASE.NAME |  |

### Start

Uniflow is now ready to be used. Let's start the [ping](/examples/ping.yaml).

To start uniflow, using the following command:
```shell
./dist/uniflow start --boot example/ping.yaml
```
`--boot` is install initially if the node does not exist in namespace.

Let's check if the started uniflow is providing a http endpoint normally.
```shell
curl localhost:8000/ping
pong#
```

If you wish to apply nodes to a running server, use the `apply`.

Run the following command for more information.
```shell
./dist/uniflow start --help
```

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
[repo_check_img]: https://github.com/siyual-park/uniflow/actions/uniflows/check.yml/badge.svg
[repo_check_url]: https://github.com/siyual-park/uniflow/actions/uniflows/check.yml
