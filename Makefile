# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN := .tmp/bin
export PATH := $(BIN):$(PATH)
export GOBIN := $(abspath $(BIN))
COPYRIGHT_YEARS := 2023
LICENSE_IGNORE := -e internal/gen -e _legacy -e testdata/
# Set to use a different compiler. For example, `GO=go1.18rc1 make test`.
GO ?= go
LATEST_VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null)
CURRENT_VERSION = $(shell git describe --tags --always --dirty)
# If not on release tag, this is a dev build. Add suffix to version.
ifneq ($(CURRENT_VERSION), $(LATEST_VERSION))
	DEV_BUILD_VERSION_DIRECTIVE = buildVersionSuffix=-$(shell git describe --exclude '*' --always --dirty)
else
	DEV_BUILD_VERSION_DIRECTIVE = buildVersion=$(CURRENT_VERSION)
endif
DEV_BUILD_VERSION_FLAG = -ldflags '-X "connectrpc.com/conformance/internal.$(DEV_BUILD_VERSION_DIRECTIVE)"'

.PHONY: help
help: ## Describe useful make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: ## Build, test, and lint (default)
	$(MAKE) test
	$(MAKE) lint
	$(MAKE) runconformance

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

.PHONY: generate
generate: $(BIN)/buf $(BIN)/license-header ## Regenerate code and licenses
	rm -rf internal/gen
	buf generate proto
	license-header \
		--license-type apache \
		--copyright-holder "The Connect Authors" \
		--year-range "$(COPYRIGHT_YEARS)" $(LICENSE_IGNORE)

.PHONY: lint
lint: $(BIN)/golangci-lint $(BIN)/buf ## Lint Go and protobuf
	test -z "$$($(BIN)/buf format -d . | tee /dev/stderr)"
	$(GO) vet ./...
	golangci-lint run
	buf lint proto

.PHONY: lintfix
lintfix: $(BIN)/golangci-lint $(BIN)/buf ## Automatically fix some lint errors
	golangci-lint run --fix
	buf format -w proto

.PHONY: install
install: ## Install all binaries
	$(GO) install $(DEV_BUILD_VERSION_FLAG) ./...

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	$(GO) get -u -t ./... && $(GO) mod tidy -v

.PHONY: checkgenerate
checkgenerate:
	@# Used in CI to verify that `make generate` doesn't produce a diff.
	test -z "$$(git status --porcelain | tee /dev/stderr)"

.PHONY: release
release: $(BIN)/goreleaser
	goreleaser release

.PHONY: checkrelease
checkrelease: $(BIN)/goreleaser
	# skips some validation and doesn't actually publish a release, just to test
	# that building a release works
	goreleaser release --clean --snapshot

.PHONY: runconformance
runconformance: runservertests runclienttests

.PHONY: runservertests
runservertests: $(BIN)/connectconformance $(BIN)/referenceserver $(BIN)/grpcserver
	$(BIN)/connectconformance -v --conf ./testdata/reference-impls-config.yaml --mode server -- $(BIN)/referenceserver
	$(BIN)/connectconformance -v --conf ./testdata/grpc-impls-config.yaml --mode server -- $(BIN)/grpcserver
	$(BIN)/connectconformance -v --conf ./testdata/grpc-web-server-impl-config.yaml --mode server -- $(BIN)/grpcserver

.PHONY: runclienttests
runclienttests: $(BIN)/connectconformance $(BIN)/referenceclient $(BIN)/grpcclient
	$(BIN)/connectconformance -v --conf ./testdata/reference-impls-config.yaml --mode client -- $(BIN)/referenceclient
	$(BIN)/connectconformance -v --conf ./testdata/grpc-impls-config.yaml --mode client -- $(BIN)/grpcclient

.PHONY: rungrpcweb
rungrpcweb: generate $(BIN)/connectconformance
	cd grpcwebclient && npm run conformance

$(BIN)/connectconformance: Makefile generate
	$(GO) build $(DEV_BUILD_VERSION_FLAG) -o $(@) ./cmd/connectconformance/

$(BIN)/referenceclient: Makefile generate
	$(GO) build $(DEV_BUILD_VERSION_FLAG) -o $(@) ./cmd/referenceclient/

$(BIN)/referenceserver: Makefile generate
	$(GO) build $(DEV_BUILD_VERSION_FLAG) -o $(@) ./cmd/referenceserver/

$(BIN)/grpcclient: Makefile generate
	$(GO) build $(DEV_BUILD_VERSION_FLAG) -o $(@) ./cmd/grpcclient/

$(BIN)/grpcserver: Makefile generate
	$(GO) build $(DEV_BUILD_VERSION_FLAG) -o $(@) ./cmd/grpcserver/

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	$(GO) install github.com/bufbuild/buf/cmd/buf@v1.26.1

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	$(GO) install github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.26.1

$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

$(BIN)/goreleaser: Makefile
	@mkdir -p $(@D)
	$(GO) install github.com/goreleaser/goreleaser@v1.16.2
