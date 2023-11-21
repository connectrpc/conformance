# Connect Conformance

[![License](https://img.shields.io/github/license/connectrpc/conformance?color=blue)][license]
[![CI](https://github.com/connectrpc/conformance/actions/workflows/ci.yaml/badge.svg?branch=main)][ci]

A test suite for Connect cross-platform compatibility and conformance.

This conformance suite works with the most recent release of Go.

## Requirements and Running the Tests

### Running the tests

TBD


## Status: Alpha

This project is currently under development.

Because of its Alpha status, the conformance suite has no exported APIs
and makes no backward compatibility guarantees at this point. The goal is to
eventually publish a stable release but please be aware we may make changes
as we gather feedback from early adopters.


## Ecosystem

### Available Implementations

* [connect-go](https://github.com/connectrpc/connect-go):
  The Go implementation of Connect
* [connect-es](https://github.com/connectrpc/connect-es):
  The TypeScript implementation of Connect
* [connect-kotlin](https://github.com/connectrpc/connect-kotlin):
  The Kotlin implementation of Connect
* [connect-swift](https://github.com/connectrpc/connect-swift):
  The Swift implementation of Connect

### Examples

* [examples-go](https://github.com/connectrpc/examples-go):
  Example RPC service powering https://demo.connectrpc.com and built with connect-go
* [examples-es](https://github.com/connectrpc/examples-es):
  Examples for using Connect with various TypeScript web frameworks and tooling

### Ancillary

* [connect-playwright-es](https://github.com/connectrpc/connect-playwright-es):
  Playwright tests for your Connect application
* [connect-query-es](https://github.com/connectrpc/connect-query-es):
  TypeScript-first expansion pack for TanStack Query that gives you Protobuf superpowers


For more on Connect, see the [announcement blog post][blog], the documentation
on [connectrpc.com][docs] (especially the [Getting Started] guide for Go), or
the [demo service][demo].

## Legal

Offered under the [Apache 2 license][license].

[license]: https://github.com/connectrpc/conformance/blob/main/LICENSE
[Getting Started]: https://connectrpc.com/docs/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[ci]: https://github.com/connectrpc/conformance/actions/workflows/ci.yaml
[connect-go]: https://github.com/connectrpc/connect-go
[connect-es]: https://github.com/connectrpc/connect-es
[demo]: https://github.com/connectrpc/examples-go
[docs]: https://connectrpc.com
[license]: https://github.com/connectrpc/conformance/blob/main/LICENSE
[protobuf-es]: https://github.com/bufbuild/protobuf-es
