# Authoring Test Cases

Test cases for the conformance runner are configured in YAML files and are located [here][test-suite-dir]. Each file 
represents a single suite of tests. 

Authoring new test cases involves either a new file, if a new suite is warranted, or an additional test case in an 
existing file/suite. 

For the Protobuf definitions for the conformance runner, see the [connectrpc/conformance module][connectrpc-repo] in the BSR.

## Test Suites

A suite represents a series of tests that all test the same general functionality (cancellation, timeouts, etc.). Each 
test suite YAML file represents a [`TestSuite`][test-suite] message and the schema of this message defines the schema 
of the YAML file. The representation of a Protobuf message in YAML is the same as its [JSON format][json-docs].

When defining a test suite in a YAML file, the only values that are required are the suite name (which should be unique 
across all suites) and at least one test case.

In addition, each suite can be configured with various directives which will apply to all tests within that suite. These directives
can be used to constrain the tests run by the conformance runner or to signal various configurations to the runner that
may be needed to execute the tests. The runner will use these directives to expand the tests in the suite into multiple
permutations. This means that a single test defined in a suite will be run several times across various permutations.

The below directives are used to constrain tests within a suite:

* `mode` can be used to specify a suite as applying only to a specific mode (i.e. client or server). For example,
  if you are writing a suite to target only clients under test, you would specify `mode: TEST_MODE_CLIENT` and the
  tests in the suite will be run only when the `mode` specified to the runner on the command line is `client`. If not 
  specified in a suite, tests are run regardless of the command line `mode`.

* `relevantProtocols` is used to limit tests only to a specific protocol, such as **Connect**, **gRPC**, or **gRPC-web**. If not
  specified, tests are run for all protocols.

* `relevantHttpVersions` is used to limit tests to certain HTTP versions, such as **HTTP 1.1**, **HTTP/2**, or **HTTP/3**. If not
  specified, tests are run for all HTTP versions.

* `relevantCodecs` is used to limit tests to certain codec formats, such as **JSON** or **binary**. If not specified, tests are
   run for all codecs.

* `relevantCompressions` is used to limit tests to certain compression algorithms, such as **gzip**, **brotli**, or **snappy**. If not
  specified, tests are run for all compressions.

* `connectVersionMode` allows you to either require or ignore validation of the Connect version header or query param. This
  should be left unspecified if the suite is agnostic to this validation behavior.

The below `reliesOn` directives are used to signal to the test runner how the reference client or server should be
configured when running tests:

* `reliesOnTls` specifies that a suite relies on TLS. If `true`, the test cases will not be run against non-TLS server 
  configurations. Defaults to `false`.

* `reliesOnTlsClientCerts` specifies that the suite relies on the _client_ using TLS certificates to authenticate with 
  the server. Note that if this is set to `true`, `reliesOnTls` must also be `true`. Defaults to `false`.

* `reliesOnConnectGet` specifies that the suite relies on the Connect GET protocol. Defaults to `false`.

* `reliesOnMessageReceiveLimit` specifies that the suite relies on support for limiting the size of received messages.
  When `true`, the `mode` property must be set to indicate whether the client or server should support the limit. Defaults
  to `false`.

## Test Cases

Test cases are specified in the `testCases` property of the suite. Each test case starts with the `request` property 
which defines the request which will be sent to a client during the test run. Each `request` must specify the following
fields:

* `testName` - For naming conventions, see [below](#naming-conventions).
* `streamType` - One of `STREAM_TYPE_UNARY`, `STREAM_TYPE_CLIENT_STREAM`, `STREAM_TYPE_SERVER_STREAM`, 
  `STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`, or `STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`.

Once the above are specified, you can then define your request. For a full list of fields to specify in the request,
see the [`ClientCompatRequest`][client-compat-request] message in the Conformance Protobuf definitions.

The fields `service` and `method` are optional as a pair when writing test cases, meaning that they can both be omitted
or must be specified together. If they are omitted, the runner will auto-populate them as follows:

* `service` - `connectrpc.conformance.v1.ConformanceService`.
* `method` - Based on `streamType` according to the following table. 

  | Stream Type                            | Method         |
  | -------------------------------------- | -------------- |
  | `STREAM_TYPE_UNARY`                    | `Unary`        |
  | `STREAM_TYPE_CLIENT_STREAM`            | `ClientStream` |
  | `STREAM_TYPE_SERVER_STREAM`            | `ServerStream` |
  | `STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`  | `BidiStream`   |
  | `STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`  | `BidiStream`   |

 > [!IMPORTANT]  
 > The `ClientCompatRequest` message contains some fields that should _not_ be specified in test cases because they are 
 > automatically populated by the test runner. These fields are:
 > * `http_version`
 > * `protocol`
 > * `codec`
 > * `compression`
 > * `host`
 > * `port`
 > * `server_tls_cert`
 > * `client_tls_creds`
 > * `message_receive_limit`
 >
 > If a test is specific to one of the first four fields, it should instead be indicated in the directives for the test suite itself.

### Raw Payloads

There are two message types in the test case schema worth noting here - [`RawHTTPRequest`][raw-http-request] and 
[`RawHTTPResponse`][raw-http-response]. Both allow for the ability to define a round-trip outside the scope of the 
Connect framework. They are used for sending or receiving anomalous payloads during a test.

#### Raw Requests (For Server Tests)

The [`RawHTTPRequest`][raw-http-request] message can be set on the `request` property in a test case. Its purpose is to model a raw HTTP 
request. This can be used to craft custom requests with odd properties (including certain kinds of malformed requests) 
to test edge cases in servers. This value is only handled by the reference client and should only appear in files where 
`mode` is set to `TEST_MODE_SERVER`.

#### Raw Responses (For Client Tests)

The [`RawHTTPResponse`][raw-http-response] message is the analog to [`RawHTTPRequest`][raw-http-request]. It can be set in the response definition for a unary 
or streaming RPC type and its purpose is to model a raw HTTP response. This can be used to craft custom responses with 
odd properties (including returning aberrant HTTP codes or certain kinds of malformed responses) to test edge cases in 
clients. This value is only handled by the reference server and should only appear in files where `mode` is set to 
`TEST_MODE_CLIENT`.

## Naming Conventions

Test suites and their tests within follow a loose naming convention. 

### Test Files and Suites

Test files should be named according to the suite inside and the general functionality being tested. Additionally:

* If a suite applies only to a certain protocol, the file name should be prefixed with that protocol. If the suite applies
  to all protocols, this can be omitted.
* If a suite only contains client or server tests, the file name prefix should include `client` or `server`. If the suite is
  for both client and server, this can be omitted.

The file names should use `lower_snake_case` convention.

For example:
* `connect_with_get.yaml` contains tests that test GET requests for the Connect protocol only.
* `grpc_web_client.yaml` contains client tests for the gRPC-web protocol. 
* `client_message_size.yaml` file contains a suite of client tests only across all protocols.

The suite inside the file should either match the file name or be close to it. Reasons it might vary from the
file name is to provide a little more detail as to the test suite's purpose.

The test suite names should use `Title Case` convention

### Tests

Test names should be hyphen-delimited. If a suite contains tests of multiple stream types, the test name should be 
prefixed with the stream type and a backslash (`/`).

 > [!NOTE]  
 > In the case of Bidi tests with separate cases for half-duplex operation vs. full-duplex
 > operation, you should also add `full-duplex` or `half-duplex` to the test name. 

 For example:

* `unary/error`
* `server-stream/error-with-responses`
* `bidi-stream/full-duplex/stream-error-returns-success-http-code`

If a suite contains tests for just a single stream type, the stream type can be omitted from the test name.

These conventions allow for more granular control over running tests via the conformance runner, such as only running tests
for a specific protocol or only running the unary tests within a suite.

Aside from the `/` for separating elements on the name, test case names should use `kebab-case` convention.

## Expected responses

The expected response for a test is auto-generated based on the request details. The conformance runner will determine 
what the response should be according to the values specified in the test suite and individual test cases. 

You also have the ability to explicitly specify your own expected response directly in the test definition itself. However, 
this is typically only needed for exception test cases. If the expected response is mostly re-stating the response definition
that appears in the requests, you should rely on the auto-generation if possible. Otherwise, specifying an expected response 
can make the test YAML overly verbose and harder to read, write, and maintain. 

If the test induces behavior that prevents the server from sending or client from receiving the full response definition, it 
will be necessary to define the expected response explicitly. Timeouts, cancellations, and exceeding message size limits are 
good examples of this.

If you do need to specify an explicit response, simply define an `expectedResponse` block for your test case and this will
override the auto-generated expected response in the test runner. 

To see tests denoting an explicit response, search the [test suites][test-suite-dir] directory for the word `expectedResponse`.

## Running and Debugging New Tests

To test new test cases, you can use `make runconformance`, to run the reference implementations against the new test
cases. But, while iterating on the test case definition, it is often valuable to just run the test cases in the new
file. This can be done using the `--test-file` option to the test runner:
```shell
# Build the test runner and the reference client and server.
make .tmp/bin/connectconformance .tmp/bin/referenceclient .tmp/bin/referenceserver
# Run just the tests in the new file
.tmp/bin/connectconformance \
    --conf ./testing/reference-impls-config.yaml \
    --mode client \
    --test-file ./testsuites/new-test-suite.yaml \
    -v --trace \
    -- \
    .tmp/bin/referenceclient
.tmp/bin/connectconformance \
    --conf ./testing/reference-impls-config.yaml \
    --mode server \
    --test-file ./testsuites/new-test-suite.yaml \
    -v --trace \
    -- \
    .tmp/bin/referenceserver
```
If the new test suite is specific to client or server mode then you'd only need to run one of the above instead
of both.

You can debug the reference client and server by running the opposite mode above. For example, to debug the
reference server, run the first command above, which runs tests in client mode. In this mode, the reference
server is run in-process, so you can start the above `connectconformance` command with a debugger attached
and then step through the server code.

## Example 

Taking all of the above into account, here is an example test suite that verifies a server returns the specified headers
and trailers. The below test will only run when `mode` is `server` and will be limited to the Connect protocol using the
JSON format and identity compression. Also, it will require the server verify the Connect version header.

```yaml
name: Example Test Suite
# Constrain to only servers under test.
mode: TEST_MODE_SERVER
# Constrain these tests to only run over the Connect protocol.
relevantProtocols:
  - PROTOCOL_CONNECT
# Constrain these tests to only use 'identity' compression.
relevantCompressions:
  - COMPRESSION_IDENTITY
# Constrain these tests to only use the JSON codec.
relevantCodecs:
  - CODEC_JSON
# Require Connect version header verification.
connectVersionMode:  CONNECT_VERSION_MODE_REQUIRE
testCases:
  - request:
      testName: returns-headers-and-trailers
      service: connectrpc.conformance.v1.ConformanceService
      method: Unary
      streamType: STREAM_TYPE_UNARY
      requestMessages:
        - "@type": type.googleapis.com/connectrpc.conformance.v1.UnaryRequest
          responseDefinition:
            responseHeaders:
            - name: x-custom-header
              value: ["foo"]
            responseData: "dGVzdCByZXNwb25zZQ=="
            responseTrailers:
            - name: x-custom-trailer
              value: ["bing"]
```

[client-compat-request]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ClientCompatRequest
[connectrpc-repo]: https://buf.build/connectrpc/conformance
[json-docs]: https://protobuf.dev/programming-guides/proto3/#json
[raw-http-request]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.RawHTTPRequest
[raw-http-response]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.RawHTTPResponse
[test-suite]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.TestSuite
[test-suite-dir]: https://github.com/connectrpc/conformance/tree/main/internal/app/connectconformance/testsuites/data
