# Testing Servers

The conformance suite provides the ability to run conformance tests against a server implementation. Testing servers 
involves the following steps:

1. Defining any configuration for what your server supports. For more information on how to do this, see the docs for [configuring and running tests](./configuring_and_running_tests.md#configuration-files).
2. Writing an executable file that can read [`ServerCompatRequest`][servercompatrequest] messages from `stdin` (and write to `stdout` -- see Step 4).
3. Starting your server according to the values in the request message.
4. Writing a [`ServerCompatResponse`][servercompatresponse] about the running server to `stdout`.
5. Implementing the [`ConformanceService`][conformanceservice] endpoints to handle requests from the reference client.

## Starting your server

When the conformance runner is executed for a server-under-test, the runner will analyze the configuration you've specified
and will use that information to build a [`ServerCompatRequest`][servercompatrequest]. This request is serialized to bytes and written to `stdin`.
It is then up to the executable file you created as part of Step 2 to read this message and start your server.

The messages written to `stdin` are size-delimited. This means that first you will need to read a fixed four-byte
preface, which returns a network-byte-order (i.e. big-endian) 32-bit integer. This integer represents the size of the
actual message. After this value is read, you should then read the number of bytes it specifies and then unmarshal those
bytes into a [`ServerCompatRequest`][servercompatrequest].

This message will contain all the details necessary for your implementation to start its server-under-test. Note that 
the request does not specify a port to listen on. Implementations should instead pick an available ephemeral port 
according to their OS and return that value in the response.

Fields in the request are:

* `protocol` which signals to the server that it must support at least this protocol. Note that it is fine to support others.
   For example if `PROTOCOL_CONNECT` is specified, the server _must_ support at least Connect, but _may_ also support
   gRPC or gRPC-web.
* `http_version` which signals to the server the minimum HTTP version to support. As with `protocol`, it is fine to support 
   other versions. 
* `use_tls` which specifies whether your server must use TLS when handling HTTP requests.

   The test-runner provides a certificate and private key that the server may use. This certificate is self-signed
   and is only valid for local traffic using "localhost" hostname or 127.0.0.1 IP address. If this sort of certificate
   is inadequate, your server must provide its own certificate. If your server provides its own certificate, make sure
   the `host` field of the [`ServerCompatResponse`][servercompatresponse] matches subject name or an alternative name
   in the certificate.

   When `use_tls` is true, a certificate _must_ be returned in the `pem_cert` field of the
   [`ServerCompatResponse`][servercompatresponse]. If the server will use the one provided by the test runner,
   it must indicate so by setting this field to the certificate provided in this request. Clients will be
   configured to trust whatever certificate the server sends in the response.

   If `use_tls` is false, the server must not use TLS and instead use a plaintext/unencrypted socket.
* `server_creds` which specifies a TLS certificate and key that the server may use. This is only present when
   `use_tls` is true. See above for more details.
* `client_tls_cert` which represents a PEM-encoded certificate that, if provided, clients will use to authenticate themselves. 
   If specified, servers should require that clients provide certificates and they should ensure the presented certificate is valid.
   Note that this will only be present if `use_tls` is `true`.
* `message_receive_limit` which specifies the maximum size in bytes for a message. If this value is non-zero, servers should reject
   any message from a client that is larger than the size indicated.

Using the values in the request, you can then start your server implementation. Once started, your implementation should 
build a [`ServerCompatResponse`][servercompatresponse] message. This will provide the conformance runner with details about your running server.

Once built, you should then write the response message to `stdout` using the same size-delimited algorithm described above. First, 
write a network-encoded 32-bit integer indicating the size of the [`ServerCompatResponse`][servercompatresponse] message. Then, serialize the
response to bytes and write that to `stdout`.

Fields in the response are:

* `host` which should be set with the host where your server is running. This should usually be `127.0.0.1`, unless your 
   program actually starts a remote server to which the client should connect.
* `port` which should be set with the port number where your server is listening.
* `pem_cert` which should contain the TLS certificate the server will use, in PEM format, if `use_tls` was set to true
   in the request. Clients will verify this certificate when connecting via TLS. If `use_tls` was set to `false`, this
   should always be empty.

## Implementing the ConformanceService

When verifying a server-under-test, the conformance runner will use a reference client 
implementation written in Connect-Go and will use this client to issue requests to the server. The reference client 
will read the server's responses and return them to the conformance runner. So, all you need to do from a server 
implementation standpoint is handle the requests accordingly.

The [`ConformanceService`][conformanceservice] defines a series of endpoints that are meant to exercise all types of RPCs. The details
in the requests will specify various attributes that the server should handle, such as determining response headers and 
trailers to return, any errors to throw, and any data to respond with. All request types contain a response definition 
which is used to instruct the server how to respond to the request. This response definition can also be unset entirely
and servers should respond in a fashion specific to their RPC type. 

In addition to any specified response data, servers must also echo back the received request data in their responses. 
This will be done via one of two ways -- either by setting the information into a `ConformancePayload` message in their
response or by setting the information into the details of an `Error` message.

### ConformancePayload

The [`ConformancePayload`][conformancepayload] message is a field on most all response types which contains the following fields:

* `data` which should be set with any response data specified in the response definition.
* `request_info` which is a nested message of type [`RequestInfo`][requestinfo] structured as:
  * `request_headers` which represents any observed request headers.
  * `timeout_ms` which indicates any timeout included in the request.
  * `requests` which is a `repeated` field of [`google.protobuf.Any`][any] types. This is used to echo back all the requests 
     received. For unary and server-streaming requests, this should always contain a single request. For client-streaming
     and half-duplex bidi-streaming, this should contain all client requests in the order received and should be present
     in each response. For full-duplex bidi-streaming, this should contain all requests in the order they were received
     since the last sent response.
  * `connect_get_info` which is only applicable for GET operations such as [`IdempotentUnary`](#idempotentunary). It should contain any
     observed query parameters in the request URL.

 > [!NOTE]  
 > The response type for the [`Unimplemented`](#unimplemented) endpoint does not contain a conformance payload as
 > implementations are not meant to echo back any information.

### Error

The [`Error`][error] message is used to return errors back to the client. It is structured as follows:

* `code` which represents the error code.
* `message` which is an optional field indicating the error message.
* `details` which is a `repeated` field of [`google.protobuf.Any`][any] types. This is used to attach arbitrary messages and
   in the conformance suite is used to echo back request information in circumstances where a conformance payload was
   unable to be returned

## Endpoints

In all, there are six total endpoints in the conformance service definition. Below is a brief description of each with helpful 
pseudocode.

### Unary

The `Unary` endpoint is a unary operation that accepts a single request of type `UnaryRequest` and returns a single 
response of type `UnaryResponse`. 

The `UnaryRequest` contains a `response_definition` field that contains a 
`oneof` which specifies whether the server should return valid response data or return an error. If an error is specified,
servers should set request information into the error details. 

Servers should also allow this response definition to be unset. In which case, they should set no response headers or trailers and return 
no response data. The returned conformance payload should only contain the observed request information.

**Pseudocode**

```text
read the request

capture any request headers and the actual request body as request info

if response definition specifies an error {
  build an error with the specified error and set the request info into the error details

  set any response headers or trailers indicated in the response definition

} else  {
  build a conformance payload with the request info 

  if a response definition exists {
    set response data into the conformance payload

    set any response headers or trailers indicated in the response definition
  }
}

sleep for any specified response delay

return the response with conformance payload or raise the error, depending on which was specified
```

For the full documentation on implementing the `Unary` endpoint, click [here][unary].


### IdempotentUnary

The `IdempotentUnary` endpoint is also a unary operation. It accepts a request of type  `IdempotentUnaryRequest` and returns
a single response of type `IdempotentUnaryResponse`. It should be handled in mostly the same way as `Unary`.

The only major difference is that this endpoint should be invoked via an HTTP `GET`. As a result, there is no request body
so the endpoint should read any query parameters and set them accordingly in the `connect_get_info` field of the 
returned `ConformancePayload`. This RPC is the only one that sets the `connect_get_info` field.

Note that this endpoint is only applicable for Connect implementations that support `GET` requests. All others do not
need to implement it.

For the full documentation on implementing the `IdempotentUnary` endpoint, click [here][idempotentunary].

### Unimplemented

The `Unimplemented` endpoint is also a unary operation, but contrary to the above unary endpoints, this endpoint should 
not be implemented at all. The server framework should instead return an `unimplemented` error. It is not necessary to 
echo back any request information or conformance payload in the error details.

For the full documentation on handling the `Unimplemented` endpoint, click [here][unimplemented].

### ClientStream

The `ClientStream` endpoint is a client-streaming operation. It accepts one-to-many requests of type `ClientStreamRequest`
and returns a single response of type `ClientStreamResponse`. Since a client-streaming operation returns a single response, 
its process is similar to `Unary`.

With client-streaming, the response definition will only be specified in the first request on the stream and should be
ignored in all subsequent requests. As with `Unary`, if an error is specified, servers should set request information into
the error details. 

Servers should also allow this response definition to be unset. In which case, they should set no response headers or 
trailers and return no response data. The returned conformance payload should only contain the observed request information.

**Pseudocode**

```text
while requests are being sent {
   read a request from the stream

   capture the request body

   if this is the first message being received {
      capture the response definition
   }
}

when requests are complete {
  if response definition specifies an error {
    build an error with the specified error and set the request info into the error details

    set any response headers or trailers indicated in the response definition

  } else  {
    build a conformance payload with the request info 

    if a response definition exists {
      set response data into the conformance payload

      set any response headers or trailers indicated in the response definition
    }
  }
}

sleep for any specified response delay

return the response with conformance payload or raise the error, depending on which was specified
```

For the full documentation on handling the `ClientStream` endpoint, click [here][clientstream].

### ServerStream

The `ServerStream` endpoint is a server-streaming operation. It accepts a single request of type `ServerStreamRequest` and
returns one-to-many response of type `ServerStreamResponse`. When echoing request information back, the `ServerStream`
implementation should only set this information in the first response sent.

Servers should immediately send response headers on the stream before sleeping
for any specified response delay and/or sending the first message so that
clients can be unblocked reading response headers.
  
If a response definition is not specified OR is specified, but response data
is empty, the server should skip sending anything on the stream. When there
are no responses to send, servers should throw an error if one is provided
and return without error if one is not. Stream headers and trailers should
still be set on the stream if provided regardless of whether a response is
sent or an error is thrown.

**Pseudocode**

```text
read the request

capture any request headers, the response definition and the actual request body as request info

if a response definition was specified {
  set any response headers or trailers on the response stream

  immediately send the headers/trailers on the stream so that they can be read by the client

  loop over any response data specified {
    build a conformance payload

    set the response data into the conformance payload

    if this is the first response being sent {
      set the request info into the conformance payload
    }
 
    sleep for any specified response delay

    send the response
  }

  if an error was specified in the response definition {
    build an error with the specified error

    if no responses have been sent yet {
      set the request info into the error details
    }
    raise the error
  }
}
```

For the full documentation on handling the `ServerStream` endpoint, click [here][serverstream].

### BidiStream

The `BidiStream` endpoint is a bidirectional-streaming operation. It accepts 
one-to-many requests of type `BidiStreamRequest` and returns one-to-many 
responses of type `BidiStreamResponse`. 

Similar to `ServerStream`, servers should immediately send response headers on 
the stream before sleeping for any specified response delay and/or sending the 
first message so that clients can be unblocked reading response headers.
  
If a response definition is not specified OR is specified, but response data
is empty, the server should skip sending anything on the stream. When there
are no responses to send, servers should throw an error if one is provided
and return without error if one is not. Stream headers and trailers should
still be set on the stream if provided regardless of whether a response is
sent or an error is thrown.

The `BidiStreamRequest` type specifies whether the operation should be full duplex
or half duplex via the `full_duplex` field.

If the `full_duplex` field is true:

* the handler should read one request and then send back one response, and
  then alternate, reading another request and then sending back another response, etc.
  
* if the server receives a request and has no responses to send, it
  should throw the error specified in the request.
  
* the service should echo back all request properties in the first response
  including the last received request. Subsequent responses should only
  echo back the last received request.
  
* if the `response_delay_ms` duration is specified, the server should wait the given
  duration after reading the request before sending the corresponding
  response.
  
If the `full_duplex` field is false:

* the handler should read all requests until the client is done sending.
  Once all requests are read, the server should then send back any responses
  specified in the response definition.
  
* the server should echo back all request properties, including all request
  messages in the order they were received, in the first response. Subsequent
  responses should only include the message data in the data field.
  
* if the `response_delay_ms` duration is specified, the server should wait that
  long in between sending each response message.

**Pseudocode for handling full duplex streams**

```text
while requests are being sent {
   read a request from the stream

   capture the request body

   if this is the first message being received {
      capture the response definition
      capture full_duplex value (`true`)

      if a response definition was specified {
        set any response headers or trailers on the response stream
         
        immediately send the headers/trailers on the stream so that they can be read by the client
      }
   }
   
   if response data was specified {
     build a conformance payload

     set the response data into the conformance payload

     if this is the first response being sent {
       set the request info into the conformance payload
     }
   
     sleep for any specified response delay

     send a response
  }
}

if an error was specified in the response definition {
  build an error with the specified error

  if no responses have been sent yet {
    set the request info into the error details
  }
  raise the error
}
```

**Pseudocode for handling half duplex streams**

```text
while requests are being sent {
   read a request from the stream

   capture the request body

   if this is the first message being received {
      capture the response definition
      capture full_duplex value (`false`)
   }
}

if response data was specified {
  immediately send the headers/trailers on the stream so that they can be read by the client

  loop over any response data specified {
    build a conformance payload

    set the response data into the conformance payload

    if this is the first response being sent {
        set the request info into the conformance payload
    }
 
    sleep for any specified response delay

    send the response
  }
}

if an error was specified in the response definition {
  build an error with the specified error

  if no responses have been sent yet {
    set the request info into the error details
  }
  raise the error
}
```

For the full documentation on handling the `BidiStream` endpoint, click [here][bidistream].

## Examples

For examples, check out the following:

* [Connect-ES conformance tests][connect-es-conformance] - This shows the entire process described above in TypeScript/JavaScript.
* [Connect-Go reference implementation][server-reference-impl] - For an example of implementing a server in Go, take a look at the reference server implementation used as part of the conformance runner.

[connect-es-conformance]: https://github.com/connectrpc/connect-es/tree/main/packages/connect-conformance 
[server-reference-impl]: https://github.com/connectrpc/conformance/blob/main/internal/app/referenceserver/impl.go
[conformanceservice]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService
[conformancepayload]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformancePayload
[servercompatrequest]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ServerCompatRequest
[servercompatresponse]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ServerCompatResponse
[requestinfo]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformancePayload.RequestInfo
[error]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.Error
[any]: https://buf.build/protocolbuffers/wellknowntypes/docs/main:google.protobuf#google.protobuf.Any
[unary]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.Unary
[idempotentunary]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.IdempotentUnary
[unimplemented]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.Unimplemented
[clientstream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.ClientStream
[serverstream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.ServerStream
[bidistream]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConformanceService.BidiStream
