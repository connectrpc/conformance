MAKEGO := make/go
MAKEGO_REMOTE := https://github.com/bufbuild/makego.git
PROJECT := connect-crosstest
GO_MODULE := github.com/bufbuild/connect-crosstest
GO_MOD_VERSION := 1.18
GO_GET_PKGS := $(GO_GET_PKGS) github.com/bufbuild/connect@main
GO_ALL_REPO_PKGS := ./internal/...
# Remove when https://github.com/golangci/golangci-lint/pull/2438 is fixed (Golang 1.18 support)
SKIP_GOLANGCI_LINT := 1
LICENSE_HEADER_LICENSE_TYPE := apache
LICENSE_HEADER_COPYRIGHT_HOLDER := Buf Technologies, Inc.
LICENSE_HEADER_YEAR_RANGE := 2020-2022
LICENSE_HEADER_IGNORES := \/testdata

include make/go/bootstrap.mk
include make/go/go.mk
include make/go/buf.mk
include make/go/license_header.mk
include make/go/dep_protoc_gen_go.mk
include make/go/dep_protoc_gen_go_grpc.mk

.PHONY: installprotoc-gen-go-connect
installprotoc-gen-go-connect:
	go install github.com/bufbuild/connect/cmd/protoc-gen-go-connect

bufgeneratedeps:: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) installprotoc-gen-go-connect

.PHONY: bufgeneratecleango
bufgeneratecleango:
	rm -rf internal/gen/proto

bufgenerateclean:: bufgeneratecleango

.PHONY: bufgeneratego
bufgeneratego:
	buf generate

bufgeneratesteps:: bufgeneratego
