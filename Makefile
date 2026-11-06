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
COPYRIGHT_YEARS := 2023-2024
BUF_VERSION := v1.44.0
LICENSE_HEADER_VERSION := v1.36.0
GOLANGCI_LINT_VERSION := v2.14.0
GORELEASER_VERSION := v2.18.0
# Each tool is installed under a path that includes its version, so it's
# fetched again only when its version changes.
TOOLS := .tmp/tools
BUF := $(TOOLS)/buf/$(BUF_VERSION)/buf
LICENSE_HEADER := $(TOOLS)/license-header/$(LICENSE_HEADER_VERSION)/license-header
GOLANGCI_LINT := $(TOOLS)/golangci-lint/$(GOLANGCI_LINT_VERSION)/golangci-lint
GORELEASER := $(TOOLS)/goreleaser/$(GORELEASER_VERSION)/goreleaser
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
GOLANGCI_LINT_OS := $(shell echo $(UNAME_OS) | tr A-Z a-z)
GOLANGCI_LINT_ARCH := $(subst aarch64,arm64,$(subst x86_64,amd64,$(UNAME_ARCH)))
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
generate: $(BUF) $(LICENSE_HEADER) ## Regenerate code and licenses
	$(BUF) generate proto
	$(LICENSE_HEADER) \
		--license-type apache \
		--copyright-holder "The Connect Authors" \
		--year-range "$(COPYRIGHT_YEARS)"

.PHONY: lint
lint: $(GOLANGCI_LINT) $(BUF) ## Lint Go and protobuf
	test -z "$$($(BUF) format -d . | tee /dev/stderr)"
	$(GO) vet ./...
	$(GOLANGCI_LINT) run
	$(BUF) lint proto

.PHONY: lintfix
lintfix: $(GOLANGCI_LINT) $(BUF) ## Automatically fix some lint errors
	$(GOLANGCI_LINT) run --fix
	$(BUF) format -w proto

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
release: $(GORELEASER)
	$(GORELEASER) release

.PHONY: checkrelease
checkrelease: $(GORELEASER)
	# skips some validation and doesn't actually publish a release, just to test
	# that building a release works
	$(GORELEASER) release --clean --snapshot

.PHONY: runconformance
runconformance: runservertests runclienttests

.PHONY: runservertests
runservertests: $(BIN)/connectconformance $(BIN)/referenceserver $(BIN)/grpcserver
	$(BIN)/connectconformance -v --conf ./testing/reference-impls-config.yaml --mode server --trace \
		--known-failing @./testing/referenceserver-known-failing.txt \
		-- $(BIN)/referenceserver
	$(BIN)/connectconformance -v --conf ./testing/grpc-impls-config.yaml --mode server --trace \
		--known-failing @./testing/grpcserver-known-failing.txt \
		-- $(BIN)/grpcserver
	$(BIN)/connectconformance -v --conf ./testing/grpc-web-server-impl-config.yaml --mode server --trace \
		--known-failing @./testing/grpcserver-web-known-failing.txt \
		-- $(BIN)/grpcserver

.PHONY: runclienttests
runclienttests: $(BIN)/connectconformance $(BIN)/referenceclient $(BIN)/grpcclient grpcwebclient
	$(BIN)/connectconformance -v --conf ./testing/reference-impls-config.yaml --mode client --trace \
		--known-failing @./testing/referenceclient-known-failing.txt \
		-- $(BIN)/referenceclient
	$(BIN)/connectconformance -v --conf ./testing/grpc-impls-config.yaml --mode client --trace \
		--known-failing @./testing/grpcclient-known-failing.txt \
		-- $(BIN)/grpcclient
	$(BIN)/connectconformance -v --conf ./testing/grpc-web-client-impl-config.yaml --mode client --trace \
		--known-failing @./testing/grpcwebclient-known-failing.txt \
		-- ./testing/grpcwebclient/bin/grpcwebclient

.PHONY: grpcwebclient
grpcwebclient: generate testing/grpcwebclient/node_modules/.package-lock.json
	cd testing/grpcwebclient && npm run build

testing/grpcwebclient/node_modules/.package-lock.json: testing/grpcwebclient/package-lock.json
	cd testing/grpcwebclient && npm ci

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

$(BUF):
	@mkdir -p $(@D)
	curl -fsSL -o $@ https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-$(UNAME_OS)-$(UNAME_ARCH)
	chmod +x $@

# license-header has no release binaries.
$(LICENSE_HEADER):
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@$(LICENSE_HEADER_VERSION)

$(GOLANGCI_LINT):
	@mkdir -p $(@D)
	curl -fsSL https://github.com/golangci/golangci-lint/releases/download/$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION:v%=%)-$(GOLANGCI_LINT_OS)-$(GOLANGCI_LINT_ARCH).tar.gz \
		| tar -xzf - -C $(@D) --strip-components 1 golangci-lint-$(GOLANGCI_LINT_VERSION:v%=%)-$(GOLANGCI_LINT_OS)-$(GOLANGCI_LINT_ARCH)/golangci-lint

$(GORELEASER):
	@mkdir -p $(@D)
	curl -fsSL https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(UNAME_OS)_$(subst aarch64,arm64,$(UNAME_ARCH)).tar.gz \
		| tar -xzf - -C $(@D) goreleaser
