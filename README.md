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
reference server implementation of the Conformance Service written with [connect-go].

Likewise, servers under test will be verified by a reference client implementation of the Conformance
Service also written with connect-go.

To verify compatibility with other protocol implementations, the conformance tests also use reference client
and server implementations that use the [gRPC-Go module](https://github.com/grpc/grpc-go) and a reference
server implementation that uses the [gRPC-Web Go server](https://github.com/improbable-eng/grpc-web).

## Documentation

Detailed guides can be found in the "docs" sub-directory of this repo. If you first want to read a
shorter overview, see the next section [below](#testing-your-implementation). But when you really get
started with any of these tasks, you'll want to read one or more of these guides.

* [Configuring and Running Tests](./docs/configuring_and_running_tests.md)
* [Testing Server Implementations](./docs/testing_servers.md)
* [Testing Client Implementations](./docs/testing_clients.md)
* [Authoring New Test Cases](./docs/authoring_test_cases.md)

## How it works

The tests are data-driven: all test cases are defined in YAML files in this repo. These files
get embedded in the test runner so that the single self-contained executable contains all of
the test case data.

The test runner first processes your configuration and uses that to select which test cases
are relevant. Even if a test case is known to fail, it will still be executed to make sure it
is still failing (and report the fact if the test actually passes).

It then groups all of the test cases by the server configuration needed. So test cases that
will use TLS and the Connect protocol are in a different group from test cases that do _not_
use TLS and use the gRPC protocol.

It then begins running the tests.

```mermaid
sequenceDiagram
  actor user
    create participant test runner
    user ->> test runner: run conformance suite

    create participant client
    test runner -->> client: start process

    rect rgb(255,250,240)
    loop for each server config
        create participant server
        test runner -->> server: start process
        test runner ->> server: send config via stdin
        server ->> test runner: send result via stdout

        rect rgb(240,255,240)
        loop for each test case
            test runner ->>+ client: send RPC details via stdin
            client ->>+ server: invoke RPC, send request(s)
            server ->> server: process RPC
            server ->>- client: send response(s)
            client ->>- test runner: send RPC results via stdout
            test runner ->> test runner: assess RPC results
        end
        end

        destroy server
        test runner --x server: terminate
    end
    end

    destroy client
    test runner --x client: terminate

    destroy test runner
    test runner ->> user: report results
```

It first starts a client process (either a client under test, if in client mode, or a
reference client).

For each server configuration, it starts a server process (either a server under test, if
in server mode, or a reference server). It sends the server configuration details by writing
them to the process's _stdin_. When the server is listening on the network and ready to
accept RPCs, it sends the details to the test runner by writing to its _stdout_.

For each test case that applies to this server configuration, it adds details to the test
case data with the server's address, so the client will know how to reach it. It then
sends the test case data to the client by writing them to the process's _stdin_. The
client then invokes the RPC. It reports the RPC results to the test runner by writing
them to its _stdout_.

The test runner decides whether the test case was successful or not by comparing the
RPC results against expected results.

After all tests have been run and all child processes stopped, it reports the
results.

## Testing your implementation

### Setup

The conformance runner has the ability to test a client, a server, or both simultaneously. This means that if you are
validating both a server _and_ a client, you can use the conformance suite to run them against each other. Testing either a client
or server in isolation will use the corresponding reference implementation to verify conformance.

Below are the basic steps needed for setting up the suite to run against your implementation:

1. The first step is to access an SDK for the Conformance protos. These can be found on the Buf Schema Registry:
   https://buf.build/connectrpc/conformance. You can download SDKs for some languages
   [here](https://buf.build/connectrpc/conformance/sdks/main).

   If you are using a language that is not supported by the BSR's selection of SDKs, you can generate one yourself
   using the [`buf`](https://buf.build/docs/generate/tutorial) command-line tool.
   After creating a [`buf.gen.yaml`](https://buf.build/docs/configuration/v1/buf-gen-yaml) file, to configure the
   code generation, you'll then run `buf generate buf.build/connectrpc/conformance`.

2. Once you have an SDK, with generated code for the
   [Conformance Service](https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService)
   and related messages, you will need to implement either the service, the client, or both (depending on which you are testing).
   To do so, follow the instructions specified in the
   [`ConformanceService`](https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/service.proto) proto.

   For working examples, refer to the Go [reference client](./internal/app/referenceclient)
   and [reference server](./internal/app/referenceserver).

3. Your service-under-test or client-under-test needs to be a program that can easily be invoked from the command-line. This
   is how the conformance test runner will invoke it, too.

4. Next, visit this repository's [Releases](https://github.com/connectrpc/conformance/releases) page and download
   the conformance runner binary: `connectconformance`. You may want to add it to your `$PATH` to make it easier
   to run interactively from the command-line.

5. Finally, integrate the conformance tests into the continuous integration and testing process for your code,
   so every change you make can validate that the implementation remains conformant. You can see an example of
   how this can be done using `make` in the connect-kotlin repo,
   [here](https://github.com/connectrpc/connect-kotlin/blob/328110c00f791d06798aaa67f142d542bfcf1f27/Makefile#L46) and
   [here](https://github.com/connectrpc/connect-kotlin/blob/328110c00f791d06798aaa67f142d542bfcf1f27/Makefile#L111-L124).


### Running the tests

The commands for testing will depend on whether you are testing a client, a server, or both.
Specifying which implementation is done via the `mode` command line argument.

Once you have completed setup, the following commands will get you started:

#### Testing a client

```bash
connectconformance --mode client -- <path/to/your/executable/client>
```

#### Testing a server

```bash
connectconformance --mode server -- <path/to/your/executable/server>
```

#### Testing both a client _and_ server

To test your client against your server, specify a mode of `both`, with the client
path first, followed by `----`, then the path to your server:

```bash
connectconformance --mode both -- <path/to/your/executable/client> ---- <path/to/your/executable/server>
```

## Running the reference tests

To test this repo and the reference clients and servers, we can use the conformance suite itself.
To run the suite, using the reference clients against the reference servers, and see the process in
action, use the following command:

```bash
make runconformance
```

This will build the necessary binaries and run tests of the following implementations.

* Connect reference client and reference server
  * These implementations are written using [connect-go], but with numerous extensions that allow
    them to more closely examine the on-the-wire format of the RPC protocol to make stronger
    assertions about conformance.
  * They support all features that can be tested by the conformance tests, which includes all three
    protocols (Connect, gRPC, gRPC-Web), all HTTP versions (HTTP 1.1, HTTP/2, and even HTTP/3), and
    a variety of compression encodings ("gzip", "br", "zstd", "deflate" and "snappy").
* gRPC client and server
  * These implementations are written using [grpc-go](https://github.com/grpc/grpc-go).
  * They support the gRPC protocol and HTTP/2. They only support the "proto" codec for message
    encoding and the "gzip" compression encoding.
  * The server also supports the gRPC-Web protocol and HTTP 1.1 using the
    [improbable-eng/grpc-web](https://github.com/improbable-eng/grpc-web) Go implementation. This
    implementation is listed in the official gRPC-Web documentation as a server/proxy option
    [here](https://github.com/grpc/grpc-web#proxy-interoperability). (Note that the gRPC client
    does _not_ support gRPC-Web.)
  * These implementations are also used against clients-under-test and servers-under-test, to
    confirm interoperability with official gRPC implementations.

Both of the above clients are tested against both the Connect reference server and the gRPC server.
Both servers are tested against the Connect reference client and the gRPC client. And since the gRPC
client does not support gRPC-Web, the servers are also tested against the official gRPC-Web JS client.

## Status: Stable

This repo is stable and follows semantic versioning.

We will _not_ make breaking changes to the command-line interface, the YAML config
formats, or the Protobuf messages used by clients and servers under test in the
1.x series of releases.

Note, however, that we reserve the right to rename, remove, or re-organize
individual test cases, which may impact the "known failing" and "known flaky"
configurations for an implementation under test. We will document these changes
in the [release notes].

We also intend to occasionally add new test cases, and occasionally these
additions may also necessitate updates to the Protobuf schemas (such as new
request or response fields). The Protobuf changes will remain compatible, so
your programs will continue to compile, but actually passing new/updated test
cases may require updates to your program, to incorporate the new fields into
the behavior of the client or server under test. These kinds of changes will
also be documented in the releases notes.

New test cases in a release could also reveal previously undetected conformance
issues which may require fixes to the implementations you are testing. So while
we aim for backwards-compatibility and making it easy to upgrade to new releases
of the conformance suite, it is expected that some releases may incur some effort
to adopt. (See [the docs][upgrading] for more details.)

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
[release notes]: https://github.com/connectrpc/conformance/releases
[upgrading]: ./docs/configuring_and_running_tests.md#upgrading
