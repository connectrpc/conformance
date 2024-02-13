# Authoring test cases

Test cases for the conformance runner are configured in YAML files and are located [here](testsuites). Each file 
represents a single suite of tests. 

Authoring new test cases involves adding either a new file if a new suite is required or added an additional test case to an 
existing file/suite. 

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

 > In the case of Bidi tests, you should also add `full_duplex` or `half_duplex` to the test name. 

 For example:

`unary/error`
`server-stream/error-with-responses`
`bidi-stream/full-duplex/stream-error-returns-success-http-code`

If a stream contains tests for just a single stream type, the stream type can be omitted from the test name.

These conventions allow for more granular control over running tests via the conformance runner, such as only running tests
for a specific protocol or only running the unary tests within a suite.

## Configuring test suites

A suite can be configured with various directives which will apply to all tests within that suite. For example, you 
can limit a suite to only run via the Connect and gRPC-web protocols using:

```yaml
relevantProtocols:
  - PROTOCOL_CONNECT
  - PROTOCOL_GRPC_WEB
```

or you can limit a suite only to the JSON codec:

```yaml
relevantCodecs:
  - CODEC_JSON
```

For a full list of the available configurations for a test suite, see the [Protobuf definition](suite-proto) in the BSR.

## Expected responses

The expected response for a test is auto-generated based on the request details. The conformance runner will determine what the response 
should be by the values specified in the test suite and individual test cases. However, you have the ability to explicitly specify your 
own expected response directly in the test definition itself. To do so, simply define an `expectedResponse` block for your test case and this will
override the auto-generated expected response in the test runner. This typically only needs done for exception test cases. For examples,
search the [test suites](testsuites) directory for `expectedResponse`.

[testsuites]: https://github.com/connectrpc/conformance/tree/main/internal/app/connectconformance/testsuites/data
[suite-proto]: https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/suite.proto
