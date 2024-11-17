-include .env

CURRENT_DIR = $(shell realpath .)
MODULE_DIRS = $(shell find $(CURRENT_DIR) -name go.mod -exec dirname {} \;)

DOCKER_IMAGE = $(shell basename -s .git $(shell git config --get remote.origin.url))
DOCKER_TAG = $(shell git tag --sort=-v:refname | grep -v '/' | head -n 1 || echo "latest")
DOCKERFILE = deployments/Dockerfile

CGO_ENABLED ?= 1

.PHONY: init generate build clean tidy update sync check test coverage benchmark lint fmt vet doc docker-build

init:
	@cp .go.work go.work
	@$(MAKE) install-tools
	@$(MAKE) install-modules

install-tools:
	@go install golang.org/x/tools/cmd/godoc@latest

install-modules:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go install -v ./...; \
	done

generate:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go generate ./...; \
	done

build:
	@go clean -cache
	@mkdir -p dist
	@cd cmd && CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "-s -w" -o ../dist ./...

clean:
	@go clean -cache
	@rm -rf dist

tidy:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go mod tidy; \
	done

update:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go get -u all && go mod tidy; \
	done

sync:
	@go work sync

check: lint test

test:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go test -race $(test-options) ./...; \
	done

coverage:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go test -race --coverprofile=coverage.out --covermode=atomic $(test-options) ./...; \
		cat coverage.out >> $(CURRENT_DIR)/coverage.out && rm coverage.out; \
	done

benchmark:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go test -run="-" -bench=".*" -benchmem $(test-options) ./...; \
	done

lint: fmt vet

fmt:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go fmt ./...; \
	done

vet:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go vet ./...; \
	done

doc: init
	@godoc -http=:6060

docker-build:
	docker build --no-cache \
		-t $(if $(DOCKER_DOMAIN),$(DOCKER_DOMAIN)/)$(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f $(DOCKERFILE) \
		--build-arg COPY_EXAMPLES=$(COPY_EXAMPLES) \
		$(CURRENT_DIR)
