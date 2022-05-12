connect-crosstest
=================

[![Build](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/connect-crosstest/actions/workflows/ci.yaml)

`connect-crosstest` runs a suite of cross-compatibility tests using every combination of the
following clients and servers:

**Servers:**
- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]

**Clients:**
- Connect, using [Connect's Go implementation][connect-go]
- gRPC, using [grpc-go][grpc-go]
- [grpc-web][grpc-web]
- [connect-web][connect-web]

The test suite is run nightly against the latest commits of [connect-go][connect-go] and
[connect-web][connect-web] to ensure that we are continuously testing for compatibility.

## Test Suite 

The test suite is a superset of [gRPC][grpc-interop] and [grpc-web][grpc-web-interop] interop
tests. Clients and servers use the [gRPC interop Protobuf definitions][test.proto] and cover
a range of expected behaviours and functionality for gRPC and Connect.

| Test Case | `connect-go`, `grpc-go` | `connect-web`, `grpc-web` |
| --- | --- | --- |
| `empty_unary` | :ballot_box_with_check: | :ballot_box_with_check: |
| `large_unary` | :ballot_box_with_check: | :ballot_box_with_check: |
| `client_streaming` | :ballot_box_with_check: | |
| `server_streaming` | :ballot_box_with_check: | :ballot_box_with_check: |
| `ping_pong` | :ballot_box_with_check: | |
| `empty_stream` | :ballot_box_with_check: | :ballot_box_with_check: |
| `fail_unary` | :ballot_box_with_check: | :ballot_box_with_check: |
| `cancel_after_begin` | :ballot_box_with_check: | |
| `cancel_after_first_response` | :ballot_box_with_check: | |
| `timeout_on_sleeping_server` | :ballot_box_with_check: | |
| `custom_metadata` | :ballot_box_with_check: | :ballot_box_with_check: |
| `duplicated_custom_metadata` | :ballot_box_with_check: | |
| `status_code_and_message` | :ballot_box_with_check: | :ballot_box_with_check: |
| `special_status_message` | :ballot_box_with_check: | :ballot_box_with_check: |
| `unimplemented_method` | :ballot_box_with_check: |
| `unimplemented_service` | :ballot_box_with_check: | :ballot_box_with_check: |

### Test Descriptions

**empty_unary**:

RPC: `EmptyCall`

Client calls `EmptyCall` with an `Empty` request and expects no errors and an empty response.

**large_unary**:

RPC: `UnaryCall`

Client calls `UnaryCall` with a payload size of 271828 bytes and expects a response with a
payload size of 314159 bytes and no errors.

**client_streaming**:

RPC: `StreamingInputCall`

Client calls `StreamingInputCall` then sends 4 requests with a payload size of 271828 bytes,
8 bytes, 1828 bytes, and 45904 bytes and expects the aggregated payload size to be 74922 when
the client closes the stream and no errors.

**server_streaming**:

RPC: `StreamingOutputCall`

Client calls `StreamingOutputCall` and receives exactly 4 times, expecting responses with
a payload size of 271828 bytes, 8 bytes, 1828 bytes, and 45904 bytes, and no errors.

**ping_pong**:

RPC: `FullDuplexCall`

Client calls `FullDuplexCall` exactly 4 times with a request with a payload of 31415 bytes
and receives a response with a payload of 27182 bytes, a request with a payload of 9 bytes
and receives a response with a payload of 8 bytes, a request with a payload of 2653 bytes
and receives a response with a payload of 1828 bytes, and a request with a payload of 58979
bytes and receives a response with a payload of 45904 bytes. Client asserts that payload sizes
are in order and then closes the stream. No errors are expected.

**empty_stream**:

RPC: `FullDuplexCall`/`StreamingOutputCall`

Client calls `FullDuplexCall` (web client calls `StreamingOutputCall`) and then closes. No
response or errors are expected.

**fail_unary**:

RPC: `FailUnary`

Client calls `FailUnary` which always responds with an error with status `RESOURCE_EXHAUSTED`
and a non-ASCII message.

**cancel_after_begin**:

RPC: `StreamingInputCall`

Client calls `StreamingInputCall`, cancels the context, then closes the stream, and expects
an error with the code `CANCELED`.

**cancel_after_first_response**:

RPC: `StreamingInputCall`

Client calls `StreamingInputCall`, receives a response, then cancels the context, then closes
the stream, and expects an error with the code `CANCELED`.

**timeout_on_sleeping_server**:

RPC: `FullDuplexCall`/`StreamingOutputCall`

Client calls `FullDuplexCall` (web client calls `StreamingOutputCall`) with a context that
cancels the deadline, closes the stream and expects to receive an error with status `DEADLINE_EXCEEDED`.

**custom_metadata**:

RPC: `UnaryCall`, `FullDuplexCall`

Client calls `UnaryCall` with a request with a custom header and custom binary trailer attached
and expects the same metadata to be attached to the response. Client calls `FullDuplexCall`
with a request with a custom header and custom binary trailer and expects the same metadata
to be attached to the response when stream is closed. The `web` flows only test the unary RPC.

**duplicated_custom_metadata**:

RPC: `UnaryCall`, `FullDuplexCall`

This is the same as the `custom_metadata` test but uses metadata values that have `,` separators
to test header and trailer behaviour.

**status_code_and_message**:

RPC: `UnaryCall`, `FullDuplexCall`

Client calls `UnaryCall` with a request containing a `code` and `message` and expects an error
with the provided status `code` and `message`in response. Client calls `FullDuplexCall` with
a request containing a `code` and `message`, closes the stream, and expects to receive an
error with the provided status `code`and `message`. The `web` flows only test the unary RPC.

**special_status_message**:

RPC: `UnaryCall`

Client calls `UnaryCall` with a request containing a `code` and `message` with whitespace
characters and Unicode and expects an error with the provided status `code` and `message`
in response.

**unimplemented_method**:

RPC: N/A

Client calls `UnimplementedCall` with an empty request and expects an error with the status
`UNIMPLEMENTED`.

**unimplemented_service**:

RPC: N/A

Client calls an unimplemented service and expects an error with the status `UNIMPLEMENTED`.


## Requirements and Running the Tests

To run these tests, you'll need Docker. The test suite uses Docker Compose.
You can run the tests using `make test-docker-compose`.

To run the tests against the latest commits of `connect-go` and `connect-web` (instead of the
latest release), set the env var `TEST_LATEST_COMMIT=1`.

```
$ TEST_LATEST_COMMIT=1 make test-docker-compose
```


> The following will no longer be needed once `connect-web` is public.

For our NPM tests, we need to pull private packages, `connect-web` and `protobuf-es` from
the NPM registry. This requires you to set a `NPM_TOKEN` env var in the environment you are
running the tests from.

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
[test.proto]: https://github.com/bufbuild/connect-crosstest/blob/main/internal/proto/grpc/testing/test.proto
