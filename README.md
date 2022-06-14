# connect-crosstest

[![License](https://img.shields.io/github/license/bufbuild/connect-crosstest?color=blue)][license]
[![CI](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml/badge.svg?branch=main)][ci]

`connect-crosstest` runs a suite of cross-compatibility tests using every combination of the
following clients and servers:

### Servers

- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]

### Clients

- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]
- [grpc-web][grpc-web]
- connect-web (still in private alpha)

The test suite is run nightly against the latest commits of [connect-go][connect-go] and
connect-web to ensure that we are continuously testing for compatibility.

For more on Connect, see the [announcement blog post][blog], the documentation
on [connect.build][docs] (especially the [Getting Started] guide for Go), or
the [demo service][demo].

## Test Suite

The test suite is a superset of [gRPC][grpc-interop] and [grpc-web][grpc-web-interop] interop
tests. Clients and servers use the [gRPC interop Protobuf definitions][test.proto] and cover
a range of expected behaviours and functionality for gRPC and Connect.

| Test Case                                | `connect-go`, `grpc-go` | `connect-web`, `grpc-web` |
|------------------------------------------|-------------------------|---------------------------|
| `empty_unary`                            | ✓                       | ✓                         |
| `large_unary`                            | ✓                       | ✓                         |
| `client_streaming`                       | ✓                       |                           |
| `server_streaming`                       | ✓                       | ✓                         |
| `ping_pong`                              | ✓                       |                           |
| `empty_stream`                           | ✓                       | ✓                         |
| `fail_unary`                             | ✓                       | ✓                         |
| `fail_server_streaming`                  | ✓                       | ✓                         |
| `cancel_after_begin`                     | ✓                       |                           |
| `cancel_after_first_response`            | ✓                       |                           |
| `timeout_on_sleeping_server`             | ✓                       | ✓                         |
| `custom_metadata`                        | ✓                       | ✓                         |
| `duplicated_custom_metadata`             | ✓                       |                           |
| `status_code_and_message`                | ✓                       | ✓                         |
| `special_status_message`                 | ✓                       | ✓                         |
| `unimplemented_method`                   | ✓                       | ✓                         |
| `unimplemented_server_streaming_method`  | ✓                       | ✓                         |
| `unimplemented_service`                  | ✓                       | ✓                         |
| `unimplemented_server_streaming_service` | ✓                       | ✓                         |
| `unresolvable_host`                      | ✓                       |                           |

### Test Descriptions

#### empty_unary**

RPC: `EmptyCall`

Client calls `EmptyCall` with an `Empty` request and expects no errors and an empty response.

#### large_unary

RPC: `UnaryCall`

Client calls `UnaryCall` with a payload size of 250 KiB bytes and expects a response with a
payload size of 500 KiB and no errors.

#### client_streaming

RPC: `StreamingInputCall`

Client calls `StreamingInputCall` then sends 4 requests with a payload size of 250 KiB,
8 bytes, 1 KiB, and 32 KiB and expects the aggregated payload size to be 289800 bytes when
the client closes the stream and no errors.

#### server_streaming

RPC: `StreamingOutputCall`

Client calls `StreamingOutputCall` and receives exactly 4 times, expecting responses with
a payload size of 250 KiB, 8 bytes, 1 KiB, and 32 KiB, and no errors.

#### ping_pong

RPC: `FullDuplexCall`

Client calls `FullDuplexCall` exactly 4 times with a request with a payload of 250 KiB
and receives a response with a payload of 500 KiB, a request with a payload of 8 bytes
and receives a response with a payload of 16 bytes, a request with a payload of 1 KiB
and receives a response with a payload of 2 KiB, and a request with a payload of 32 KiB
and receives a response with a payload of 64 kiB. Client asserts that payload sizes
are in order and then closes the stream. No errors are expected.

#### empty_stream

RPC: `FullDuplexCall`/`StreamingOutputCall`

Client calls `FullDuplexCall` (web client calls `StreamingOutputCall`) and then closes. No
response or errors are expected.

#### fail_unary

RPC: `FailUnary`

Client calls `FailUnary` which always responds with an error with status `RESOURCE_EXHAUSTED`
and a non-ASCII message with error details.

#### fail_server_streaming

RPC: `FailStreamingOutputCall`

Client calls `FailStreamingOutputCall` which always responds with an error with status `RESOURCE_EXHAUSTED`
and a non-ASCII message with error details.

#### cancel_after_begin

RPC: `StreamingInputCall`

Client calls `StreamingInputCall`, cancels the context, then closes the stream, and expects
an error with the code `CANCELED`.

#### cancel_after_first_response

RPC: `StreamingInputCall`

Client calls `StreamingInputCall`, receives a response, then cancels the context, then closes
the stream, and expects an error with the code `CANCELED`.

#### timeout_on_sleeping_server

RPC: `FullDuplexCall`/`StreamingOutputCall`

Client calls `FullDuplexCall` (web client calls `StreamingOutputCall`) with a timeout, closes
the stream and expects to receive an error with status `DEADLINE_EXCEEDED`.

#### custom_metadata

RPC: `UnaryCall`, `StreamingOutputCall`, `FullDuplexCall`

Client calls `UnaryCall` with a request with a custom header and custom binary trailer attached
and expects the same metadata to be attached to the response. Client calls `StreamingOutputCall`
with a request with a custom header and custom binary trailer and expects the same metadata
to be attached to the response when stream is closed. Client calls `FullDuplexCall`
with a request with a custom header and custom binary trailer and expects the same metadata
to be attached to the response when stream is closed. The `web` flows only test the unary and 
server streaming RPC.

#### duplicated_custom_metadata

RPC: `UnaryCall`, `StreamingOutputCall`, `FullDuplexCall`

This is the same as the `custom_metadata` test but uses metadata values that have `,` separators
to test header and trailer behaviour.

#### status_code_and_message

RPC: `UnaryCall`, `FullDuplexCall`

Client calls `UnaryCall` with a request containing a `code` and `message` and expects an error
with the provided status `code` and `message`in response. Client calls `FullDuplexCall` with
a request containing a `code` and `message`, closes the stream, and expects to receive an
error with the provided status `code`and `message`. The `web` flows only test the unary RPC.

#### special_status_message

RPC: `UnaryCall`

Client calls `UnaryCall` with a request containing a `code` and `message` with whitespace
characters and Unicode and expects an error with the provided status `code` and `message`
in response.

#### unimplemented_method

RPC: N/A

Client calls `UnimplementedCall` with an empty request and expects an error with the status
`UNIMPLEMENTED`.

#### unimplemented_server_streaming_method

RPC: N/A

Client calls `UnimplementedStreamingOutputCall` with an empty request and expects an error with the status
`UNIMPLEMENTED`.

#### unimplemented_service

RPC: N/A

Client calls an unimplemented service and expects an error with the status `UNIMPLEMENTED`.

#### unimplemented_server_streaming_service

RPC: N/A

Client calls `UnimplementedStreamingOutputCall` to an unimplemented service with an empty request and expects
an error with the status `UNIMPLEMENTED`.

#### unresolvable_host

RPC: N/A

Client calls an unresolvable host and expects an error with the status `UNAVAILABLE`.

## Requirements and Running the Tests

### Github Actions

There is a Github Actions workflow configured for running the nightly crosstest. This can
also be used to trigger a manual run.

In the [Github Action crosstest workflow][github-action], you can trigger a manual run of
crosstest using the "Run workflow" button. This will also allow you to configure a branch
for `connect-go`, `protobuf-es`, or `connect-web` if you want to use the Github Actions to
test against a development branch of any of these packages.

### Locally

To run these locally tests, you'll need Docker. The test suite uses Docker Compose.
Please note, that if you are running the tests on MacOS, you'll need to [enable Docker
Compose V2][docker-compose-v2]. In short, Docker Desktop -> Preferences -> General -> Use Docker
Compose V2, then click Apply & Restart.

You can run the tests using `make dockercomposetest`.

> The following will no longer be needed once `connect-web` is public.

For our NPM tests, we need to pull private packages, `connect-web` and `protobuf-es` from
the NPM registry. This requires you to set a `NPM_TOKEN` env var in the environment you are
running the tests from.

## Support and Versioning

`connect-crosstest` works with:

* The most recent release of Go.

Unlike Connect's Go implementation, `connect-crosstest` has no exported APIs
and makes no backward compatibility guarantees. We'd like to release it as an
interoperability testing toolkit eventually, but don't have a concrete timeline
in mind.

[Getting Started]: https://connect.build/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[ci]: https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml
[connect-go]: https://github.com/bufbuild/connect-go
[demo]: https://github.com/bufbuild/connect-demo
[docker-compose-v2]: https://www.docker.com/blog/announcing-compose-v2-general-availability/#still-using-compose-v1
[docs]: https://connect.build
[github-action]: https://github.com/bufbuild/connect-crosstest/actions/workflows/crosstest.yaml
[go-support-policy]: https://golang.org/doc/devel/release#policy
[grpc-go]: https://github.com/grpc/grpc-go
[grpc-interop]: https://github.com/grpc/grpc/blob/master/doc/interop-test-descriptions.md
[grpc-web-interop]: https://github.com/grpc/grpc-web/blob/master/doc/interop-test-descriptions.md
[grpc-web]: https://github.com/grpc/grpc-web
[license]: https://github.com/bufbuild/connect-crosstest/blob/main/LICENSE
[test.proto]: https://github.com/bufbuild/connect-crosstest/blob/main/internal/proto/grpc/testing/test.proto
