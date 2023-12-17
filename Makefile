-include .env

.PHONY: init
init:
	@find . -name go.mod -execdir go install -v ./... ';'

.PHONY: init-staticcheck
init-staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: init-godoc
init-godoc:
	@go install golang.org/x/tools/cmd/godoc@latest	

.PHONY: generate
generate:
	@find . -name go.mod -execdir go generate ./... ';'

.PHONY: build
build:
	@go clean -cache
	@mkdir -p dist
	@cd cmd && go build -ldflags "-s -w" -o ../dist ./...

.PHONY: clean
clean:
	@go clean -cache
	@rm -rf dist

.PHONY: tidy
tidy:
	@find . -name go.mod -execdir go mod tidy ';'

.PHONY: check
check: lint test

.PHONY: test
test:
	@find . -name go.mod -execdir go test $(test-options) ./... ';'

.PHONY: race
race:
	@find . -name go.mod -execdir go test -race $(test-options) ./... ';'

.PHONY: coverage
coverage:
	@find . -name go.mod -execdir go test -coverprofile coverage.out -covermode count ./... ';'
	@find . -name go.mod -execdir go tool cover -func=coverage.out | grep total ';'

.PHONY: benchmark
benchmark:
	@find . -name go.mod -execdir go test -run="-" -bench=".*" -benchmem ./... ';'

.PHONY: lint
lint: fmt vet staticcheck

.PHONY: vet
vet:
	@find . -name go.mod -execdir go vet ./... ';'

.PHONY: fmt
fmt:
	@find . -name go.mod -execdir go fmt ./... ';'

.PHONY: staticcheck
staticcheck: init-staticcheck
	@find . -name go.mod -execdir staticcheck ./... ';'

.PHONY: doc
doc: init-godoc
	@godoc -http=:6060
