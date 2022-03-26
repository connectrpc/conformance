connect-crosstest
=================

[![Build](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml)

`connect-crosstest` runs a suite of cross-compatibility tests using every
combination of clients and servers from [Connect's Go
implementation][connect-go] and `grpc-go`. The test suite is a superset of the
[gRPC interop tests][interop].

As long as you have `bash` and `curl` installed, you can run the tests yourself
by cloning this repository and running `make`.

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
[Connect]: https://github.com/bufbuild/connect
[interop]: https://github.com/grpc/grpc/blob/master/doc/interop-test-descriptions.md
[go-support-policy]: https://golang.org/doc/devel/release#policy
[license]: https://github.com/bufbuild/connect-crosstest/blob/main/LICENSE.txt
