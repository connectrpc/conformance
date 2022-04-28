connect-crosstest
=================

[![Build](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml)

`connect-crosstest` runs a suite of cross-compatibility tests using every
combination of the following clients and servers:

**Servers:**
- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]

**Clients:**
- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]
- [grpc-web][grpc-web]
- [connect-web][connect-web]

The test suite is a superset of [gRPC][grpc-interop] and [grpc-web][grpc-web-interop] interop tests.
The test suite is also run nightly against the latest versions of [connect-go][connect-go] to
ensure that we are continuously testing for compatibility.

## Requirements and Running the Tests

You'll need Docker running on your machine and the test suite uses Docker Compose.
You can run the tests using `make test-docker-compose`.

## Support and Versioning

`connect-crosstest` works with:

* The most recent release of Go.
* [APIv2] of protocol buffers in Go (`google.golang.org/protobuf`).

Unlike Connect's Go implementation, `connect-crosstest` has no exported APIs
and makes no backward compatibility guarantees. We'd like to release it as an
interoperability testing toolkit eventually, but don't have a concrete timeline
in mind.

## Legal

Offered under the [Apache 2 license][license].

[APIv2]: https://blog.golang.org/protobuf-apiv2
[connect-go]: https://github.com/bufbuild/connect-go
[grpc-go]: https://github.com/grpc/grpc-go
[grpc-web]: https://github.com/grpc/grpc-web
[connect-web]: https://github.com/bufbuild/connect-web
[grpc-interop]: https://github.com/grpc/grpc/blob/master/doc/interop-test-descriptions.md
[grpc-web-interop]: https://github.com/grpc/grpc-web/blob/master/doc/interop-test-descriptions.md
[go-support-policy]: https://golang.org/doc/devel/release#policy
[license]: https://github.com/bufbuild/connect-crosstest/blob/main/LICENSE.txt
