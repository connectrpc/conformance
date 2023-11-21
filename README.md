# Connect Conformance

[![License](https://img.shields.io/github/license/connectrpc/conformance?color=blue)][license]
[![CI](https://github.com/connectrpc/conformance/actions/workflows/ci.yaml/badge.svg?branch=main)][ci]

A test suite for Connect cross-platform compatibility and conformance.

This conformance suite works with the most recent release of Go.

## Status: Alpha

This project is currently under development.

Because of its Alpha status, the conformance suite has no exported APIs
and makes no backward compatibility guarantees at this point. The goal is to
eventually publish a stable release but please be aware we may make changes
as we gather feedback from early adopters.


## Legal

Offered under the [Apache 2 license][license].

[license]: https://github.com/connectrpc/conformance/blob/main/LICENSE

## Support and Versioning




# Ecosystem



For more on Connect, see the [announcement blog post][blog], the documentation
on [connectrpc.com][docs] (especially the [Getting Started] guide for Go), or
the [demo service][demo].

## Test Suite

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

Client sends exactly 4 requests over a `FullDuplexCall` stream:

- A request with a payload of 250 KiB, expecting a response with a payload of 500 KiB
- A request with a payload of 8 bytes, expecting a response with a payload of 16 bytes
- A request with a payload of 1 KiB, expecting a response with a payload of 2 KiB
- A request with a payload of 32 KiB, expecting a response with a payload of 64 KiB

Client asserts that payload sizes
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

### Running the tests

TBD




[Getting Started]: https://connectrpc.com/docs/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[ci]: https://github.com/connectrpc/conformance/actions/workflows/ci.yaml
[connect-go]: https://github.com/connectrpc/connect-go
[connect-es]: https://github.com/connectrpc/connect-es
[demo]: https://github.com/connectrpc/examples-go
[docs]: https://connectrpc.com
[license]: https://github.com/connectrpc/conformance/blob/main/LICENSE
[protobuf-es]: https://github.com/bufbuild/protobuf-es
