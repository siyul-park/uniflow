-include .env

.PHONY: init
init:
	@find . -name go.mod -execdir go install -v ./... \;

.PHONY: init-staticcheck
init-staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: init-godoc
init-godoc:
	@go install golang.org/x/tools/cmd/godoc@latest	

.PHONY: generate
generate:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go generate ./...'

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
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go mod tidy'

.PHONY: check
check: lint test

.PHONY: test
test:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test $(test-options) ./...'

.PHONY: race
race:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -race $(test-options) ./...'

.PHONY: coverage
coverage:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -coverprofile coverage.out -covermode count $(test-options) ./...'

.PHONY: benchmark
benchmark:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -run="-" -bench=".*" -benchmem $(test-options) ./...'

.PHONY: lint
lint: fmt vet staticcheck

.PHONY: vet
vet:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go vet ./...'

.PHONY: fmt
fmt:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go fmt ./...'

.PHONY: staticcheck
staticcheck: init-staticcheck
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; staticcheck ./...'

.PHONY: doc
doc: init-godoc
	@godoc -http=:6060
