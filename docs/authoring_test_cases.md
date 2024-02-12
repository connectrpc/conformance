## Authoring test cases

Test cases for the conformance runner are configured in YAML files and are located [here](testsuites). The basic structure is 
that each file represents a single suite of tests. A suite can be configured with various directives which will apply to all
tests within that suite.

For example, you can limit a suite to only run via the Connect and gRPC-web protocols using:

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

Authoring new test cases involves adding either a new file if a new suite is required or added an additional test case to an 
existing file/suite. Suites are loosely defined and are grouped based on various criteria. When deciding whether to create a new
suite or add to an existing, the first factor is usually whether the tests apply to a single protocol or all of them. Tests that apply
to a single protocol are located in files with the protocol as the prefix. For example, `connect_idempotency.yaml` contains tests that
test GET requests for the Connect protocol only. The `grpc_web_client.yaml` contains client tests for the gRPC-web protocol. Files that
are not prefixed with a protocol name apply to all protocols.

### Naming conventions

The suites and tests within follow a loose naming convention:

#### Test files

Test files should be named according to the suite inside and the general functionality being tested. In addition:

* If a suite applies only to a certain protocol, the file name should be prefixed with that protocol. If the suite applies
  to all protocols, this can be omitted.
* If a suite only contains client or server tests, the file name prefix should include `client` or `server`. If the suite is
  for both client and server, this can be omitted.


For example: `connect_idempotency.yaml` contains a suite for testing idempotency (`GET` support) for the Connect protocol only.
The `client_message_size.yaml` file contains a suite of client tests only across all protocols. The `connect_client_code_to_http_code.yaml` 
file is comprised of client tests for the Connect protocol only.


#### Tests

Tests should be named according to the following convention:

`{stream type}/{test_description}`

In the case of Bidi tests, you should also add `full_duplex` or `half_duplex` to the test name. For example:

`unary/`
`server-stream/`
`bidi-stream/full-duplex/`

The above conventions allow for a more granular control over running tests via the conformance runner. For example, you can run
only unary tests within a file xxxxx, you can only run suites that apply to the Connect protocol, or you can specify known failing tests for xxxx.

### Expected responses

The expected response for a test is auto-generated based on the request details. The conformance runner will determine what the response 
should be by the values specified in the test suite and individual test cases. However, you have the ability to explicitly specify your 
own expected response directly in the test itself. To do so, simply define an `expectedResponse` block for your test case and this will
override the auto-generated expected response in the test runner. This typically only needs done for exception test cases. For an example,
take a look at xxxxxxxxxxxxxxxx.

[testsuites]: https://github.com/connectrpc/conformance/blob/main/internal/app/connectconformance/testsuites/data
[suite-proto]: https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/suite.proto
