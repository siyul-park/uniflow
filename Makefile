-include .env

CURRENT_DIR = $(shell realpath .)

.PHONY: init
init:
	@cp .go.work go.work
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go install -v ./...'
	@go install honnef.co/go/tools/cmd/staticcheck@latest
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

.PHONY: update
update:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go get -u all && go mod tidy'

.PHONY: sync
sync:
	@go work sync

.PHONY: check
check: lint test

.PHONY: test
test:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -race $(test-options) ./...'

.PHONY: coverage
coverage:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -race --coverprofile=coverage.out --covermode=atomic $(test-options) ./...'
	@find $(realpath .) -name go.mod | xargs dirname | grep '${CURRENT_DIR}/' | xargs -I {} sh -c 'cd {}; cat coverage.out >> '${CURRENT_DIR}/coverage.out' && rm coverage.out'

.PHONY: benchmark
benchmark:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go test -run="-" -bench=".*" -benchmem $(test-options) ./...'

.PHONY: lint
lint: fmt vet staticcheck

.PHONY: fmt
fmt:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go fmt ./...'

.PHONY: vet
vet:
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; go vet ./...'

.PHONY: staticcheck
staticcheck: init
	@find $(realpath .) -name go.mod | xargs dirname | xargs -I {} sh -c 'cd {}; staticcheck ./...'

.PHONY: doc
doc: init
	@godoc -http=:6060
