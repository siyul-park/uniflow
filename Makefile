-include .env

CURRENT_DIR = $(shell realpath .)
MODULE_DIRS = $(shell find $(CURRENT_DIR) -name go.mod -exec dirname {} \;)
PLUGIN_DIRS = $(shell find $(CURRENT_DIR)/plugin -name go.mod -exec dirname {} \;)

DOCKER_IMAGE = $(shell basename -s .git $(shell git config --get remote.origin.url))
DOCKER_TAG = $(shell git tag --sort=-v:refname | grep -v '/' | head -n 1 || echo "latest")
DOCKERFILE = deployments/Dockerfile

.PHONY: init generate build clean tidy update check test coverage benchmark lint fmt vet doc docker-build
all: lint test build

init:
	@cp .go.work go.work
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install github.com/incu6us/goimports-reviser/v3@latest

generate:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go generate ./...; \
	done

build-all: build build-plugin

build:
	@go clean -cache
	@mkdir -p dist
	@cd cmd && go build -ldflags "-s -w" -o ../dist ./...

build-plugin:
	@mkdir -p dist
	@for dir in $(PLUGIN_DIRS); do \
  		cd $$dir && go test -buildmode=plugin -ldflags "-s -w" -o $(CURRENT_DIR)/dist .; \
	done

build-docker:
	docker build --no-cache \
		-t $(if $(DOCKER_DOMAIN),$(DOCKER_DOMAIN)/)$(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f $(DOCKERFILE) \
		--build-arg COPY_EXAMPLES=$(COPY_EXAMPLES) \
		$(CURRENT_DIR)

clean:
	@go clean -cache
	@rm -rf dist

tidy:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go mod tidy; \
	done

update:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go get -u all; \
	done

clean-sum:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && rm go.sum; \
	done

clean-cache:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go clean -modcache; \
	done

check: lint test staticcheck

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
		cd $$dir && goimports-reviser -rm-unused -format ./...; \
	done

vet:
	@for dir in $(MODULE_DIRS); do \
		cd $$dir && go vet ./...; \
	done

doc: init
	@godoc -http=:6060
