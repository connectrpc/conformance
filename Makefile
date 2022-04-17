# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN=.tmp/bin
# Set to use a different compiler. For example, `GO=go1.18rc1 make test`.
GO ?= go
COPYRIGHT_YEARS := 2022
# Which commit of bufbuild/makego should we source checknodiffgenerated.bash
# from?
MAKEGO_COMMIT := 383cdab9b837b1fba0883948ff54ed20eedbd611
LICENSE_IGNORE := -e internal/proto/grpc -e internal/interop/grpc -e grpcweb_client_test.js

.PHONY: help
help: ## Describe useful make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: ## Build, test, and lint (default)
	$(MAKE) test
	$(MAKE) lint
	$(MAKE) checkgenerate

.PHONY: clean
clean: ## Delete intermediate build artifacts
	@# -X only removes untracked files, -d recurses into directories, -f actually removes files/dirs
	git clean -Xdf

.PHONY: test
test: build ## Run unit tests
	$(GO) test -vet=off -race -cover ./...

.PHONY: build
build: generate ## Build all packages
	$(GO) build ./...

.PHONY: lint
lint: $(BIN)/gofmt $(BIN)/buf ## Lint Go and protobuf
	test -z "$$($(BIN)/gofmt -s -l . | tee /dev/stderr)"
	test -z "$$($(BIN)/buf format -d . | tee /dev/stderr)"
	@# TODO: replace vet with golangci-lint when it supports 1.18
	@# Configure staticcheck to target the correct Go version and enable
	@# ST1020, ST1021, and ST1022.
	$(GO) vet ./...
	@# We only have vendored protobuf for now, and it fails all lint checks.
	@# $(BIN)/buf lint --exclude-path proto/grpc

.PHONY: lintfix
lintfix: $(BIN)/gofmt $(BIN)/buf ## Automatically fix some lint errors
	$(BIN)/gofmt -s -w .
	$(BIN)/buf format -w .

.PHONY: generate
generate: $(BIN)/buf $(BIN)/protoc-gen-go $(BIN)/protoc-gen-connect-go $(BIN)/protoc-gen-go-grpc $(BIN)/protoc-gen-es $(BIN)/protoc-gen-connect-web $(BIN)/license-header ## Regenerate code and licenses
	rm -rf internal/gen
	rm -rf web/gen
	PATH=$(BIN) $(BIN)/buf generate
	@# We want to operate on a list of modified and new files, excluding
	@# deleted and ignored files. git-ls-files can't do this alone. comm -23 takes
	@# two files and prints the union, dropping lines common to both (-3) and
	@# those only in the second file (-2). We make one git-ls-files call for
	@# the modified, cached, and new (--others) files, and a second for the
	@# deleted files.
	@$(BIN)/license-header \
		--license-type apache \
		--copyright-holder "Buf Technologies, Inc." \
		--year-range "$(COPYRIGHT_YEARS)" \
		$(shell comm -23 \
			<(git ls-files --cached --modified --others --no-empty-directory --exclude-standard | sort -u | grep -v $(LICENSE_IGNORE)) \
			<(git ls-files --deleted | sort -u))

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v

.PHONY: checkgenerate
checkgenerate: $(BIN)/checknodiffgenerated.bash
	$(BIN)/checknodiffgenerated.bash $(MAKE) generate

$(BIN)/gofmt:
	@mkdir -p $(@D)
	$(GO) build -o $(@) cmd/gofmt

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/buf/cmd/buf@v1.2.1

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install \
		  github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.2.1

$(BIN)/protoc-gen-connect-go: Makefile
	@mkdir -p $(@D)
	@# Pinned by go.mod.
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go

$(BIN)/protoc-gen-go-grpc: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

$(BIN)/protoc-gen-go: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1

$(BIN)/protoc-gen-es: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/protobuf-es/cmd/protoc-gen-es@v0.0.0-20220404100843-2bf5c0f2d1c3

$(BIN)/protoc-gen-connect-web: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/connect-web/cmd/protoc-gen-connect-web@v0.0.0-20220407075159-6fda16455846

$(BIN)/checknodiffgenerated.bash:
	@mkdir -p $(@D)
	curl -SsLo $(@) https://raw.githubusercontent.com/bufbuild/makego/$(MAKEGO_COMMIT)/make/go/scripts/checknodiffgenerated.bash
	chmod u+x $(@)

docker-compose-clean:
	docker-compose down --rmi local --remove-orphans

test-docker-compose: docker-compose-clean
	@# docker build is a work around for the --ssh as it is not yet supported by docker-compose (github.com/docker/compose/issues/7025), can be removed when either it is supported or connect-go become public
	docker build --ssh default -f Dockerfile.crosstest .
	docker-compose run client-connect-to-connect
	docker-compose run client-connect-to-grpc
	docker-compose run client-grpc-to-connect
	docker-compose run client-grpc-to-grpc
	docker-compose run client-grpc-web-to-connect-h1
	docker-compose run client-grpc-web-to-envoy-connect
	docker-compose run client-grpc-web-to-envoy-grpc
	$(MAKE) docker-compose-clean
