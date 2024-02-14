# Authoring test cases

Test cases for the conformance runner are configured in YAML files and are located [here][testsuites]. Each file 
represents a single suite of tests. 

Authoring new test cases involves either a new file, if a new suite is warranted, or an additional test case in an 
existing file/suite. 

For the Protobuf definitions for the conformance runner, see the [connectrpc repository][connectrpc-repo] in the BSR.

## Test suites

A suite represents a series of tests that all test the same general functionality (cancellation, timeouts, etc.). When
defining a test suite in a YAML file, the only values that are required are the suite name (which should be unique across all suites)
and at least one test case.

Each suite can be configured with various directives which will apply to all tests within that suite. These directives
can be used to constrain the tests run by the conformance runner or to signal various configurations to the runner that
may be needed to execute the tests. The runner will use these directives to expand the tests in the suite into multiple
permutations. This means that a single test defined in a suite will be run several times across various permutations.

A rundown of the available directives:

The below directives are used to constrain tests within a suite:

* `mode` can be used to specify a suite as applying only to a specific mode (i.e. client or server). For example,
  if you are writing a suite to target only clients under test, you would specify `mode: TEST_MODE_CLIENT` and the
  tests in the suite will be run only when the `mode` specified to the runner on the command line is `client`. If not 
  specified, tests are run regardless of the `mode` set on the command line.

* `relevantProtocols` is used to limit tests only to a specific protocol, such as Connect, gRPC, or gRPC-web. If not
  specified, tests are run for all protocols.

* `relevantHttpVersions` is used to limit tests to certain HTTP versions, such as HTTP 1.1, HTTP/2, or HTTP/3.

* `relevantCodecs` is used to limit tests to certain codec formats, such as JSON or binary.

* `relevantCompressions` is used to limit tests to certain compression algorithms, such as **gzip**, **brotli**, or **snappy**.

* `connectVersionMode` allows you to either require or ignore validation of the Connect version header or query param. This
  should be left unspecified if the suite is agnostic to this validation behavior.

The below `reliesOn` directives are used to signal to the test runner how the reference client or server should be
configured when running tests:

* `reliesOnTls` specifies that a suite relies on TLS. 

* `reliesOnTlsClientCerts` specifies that the suite relies on the _client_ using TLS certificates to authenticate with 
  the server. Note that if this is set to `true`, `reliesOnTls` must also be `true`.

* `reliesOnConnectGet` specifies that the suite relies on the Connect GET protocol.

* `reliesOnMessageReceiveLimit` specifies that the suite relies on support for limiting the size of received messages.
  When `true`, the `mode` property must be set to indicate whether the client or server should support the limit.  

## Test cases

Test cases are specified in the `testCases` property of the suite. Each test case starts with the `request` property 
which defines the request which will be sent to a client during the test run. Each `request` must specify the following
fields:

* `testName` - For naming conventions, see [below](#naming-conventions).
* `service` - This will most likely be the fully-qualified path to the `ConformanceService`.
* `method` - This is a string specifying the method on `service` that will be called.
* `streamType` - One of `STREAM_TYPE_UNARY`, `STREAM_TYPE_CLIENT_STREAM`, `STREAM_TYPE_SERVER_STREAM`, 
  `STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`, or `STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`.

Once the above are specified, you can then define your request. For a full list of fields to specify in the request,
see the [`ClientCompatRequest`][client-compat-proto] message in the Conformance Protobuf definitions.

 > [!IMPORTANT]  
 > The `ClientCompatRequest` message contains some fields that should _not_ be specified in test cases.
 > These fields include:
 > * Fields 1 through 8 in the message definition. These fields are automatically populated by the test runner.
 >   If a test is specific to one of these values, it should instead be indicated in the directives for the test suite
 >   itself.
 > * Field 20 (`raw_request`). This field is only used by the reference client for sending anomalous requests.

## Naming conventions

Test suites and their tests within follow a loose naming convention. 

### Test files (Suites)

Test files should be named according to the suite inside and the general functionality being tested. In addition:

* If a suite applies only to a certain protocol, the file name should be prefixed with that protocol. If the suite applies
  to all protocols, this can be omitted.
* If a suite only contains client or server tests, the file name prefix should include `client` or `server`. If the suite is
  for both client and server, this can be omitted.

For example:
* `connect_with_get.yaml` contains tests that test GET requests for the Connect protocol only.
* `grpc_web_client.yaml` contains client tests for the gRPC-web protocol. 
* `client_message_size.yaml` file contains a suite of client tests only across all protocols. 

### Tests

Test names should be hyphen-delimited. If a suite contains tests of multiple stream types, the test name should be 
prefixed with the stream type and a backslash (`/`).

 > [!NOTE]  
 > In the case of Bidi tests with separate cases for half-duplex operation vs. full-duplex
 > operation, you should also add `full_duplex` or `half_duplex` to the test name. 

 For example:

* `unary/error`
* `server-stream/error-with-responses`
* `bidi-stream/full-duplex/stream-error-returns-success-http-code`

If a suite contains tests for just a single stream type, the stream type can be omitted from the test name.

These conventions allow for more granular control over running tests via the conformance runner, such as only running tests
for a specific protocol or only running the unary tests within a suite.

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

To tests denoting an explicit response, search the [test suites](testsuites) directory for the word `expectedResponse`.

## Example 

Taking all of the above into account, here is an example test suite:

```yaml
name: TLS Client Certs
# This just does the basics with a client-cert, instead of running every test case with them.
# TODO - Add unary and other stream type tests here also
testCases:
  - request:
      testName: client-stream
      service: connectrpc.conformance.v1.ConformanceService
      method: ClientStream
      streamType: STREAM_TYPE_CLIENT_STREAM
      requestHeaders:
        - name: X-Conformance-Test
          value: ["Value1","Value2"]
      requestMessages:
        - "@type": type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest
          responseDefinition:
            responseHeaders:
              - name: x-custom-header
                value: ["foo"]
            responseData: "dGVzdCByZXNwb25zZQ=="
            responseTrailers:
              - name: x-custom-trailer
                value: ["bing"]
        - "@type": type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest
          requestData: "dGVzdCByZXNwb25zZQ=="
```

[testsuites]: https://github.com/connectrpc/conformance/tree/main/internal/app/connectconformance/testsuites/data
[suite-proto]: https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/suite.proto
[client-compat-proto]: https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/client_compat.proto
[connectrpc-repo]: https://buf.build/connectrpc/conformance
