# Conformance

[![License](https://img.shields.io/github/license/connectrpc/conformance?color=blue)][license]
[![CI](https://github.com/connectrpc/conformance/actions/workflows/ci.yaml/badge.svg?branch=main)][ci]
[![conformance-go](https://github.com/connectrpc/conformance/actions/workflows/conformance-go.yaml/badge.svg?branch=main)][github-action-go]
[![conformance-web](https://github.com/connectrpc/conformance/actions/workflows/conformance-web.yaml/badge.svg?branch=main)][github-action-web]
[![conformance-cc](https://github.com/connectrpc/conformance/actions/workflows/conformance-cc.yaml/badge.svg?branch=main)][github-action-cc]

This repo runs a suite of cross-compatibility tests using every combination of the
following clients and servers:

### Servers

- [connect-go][connect-go] (Connect protocol, gRPC protocol, and gRPC-web protocol)
- [connect-es][connect-es] (Connect for Node.js serving Connect protocol, gRPC protocol, and gRPC-web protocol)
- [grpc-go][grpc-go]

### Clients

- [connect-go][connect-go] (Connect protocol, gRPC protocol and gRPC-web protocol)
- [connect-es][connect-es] (Connect for Web using Connect protocol and gRPC-web protocol)
- [grpc-go][grpc-go]
- [grpc-web][gRPC-web]

The test suite is run daily against the latest commits of [connect-go][connect-go], [connect-es][connect-es]
and [protobuf-es][protobuf-es] to ensure that we are continuously testing for compatibility.

For more on Connect, see the [announcement blog post][blog], the documentation
on [connectrpc.com][docs] (especially the [Getting Started] guide for Go), or
the [demo service][demo].

## Test Suite

The test suite is a superset of [gRPC][grpc-interop] and [gRPC-web][grpc-web-interop] interop
tests. Clients and servers use the [gRPC interop Protobuf definitions][test.proto] and cover
a range of expected behaviors and functionality for gRPC and Connect.

| Test Case                                | `connect-go`, `grpc-go` | `connect-es`, `grpc-web` |
|------------------------------------------|-------------------------|---------------------------|
| `empty_unary`                            | ✓                       | ✓                         |
| `cacheable_unary`                        | ✓                       | ✓                         |
| `large_unary`                            | ✓                       | ✓                         |
| `client_streaming`                       | ✓                       |                           |
| `server_streaming`                       | ✓                       | ✓                         |
| `ping_pong`                              | ✓                       |                           |
| `empty_stream`                           | ✓                       | ✓                         |
| `fail_unary`                             | ✓                       | ✓                         |
| `fail_server_streaming`                  | ✓                       | ✓                         |
| `fail_server_streaming_after_response`   |                         | ✓                         |
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

#### empty_unary

RPC: `EmptyCall`

Client calls `EmptyCall` with an `Empty` request and expects no errors and an empty response.

#### large_unary

RPC: `UnaryCall`

Client calls `UnaryCall` with a payload size of 250 KiB bytes and expects a response with a
payload size of 500 KiB and no errors.

#### cacheable_unary

RPC: `CacheableUnaryCall`

Client calls `CacheableUnaryCall` with a small payload. Expected to be called via `GET`
request when called with the Connect protocol.

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

#### fail_server_streaming_after_response

RPC: `FailStreamingOutputCall`

Client calls `FailStreamingOutputCall`, and asks for four response messages. The server
responds with the messages, the status `RESOURCE_EXHAUSTED` and a non-ASCII message, and
error details. The client verifies that four response messages and the error status with
code, message, and details was received.

#### cancel_after_begin

RPC: `StreamingInputCall`

Client calls `StreamingInputCall`, cancels the context, then closes the stream, and expects
an error with the code `CANCELED`.

#### cancel_after_first_response

RPC: `FullDuplexCall`

Client calls `FullDuplexCall`, receives a response, then cancels the context, then closes
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

There are Github Actions workflows for [go][github-action-go] and [web][github-action-web] configured for running the daily conformance suite against
the latest commits of [connect-go][connect-go], [connect-es][connect-es] and [protobuf-es][protobuf-es].

In addition, there is a Github Action workflow for [C++][github-action-cc] which runs a gRPC C++ client against the
conformance suite.

### Running the tests

To run these tests locally, you'll need Docker. The test suite uses Docker Compose.
Please note, that if you are running the tests on MacOS, you'll need to [enable Docker
Compose V2][docker-compose-v2]. In short, Docker Desktop -> Preferences -> General -> Use Docker
Compose V2, then click Apply & Restart.

You can run the tests using `make dockercomposetest`.

## Support and Versioning

This conformance suite works with the most recent release of Go.

Unlike Connect's Go implementation, the conformance suite has no exported APIs
and makes no backward compatibility guarantees. We'd like to release it as an
interoperability testing toolkit eventually, but don't have a concrete timeline
in mind.

## Legal

Offered under the [Apache 2 license][license].

[Getting Started]: https://connectrpc.com/docs/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[ci]: https://github.com/connectrpc/conformance/actions/workflows/ci.yaml
[connect-go]: https://github.com/connectrpc/connect-go
[connect-es]: https://github.com/connectrpc/connect-es
[demo]: https://github.com/connectrpc/examples-go
[docker-compose-v2]: https://www.docker.com/blog/announcing-compose-v2-general-availability/#still-using-compose-v1
[docs]: https://connectrpc.com
[github-action-go]: https://github.com/connectrpc/conformance/actions/workflows/conformance-go.yaml
[github-action-web]: https://github.com/connectrpc/conformance/actions/workflows/conformance-web.yaml
[github-action-cc]: https://github.com/connectrpc/conformance/actions/workflows/conformance-cc.yaml
[go-support-policy]: https://golang.org/doc/devel/release#policy
[grpc-go]: https://github.com/grpc/grpc-go
[grpc-interop]: https://github.com/grpc/grpc/blob/master/doc/interop-test-descriptions.md
[grpc-web-interop]: https://github.com/grpc/grpc-web/blob/master/doc/interop-test-descriptions.md
[grpc-web]: https://github.com/grpc/grpc-web
[license]: https://github.com/connectrpc/conformance/blob/main/LICENSE
[protobuf-es]: https://github.com/bufbuild/protobuf-es
[test.proto]: https://github.com/connectrpc/conformance/blob/main/proto/connectrpc/conformance/v1/test.proto
