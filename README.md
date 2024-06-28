# Uniflow

[![check][repo_check_img]][repo_check_url]
[![code coverage][go_code_coverage_img]][go_code_coverage_url]
[![go report][go_report_img]][go_report_url]
[![codefactor][repo_codefactor_img]][repo_codefactor_url]
[![release][repo_releases_img]][repo_releases_url]

> Low-Code Engine for Backend Workflows

Uniflow is a low-code engine that enables fast and efficient construction and execution of backend workflows.

## Getting Started

[Download Go][go_download_url] and install (version `1.21` or higher is required).

### Installation

To integrate into your project, use the following import statements:

```shell
go get -u github.com/siyul-park/uniflow
go get -u github.com/siyul-park/uniflow/extend
```

### Usage Example

Here's a basic example for incorporating it into your application:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/extend/pkg/control"
	"github.com/siyul-park/uniflow/extend/pkg/network"
)

func main() {
	hb := hook.NewBuilder()
	sb := scheme.NewBuilder()

	hb.Register(network.AddToHook())
	sb.Register(control.AddToScheme())
	sb.Register(network.AddToScheme())
	
	hk, _ := hb.Build()
	sc, _ := sb.Build()

	db := memdb.New("")

	r, _ := New(context.TODO(), runtime.Config{
		Hook:     hk,
		Scheme:   sc,
		Database: db,
	})
	defer r.Close()

	r.Start(context.TODO())
}
```

Customize the code according to your specific requirements.

### Building

To build the project, follow these steps:

**Clone the Repository:**
```shell
git clone https://github.com/siyul-park/uniflow
cd uniflow
```

**Initialize and Build the Project:**
```shell
make init
make build
```
These commands initialize the project, set up dependencies, and compile the source code to generate the executable.

**Verify the Result:**
```shell
ls /dist
uniflow
```
Ensure the executable named "uniflow" is present in the `/dist` directory.

Following these steps ensures that the project is properly set up, built, and the executable is available for use in the `/dist` directory.

### Starting

Now ready to be used. To initiate the [ping](/examples/ping.yaml) example, execute the following command:

```shell
./dist/uniflow start --filename example/ping.yaml
```

The `--filename` flag automatically installs if the node does not exist in the namespace.

Check if the instance is providing an HTTP endpoint as expected:

```shell
curl localhost:8000/ping
pong#
```

If you want to apply nodes to a running server, utilize the `apply` command.

For additional details, run the following command:

```shell
./dist/uniflow start --help
```

### Configuration

You can set environment variables before executing any command.

Configuration can be done using `.uniflow.toml` or system environment variables.

| TOML Key | Env Key | Example |
|---|---|---|
| `database.url` | `DATABASE.URL` | `mem://` or `mongodb://` |
| `database.name` | `DATABASE.NAME` | - |

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
