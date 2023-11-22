# Connect Conformance

[![License](https://img.shields.io/github/license/connectrpc/conformance?color=blue)][license]
[![CI](https://github.com/connectrpc/conformance/actions/workflows/ci.yaml/badge.svg?branch=main)][ci]

A test suite for [Connect](https://connectrpc.com) that verifies cross-platform conformance for
both clients and servers.

## Summary

The Connect conformance test suite is a series of tests that are run using a client and server to validate interoperability,
compatibility, and conformance across the Connect, gRPC, and gRPC-Web protocols. The test suite is meant to exercise
various scenarios with a client-server interaction to ensure the results are as expected across platforms.

Tests are divided into two types: client tests and server tests. Those which verify clients are run against a
reference server implementation of the Conformance Service written with [Connect Go](https://github.com/connectrpc/connect-go).

Likewise, servers under test will be verified by a reference client implementation of the Conformance
Service also written in Connect Go.

To verify compatibility with gRPC, at least initially, the conformance test also uses reference implementations of a gRPC-Go
server and client.

## Testing your implementation

### Setup
<!-- TODO - How in-depth do we want to get here with instructions? Should we specify how to generate files
from the BSR (which would also depend on their language). Or do we just specify a simple command for getting
the protos -->
The conformance runner has the ability to test a client, a server, or both simultaneously. This means that if you are
implementing both a server _and_ a client, you can run the conformance suite against each other. Testing either a client
or server in isolation will use the corresponding reference implementation to verify conformance.

Below are the basic steps needed for setting up the suite to run against your implementation:

1. The first step is to download the Conformance protos, which can be found on the Buf Schema Registry [here](TODO).
   From there, you will need to generate the code for the language you are testing.

2. Next, you will need to implement either the service, the client, or both (depending on which you are testing).

   To do so, you will need to implement the `ConformanceService` according to the instructions specified in the
   proto. For examples on how to implement, please refer to the Go [reference client](./internal/app/referenceclient)
   and [reference server](./internal/app/referenceserver).

3. Once implemented, your file should then be made executable in your target language. For example, if implementing
  `ConformanceService` in Go, you would build a binary for your implemented client or service.

4. Next, download the conformance runner and add it to your `$PATH`. The conformance test runner is published as a
   binary, released via Github artifacts. Visit the [Releases](https://github.com/connectrpc/conformance/releases) page to download.


### Running the tests

The commands for testing will depend on whether you are testing a client, a server, or both.
Specifying which implementation is done via the `mode` command line argument.

Once you have completed setup, the following commands will get you started:

#### Testing a client

```bash
connectconformance --mode client <path/to/your/executable/client>
```

#### Testing a server

```bash
connectconformance --mode server <path/to/your/executable/server>
```

#### Testing both a client _and_ server

To test your client against your server, specify a mode of `both`, with the client
path first, followed by `----`, then the path to your server:

```bash
connectconformance --mode both -- <path/to/your/executable/client> ---- <path/to/your/executable/server>
```

## Running the reference tests

To run the suite using the reference client against the reference server and see
the process in action, use the following command:

```bash
make runconformance
```

This will build the necessary binaries and run tests with the following setup:

* Connect reference client against a Connect reference server
* gRPC reference client against a Connect reference server
* Connect reference client against a gRPC reference server

<!-- What should we put as our status which properly conveys this is still a
work in progress? Alpha seems wrong bc our stuff is all at a v1 -->
## Status: Alpha

This project is currently under development.

Because of its Alpha status, the conformance suite has no exported APIs
and makes no backward compatibility guarantees at this point. The goal is to
eventually publish a stable release but please be aware we may make changes
as we gather feedback from early adopters.


## Ecosystem

### Implementations

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
