# Testing Clients

The conformance suite provides the ability to run conformance tests against a client
implementation. Testing clients involves the following steps:

1. Defining any configuration for what your client supports. For more information on
   how to do this, see the docs for
   [configuring and running tests](./configuring_and_running_tests.md#configuration-files).
2. Writing an executable file that can read [`ClientCompatRequest`][clientcompatrequest]
   messages from `stdin` and write results to `stdout`.
3. Each message read from `stdin` describes an RPC to issue to a server. The client
   must issue the RPC and then construct a  [`ClientCompatResponse`][clientcompatresponse]
   message that describes the RPC result and write that to `stdout`.
4. When EOF on `stdin` is reached, the program should wait for any in-progress RPCs to
   complete and then exit. If the program receives a `SIGTERM` signal, it should exit
   immediately.

## Interacting with the test runner

When the conformance runner is executed for a client-under-test, the runner will analyze
the configuration you've specified and will use that information to build a number of
[`ClientCompatRequest`][clientcompatrequest] messages. These requests are serialized to
bytes and written to the client program's `stdin`. It is then up to the executable file
you created as part of Step 2 to read these messages and send RPCs as described in the
messages.

The messages written to `stdin` are size-delimited. This means that first you will need to
read a fixed four-byte preface, which returns a network-byte-order (i.e. big-endian) 32-bit
integer. This integer represents the size of the actual message. After this value is read,
you should then read the number of bytes it specifies and then unmarshal those bytes into a
[`ClientCompatRequest`][clientcompatrequest].

This message will contain all the details necessary for your implementation to create an
RPC client connection to a reference server and to issue an RPC. After the RPC has completed,
your implementation should build a [`ClientCompatResponse`][clientcompatresponse] message.
This will provide the conformance runner with details about the operation. This message is
then written to `stdout` using the same size-delimited algorithm described above. First,
write a network-encoded 32-bit integer indicating the size of the message. Then, serialize the
response to bytes and write that to `stdout`.

The test runner will usually send multiple such requests, so your program should use a loop
to keep reading these requests until it reaches EOF. The simplest programs will simply read
one message, execute the RPC, write its result, and then repeat. But it is acceptable for
the results to be written to `stdout` out of order. So the client could use parallelism (to
speed up the test run), and write the results to `stdout` as they are available. Care must
be taken so that concurrent writes to `stdout` do not interleave and corrupt the output.

The first field in the request provides the full name of the test case: `test_name`.
There are two other kinds of fields in the request:

1. **Creating an RPC client**: These first fields describe how to connect to the RPC server
   and what kinds of options will be used to construct an RPC client/stub.
   * `http_version`: If TLS is in use, it is acceptable to ignore this field and have your
     implementation use protocol negotiation during the TLS handshake. The reference server
     will only support the given version. You will only receive requests with HTTP versions
     that your implementation supports, as described by your config YAML file.
   * `protocol`: This indicates which protocol to use to invoke the RPC. This will indicate
     Connect, gRPC, or gRPC-Web.
   * `codec`: This indicates which codec to use for message serialization. This is also
     sometimes called "sub-format" or "message encoding".
   * `compression`: Which compression format to use. The "Identity" option means to not use
     compression.
   * `host` and `port`: The host name or IP address and the IP port on which the server
     is listening for requests. This is used to establish a network connection for issuing
     requests.
   * `server_tls_cert`: If TLS is in use, this field will be non-empty and contain the
     PEM-encoded bytes of the server's certificate. The client should be configured to
     treat this certificate as a trusted certificate authority (also called a trusted
     root or trusted certificate issuer). If this field is empty, TLS is _not_ used.
     When `http_version` indicates HTTP/2, and this field is empty, the client must use
     H2C (HTTP/2 Cleartext).
   * `client_tls_creds`: This will only be present when `server_tls_cert` is non-empty.
     This instructs the client to use the given client certificate. The server will reject
     connections that do not use this certificate to authenticate themselves. The
     credentials are provided in the form of PEM-encoded bytes for the private key and for
     certificate (which includes the public key and chain of trust).
   * `message_receive_limit`: This option indicates the maximum size of response messages
     that the client can receive. If the client does not support such an option, it should
     be correctly configured in the config YAML, and this field can then be ignored.
2. **Invoking the RPC**: The next group of fields describe how to actually invoke the RPC,
   indicating the method to invoke, the metadata and request data to send, and optionally
   when to cancel the RPC.
   * `service`: The fully-qualified name of the RPC service to send. Client-under-test
     programs only need to handle a value of "connectrpc.conformance.v1.ConformanceService".
   * `method`: The name of the method to invoke.
   * `stream_type`: This indicates the kind of operation that will be used. The stream type
     can also be inferred from the `method`, except for the "BidiStream" method, where this
     stream type must be consulted to decide whether to do a half-duplex or full-duplex
     bidirectional stream operation.
   * `use_get_http_method`: If true, the client should use the "GET" HTTP method when sending
     the request. This is often not something that can be configured explicitly in the client,
     in which case it should suffice to simply enable the use of "GET". The actual RPC
     described is side-effect-free, so a correct implementation should automatically use "GET"
     when correctly configured. This field can be ignored when the `protocol` is not Connect.
     If the client implementation does not support "GET", it should be correctly configured
     in the config YAML, and this field can be ignored.
   * `request_headers`: Request header metadata to send with the RPC.
   * `request_messages`: Request message data to send as part of the RPC request.
   * `timeout_ms`: A timeout for the RPC, which is used to set a deadline.
   * `request_delay_ms`: An arbitrary delay to wait before sending each message in a client
     or bidirectional stream. (Can be ignored for unary and server stream RPCs.) This is used
     to insert transmission delays and can be useful to testing timeouts and other kinds of
     interactions.
   * `cancel`: If present, describes when the client should cancel the RPC.

In the response message, the program must echo back the test name from the request, in the
`test_name` field of `ClientCompatResponse`. This allows the test runner to correlate results
with the original request, which is mainly necessary if your client uses parallelism and may
write results in a different order than it received the requests.

There are two other fields in the response message:

1. `error`: This field is only used if something prevents your client from invoking the RPC.
   Note that this does **not** include RPC errors. If the RPC was invoked, but an error
   occurs after that, then the error should be described in the other field. This is for
   other kinds of failures, such as inability to even create an RPC stub.
2. `response`: This field describes the results of issuing the RPC. This is a
   [`ClientResponseResult`][clientresponseresult] message, which has the following fields:
   * `response_headers`: Response header metadata received from the server.
   * `payloads`: These describe the response message data. Each RPC response message that the
     client must handle has a field of this type. This is a repeated field since server and
     bidirectional streams can have more than one response message. This field should describe
     all response messages received.
   * `error`: If an RPC error occurs, it should be described in this field. Note that unary
     and client-stream RPCs that encounter an RPC error will have _zero_ payloads. But for
     server and bidirectional streams, an error could occur after one or more response messages
     have been received, in which case the client program must populate both this field and
     the `payloads` field.
   * `response_trailers`: Response trailer metadata received from the server. In some client APIs,
     the trailers (and even headers) may not be directly accessible when an error occurs with a
     unary or client-stream operation. In these cases, the error API should provide access to
     metadata, which should be used to populate this field.
   * `num_unsent_requests`: For server and bidirectional streams, it is possible for the operation
     to end _before_ all requests are sent. If a send operation fails, the client program should
     record that, and any other remaining request messages described in the `ClientCompatRequest`,
     as an unsent request.

## Implementing the Client

When verifying a client-under-test, the conformance runner will use a reference server
implementation written in Connect-Go. The actual service used, [`ConformanceService`][conformanceservice],
defines a total of six endpoints that are meant to exercise all types of RPCs. The
client should only need to handle these six endpoints. If a client program sees an
unrecognized service or method it should send back an error result.

### Request messages and `Any`

The request format uses [`google.protobuf.Any`][any] messages to describe the requests.

If the client implementation has a Protobuf runtime with special library support for this
well-known type, then it is typically very straight-forward to work with this type. You
will need to "unmarshal" or "unpack" the message inside the `Any`, and then make sure that
is the right request type for the method being invoked. For example, when invoking the
`Unary` RPC method, the request type is `UnaryRequest`, so the code would need to verify
this before trying to send that message as a request. Some library support requires
providing some form of message registry when unpacking an `Any`. This provides information
about known message types; if the `Any` contains any other message type, it can't be
unpacked. The client program only needs to include the six request message types for the
six methods of `ConformanceService` in such a registry.

It is okay if a client implementation does _not_ have special library support for this
well-known type. If not, then the client program must verify that the type name matches
the expected request type name, and then unmarshal the value (bytes in the Protobuf
binary format) into the right request type. To verify the type name, the code must first
strip away the type URL prefix by finding the _last_ occurrence of the slash `/` character
and discarding it and everything before it. What remains is a fully-qualified type name.
This can be compared to the fully-qualified request type name. For example, the `UnaryRequest`
type has a fully-qualified name of `connectrpc.conformance.v1.UnaryRequest`. If the client
implementation's Protobuf runtime doesn't have a programmatic way to access this full name
(which is the case for some "lite" runtimes), it is acceptable to hardcode them. (There
are only six such constants that will be needed.)

### Error handling

Handling RPC errors, when an invocation fails, can depend greatly on the API provided
by the client implementation:

* Some APIs return values that indicate failure; some throw exceptions.
* Some APIs provide direct access to headers and trailers, some may only provide this
  with successful responses and/or for response streams (i.e. server and bidirectional
  stream RPCs). When they are not accessible, they may instead be provided as an
  attribute of the error. If the error has a "metadata" attribute, treat them as
  trailers (and if headers are inaccessible in such cases, treat them as empty).
* Some APIs may return a sentinel error for "send" operations, requiring you to
  follow up with a "receive" operation to observe the actual cause of error.
* Some APIs that use an "observer" pattern may provide an alternative way to
  communicate the error to an observer after an attempt to send a message fails.

In general, all invocations should abort when a failure occurs. The error or exception
should be converted to an [`Error`][error] message. If an unexpected error or exception
occurs, and no RPC code or details can be extracted, construct an error with a code
of "unknown" and a message string of whatever message the error or exception contains.

When the operation is aborted, the client program must still extract headers and trailers
if they are available and include them in the result. If an RPC error occurs, then any
error or exception returned/thrown by an attempt to access headers or trailers can be
ignored (and the headers and/or trailers treated as empty).

### Cancellation

The `ClientCompatRequest` can contain instructions for the client program to cancel the
RPC before it has completed. Ideally, when the RPC is canceled based on these instructions,
the rest of the invocation logic should proceed as if it had _not_ been canceled. This way,
the client program can exercise the client implementation's cancellation handling, and how
it impacts subsequent operations for the call. This allows the conformance suite to verify
that asynchronous cancellations are handled correctly by the implementation and result in
proper notification of the cancellation to the code that is consuming the RPC results.

### Stream types

The `stream_type` field of the `ClientCompatRequest` is used to interpret the other
request fields and informs how the client program must interact with the RPC server.

Three of the six methods of `ConformanceService` are unary RPCs. There is one each
for client and server streaming RPCs. And the sixth method, which is a bidirectional
stream, is used for both half-duplex and full-duplex stream types. If the language
used for the client implementations supports generic, that is an ideal way to implement
the way the client interacts with unary RPCs, where generic type parameters can be used
to define the request and response types as well as the function (or method on a stub)
that invokes the RPC.

There are five different stream types, each of which is described below, including
pseudocode for how to dispatch the RPC and interact with the RPC server.

#### Unary

The [`Unary`][unary], [`Unimplemented`][unimplemented], and [`IdempotentUnary`][idempotentunary]
methods all use the "unary" stream type. This style of RPC accepts exactly one request message
and results in _either_ one response message and no RPC error _or_ zero response messages and an error.

For this stream type, the `request_messages` field of `ClientCompatRequest` will contain exactly
one item.

**Pseudocode**

```text
invoke the method using the given request headers,
   the one request message, and a timeout if provided

* if we should cancel after close send {
   delay the indicated number of milliseconds   
   cancel the RPC (but do not return)
}
receive the response

if the operation fails {
   abort, returning a result that describes the error and any
      available headers and trailers
}
† extract the payload field from the response
construct a result using the payload and any available headers
   and trailers
```

_*_ Note: some client APIs will provide a blocking operation for unary RPCs,
    which doesn't return until the RPC response is received. For these cases,
    you must arrange for the RPC to be canceled asynchronously after the indicated
    number of milliseconds, and then invoke the blocking operation.

_†_ Note: an empty response message is possible. In these cases, the client should
    use an empty `ConformancePayload` message value as the payload.

#### Client stream

The [`ClientStream`][clientstream] method uses the "client stream" stream type. This style of RPC
accepts zero or more request messages and then results in _either_ one response message and no RPC
error _or_ zero response messages and an error.

**Pseudocode**

```text
invoke the method using the given request headers and a timeout if provided

if the operation fails {
   abort, returning a result that describes the error and any
      available headers and trailers
}

for each request message {
   delay for the indicated number of milliseconds
   send the request message
   if an error occurs {
      record the number of unsent requests (including this one)
         and include in the result
      abort, returning a result that describes the error and any
         available headers and trailers
   }
}

if we should cancel before close send {
   cancel the RPC (but do not return)
}

* close send (aka "close request")
if we should cancel after close send {
   delay the indicated number of milliseconds   
   cancel the RPC (but do not return)
}
receive the response

if an error occurs {
   abort, returning a result that describes the error and any
      available headers and trailers
}

† extract the payload field from the response
construct a result using the payload and any available headers
   and trailers
```

_*_ Note: some client APIs will provide a single, atomic "close and receive"
    operation for client stream RPCs. In that case, since you can't independently
    close and then receive, you also can't cancel in between. Instead, you must
    arrange for the RPC to be canceled asynchronously after the indicated number
    of milliseconds, and then "close and receive".

_†_ Note: an empty response message is possible. In these cases, the client should
    use an empty `ConformancePayload` message value as the payload.

#### Server stream

The [`ServerStream`][serverstream] method uses the "server stream" stream type. This style of RPC
accepts exactly one request message and then results in zero or more response messages and an
optional RPC error.

For this stream type, the `request_messages` field of `ClientCompatRequest` will contain exactly
one item.

**Pseudocode**

```text
* invoke the method using the given request headers,
   the one request message, and a timeout if provided

if the operation fails {
   abort, returning a result that describes the error and any
      available headers and trailers
}

if we should cancel after close send {
   delay the indicated number of milliseconds   
   cancel the RPC (but do not return)
}

use array to accumulate payload values
for each response message {
   † extract the payload field from the response and
      record it in array of payload values
      
   if we should cancel after N response messages and this is the Nth {
      cancel the RPC (but do not return)
   }
    
   ‡ if an error occurs {
      abort, returning a result that describes payload
         values accumulated so far, the error, and any
         available headers and trailers
   }
}

construct a result using the accumulated payloads and any available
   headers and trailers
```

_*_ Note: some client APIs may provide an "invoke" operation for server stream RPCs
    that does not accept the request message. In these cases, it is typically up to the
    caller to explicitly send the request message immediately after the stream is created.

_†_ Note: an empty response message is possible. In these cases, the client should
    use an empty `ConformancePayload` message value as the payload.

_‡_ Note: some client APIs may return an error or throw an exception if an attempt is made
    to receive a response message but there are none remaining. Such APIs will typically
    use a sentinel error or exception type that simply means "end-of-stream". In these
    cases, such a sentinel should cause the client to break out of this loop and _not_ treat
    this as an error case.

#### Half-Duplex Bidi Stream

The [`BidiStream`][bidistream] method can be used for this stream type. This style of RPC
allows the client to send zero or more request messages and then the server can respond
with zero or more response messages and an optional error. In general, a bidirectional
stream method simply allows either side (client and server) to send an arbitrary number
of messages. But "half-duplex" is a style of bidirectional stream where the client sends
all of its data first, before the server replies with its first response message. This
style of RPC can technically be supported over HTTP 1.1, which does not otherwise allow
full-duplex communication during the life of a single HTTP call.

**Pseudocode**

```text
invoke the method using the given request headers and a timeout if provided

if the operation fails {
   abort, returning a result that describes the error and any
      available headers and trailers
}

for each request message {
   delay for the indicated number of milliseconds
   send the request message
   if an error occurs {
      record the number of unsent requests (including this one)
         and include in the result
      abort, returning a result that describes the error and any
         available headers and trailers
   }
}

if we should cancel before close send {
   cancel the RPC (but do not return)
}

close send (aka "close request")
if we should cancel after close send {
   delay the indicated number of milliseconds   
   cancel the RPC (but do not return)
}

use array to accumulate payload values
for each response message {
   * extract the payload field from the response and
      record it in array of payload values
      
   if we should cancel after N response messages and this is the Nth {
      cancel the RPC (but do not return)
   }
    
   † if an error occurs {
      abort, returning a result that describes payload
         values accumulated so far, the error, and any
         available headers and trailers
   }
}

construct a result using the accumulated payloads and any available
   headers and trailers
```
_*_ Note: an empty response message is possible. In these cases, the client should
    use an empty `ConformancePayload` message value as the payload.

_†_ Note: some client APIs may return an error or throw an exception if an attempt is made
    to receive a response message but there are none remaining. Such APIs will typically
    use a sentinel error or exception type that simply means "end-of-stream". In these
    cases, such a sentinel should cause the client to break out of this loop and _not_ treat
    this as an error case.

It is no coincidence that the above pseudo-code looks like merging the first half of the
client stream logic with the latter half of the server stream logic. A client program may
choose to consolidate the shared logic into helper functions to make the client program
more concise and less repetitive (especially if the language is dynamically typed or supports
generics).

#### Full-Duplex Bidi Stream

Like half-duplex streams, the [`BidiStream`][bidistream] method can be used for this stream
type. The difference is that full-duplex allows the request stream and response streams to
overlap -- so the server may send one or more response messages before the client has finished
sending all the request messages.

**Pseudocode**

```text
invoke the method using the given request headers and a timeout if provided

if the operation fails {
   abort, returning a result that describes the error and any
      available headers and trailers
}

use array to accumulate payload values
for each request message {
   delay for the indicated number of milliseconds
   send the request message
   if an error occurs {
      record the number of unsent requests (including this one)
         and include in the result
      abort, returning a result that describes the error and any
         available headers and trailers
   }
   
   receive a response message
   * if an error occurs {
      record the number of unsent requests and include in the result
      abort, returning a result that describes the error and any
         available headers and trailers
   }

   † extract the payload field from the response and
      record it in array of payload values

   if we should cancel after N response messages and this is the Nth {
      cancel the RPC (but do not return)
   }
}

if we should cancel before close send {
   cancel the RPC (but do not return)
}

close send (aka "close request")
if we should cancel after close send {
   delay the indicated number of milliseconds   
   cancel the RPC (but do not return)
}

for each remaining response message {
   † extract the payload field from the response and
      record it in array of payload values
      
   if we should cancel after N response messages and this is the Nth {
      cancel the RPC (but do not return)
   }
    
   * if an error occurs {
      abort, returning a result that describes payload
         values accumulated so far, the error, and any
         available headers and trailers
   }
}

construct a result using the accumulated payloads and any available
   headers and trailers
```

_*_ Note: some client APIs may return an error or throw an exception if an attempt is made
    to receive a response message but there are none remaining. Such APIs will typically
    use a sentinel error or exception type that simply means "end-of-stream". In these
    cases, such a sentinel should cause the client to break out of this loop and _not_ treat
    this as an error case.

_†_ Note: an empty response message is possible. In these cases, the client should
    use an empty `ConformancePayload` message value as the payload.

## Examples

For examples, check out the following:

* [Connect-ES conformance tests][connect-es-conformance] - This shows the entire process described above in TypeScript/JavaScript.
* [Connect-Kotlin conformance client][connect-kotlin-conformance] - This shows the process described above in Kotlin.

[connect-es-conformance]: https://github.com/connectrpc/connect-es/tree/main/packages/connect-conformance
[connect-kotlin-conformance]: https://github.com/connectrpc/connect-kotlin/tree/main/conformance/client/src/main/kotlin/com/connectrpc/conformance/client
[conformanceservice]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService
[conformancepayload]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformancePayload
[clientcompatrequest]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ClientCompatRequest
[clientcompatresponse]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ClientCompatResponse
[clientresponseresult]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ClientResponseResult
[any]: https://buf.build/protocolbuffers/wellknowntypes/docs/main:google.protobuf#google.protobuf.Any
[error]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.Error
[unary]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.Unary
[idempotentunary]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.IdempotentUnary
[unimplemented]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.Unimplemented
[clientstream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.ClientStream
[serverstream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.ServerStream
[bidistream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.BidiStream