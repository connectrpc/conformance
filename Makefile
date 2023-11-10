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
	go test -vet=off -race -cover ./...

.PHONY: build
build: generate ## Build all packages
	go build ./...

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
	go vet ./...
	golangci-lint run
	buf lint proto

.PHONY: lintfix
lintfix: $(BIN)/golangci-lint $(BIN)/buf ## Automatically fix some lint errors
	golangci-lint run --fix
	buf format -w proto

.PHONY: install
install: ## Install all binaries
	go install ./...

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v

.PHONY: checkgenerate
checkgenerate:
	@# Used in CI to verify that `make generate` doesn't produce a diff.
	test -z "$$(git status --porcelain | tee /dev/stderr)"

.PHONY: runconformance
runconformance: runservertests runclienttests

.PHONY: runservertests
runservertests: $(BIN)/connectconformance $(BIN)/referenceserver
	$(BIN)/connectconformance --mode server $(BIN)/referenceserver

.PHONY: runclienttests
runclienttests: $(BIN)/connectconformance $(BIN)/referenceclient
	$(BIN)/connectconformance --mode client $(BIN)/referenceclient

$(BIN)/connectconformance: Makefile generate
	go build -o $(@) ./cmd/connectconformance/

$(BIN)/referenceclient: Makefile generate
	go build -o $(@) ./cmd/referenceclient/

$(BIN)/referenceserver: Makefile generate
	go build -o $(@) ./cmd/referenceserver/

$(BIN)/grpcclient: Makefile generate
	go build -o $(@) ./cmd/grpcclient/

$(BIN)/grpcserver: Makefile generate
	go build -o $(@) ./cmd/grpcserver/

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	go install github.com/bufbuild/buf/cmd/buf@v1.26.1

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	go install github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.26.1

$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
