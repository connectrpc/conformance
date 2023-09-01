# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN := .tmp/bin
COPYRIGHT_YEARS := 2022
LICENSE_IGNORE := -e proto/grpc -e internal/interopgrpc -e web/spec/grpc-web.spec.ts -e web/server/fastify/program.ts
# Set to use a different compiler. For example, `GO=go1.18rc1 make test`.
GO ?= go

.PHONY: help
help: ## Describe useful make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: ## Build, test, and lint (default)
	$(MAKE) test
	$(MAKE) lint

.PHONY: clean
clean: ## Delete intermediate build artifacts
	@# -X only removes untracked files, -d recurses into directories, -f actually removes files/dirs
	git clean -Xdf

.PHONY: test
test: build ## Run unit tests
	$(GO) test -vet=off -race -cover ./...

.PHONY: shorttest
shorttest: build ## Run unit tests
	$(GO) test -test.short -vet=off -race -cover ./...

.PHONY: build
build: generate ## Build all packages
	$(GO) build ./...

.PHONY: lint
lint: $(BIN)/golangci-lint $(BIN)/buf ## Lint Go and protobuf
	test -z "$$($(BIN)/buf format -d . | tee /dev/stderr)"
	$(GO) vet ./...
	$(BIN)/golangci-lint run
	$(BIN)/buf lint
	cd web; npm install; npm run lint

.PHONY: lintfix
lintfix: $(BIN)/golangci-lint $(BIN)/buf ## Automatically fix some lint errors
	$(BIN)/golangci-lint run --fix
	$(BIN)/buf format -w .

.PHONY: generate
generate: $(BIN)/buf $(BIN)/protoc-gen-go $(BIN)/protoc-gen-connect-go $(BIN)/protoc-gen-go-grpc $(BIN)/license-header $(BIN)/git-ls-files-unstaged ## Regenerate code and licenses
	rm -rf cc/gen
	rm -rf internal/gen
	rm -rf web/gen
	PATH=$(abspath $(BIN)) $(BIN)/buf generate
	$(BIN)/git-ls-files-unstaged | \
		grep -v $(patsubst %,-e %,$(sort $(LICENSE_IGNORE))) | \
		xargs $(BIN)/license-header \
			--license-type "apache" \
			--copyright-holder "Buf Technologies, Inc." \
			--year-range "$(COPYRIGHT_YEARS)"

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v

.PHONY: checkgenerate
checkgenerate:
	@# Used in CI to verify that `make generate` doesn't produce a diff.
	test -z "$$(git status --porcelain | tee /dev/stderr)"

.PHONY: dockercomposetestgo
dockercomposetestgo: dockercomposeclean
	docker-compose run client-connect-go-to-server-connect-go-h1
	docker-compose run client-connect-go-to-server-connect-go-h2
	docker-compose run client-connect-go-to-server-connect-go-h3
	docker-compose run client-connect-go-to-server-connect-node-fastify-h1
	docker-compose run client-connect-go-to-server-connect-node-fastify-h2
	docker-compose run client-connect-go-grpc-to-server-connect-node-fastify-h1
	docker-compose run client-connect-go-grpc-to-server-connect-node-fastify-h2
	docker-compose run client-connect-go-grpc-web-to-server-connect-node-fastify-h1
	docker-compose run client-connect-go-grpc-web-to-server-connect-node-fastify-h2
	docker-compose run client-connect-go-grpc-to-server-connect-go-h1
	docker-compose run client-connect-go-grpc-to-server-connect-go-h2
	docker-compose run client-connect-go-grpc-web-to-server-connect-go-h1
	docker-compose run client-connect-go-grpc-web-to-server-connect-go-h2
	docker-compose run client-connect-go-grpc-web-to-server-connect-go-h3
	docker-compose run client-connect-go-grpc-web-to-envoy-server-connect-go-h1
	docker-compose run client-connect-go-grpc-web-to-envoy-server-grpc-go-h1
	docker-compose run client-connect-go-grpc-web-to-envoy-server-connect-node-fastify-h1
	docker-compose run client-connect-go-grpc-to-server-grpc-go
	docker-compose run client-grpc-go-to-server-connect-go
	docker-compose run client-grpc-go-to-server-grpc-go
	docker-compose run client-grpc-go-to-server-connect-node-fastify
	$(MAKE) dockercomposeclean

.PHONY: dockercomposetestweb
dockercomposetestweb: dockercomposeclean
	docker-compose run client-connect-web-to-server-connect-go-h1
	docker-compose run client-connect-web-grpc-web-to-server-connect-go-h1
	docker-compose run client-connect-web-grpc-web-to-envoy-server-connect-go
	docker-compose run client-connect-web-grpc-web-to-envoy-server-grpc-go
	docker-compose run client-connect-web-grpc-web-to-envoy-server-connect-node-fastify
	docker-compose run client-grpc-web-to-server-connect-go-h1
	docker-compose run client-grpc-web-to-envoy-server-connect-go
	docker-compose run client-grpc-web-to-envoy-server-grpc-go
	docker-compose run client-grpc-web-to-envoy-server-connect-node-fastify
	docker-compose run client-connect-web-to-server-connect-node-fastify-h1
	docker-compose run client-connect-web-grpc-web-to-server-connect-node-fastify-h1
	docker-compose run client-grpc-web-to-server-connect-node-fastify-h1
	$(MAKE) dockercomposeclean

.PHONY: dockercomposetestcc
dockercomposetestcc: dockercomposeclean
	docker-compose run client-cc-grpc-to-server-connect-go-h2-no-tls
	docker-compose run client-cc-grpc-to-server-connect-go-h2-with-tls
	$(MAKE) dockercomposeclean

.PHONY: dockercomposetest
dockercomposetest:
	$(MAKE) dockercomposetestgo
	$(MAKE) dockercomposetestweb

.PHONY: dockercomposeclean
dockercomposeclean:
	-docker-compose down --rmi local --remove-orphans

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/buf/cmd/buf@v1.24.0

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install \
		  github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.24.0

$(BIN)/git-ls-files-unstaged: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(BIN)) go install github.com/bufbuild/buf/private/pkg/git/cmd/git-ls-files-unstaged@v1.1.0

$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.0

$(BIN)/protoc-gen-connect-go: Makefile go.mod
	@mkdir -p $(@D)
	@# Pinned by go.mod.
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go

$(BIN)/protoc-gen-go-grpc: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

$(BIN)/protoc-gen-go: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
