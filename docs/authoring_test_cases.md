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


This should describe the test suites folder and link to the YAML config schema by way of the relevant message in the BSR generated docs.

This should describe how the expected responses is auto-generated based on the request details. It usually only needs to be explicitly provided for exception test cases.

[testsuites]: https://github.com/connectrpc/conformance/blob/main/internal/app/connectconformance/testsuites/data
[suite-proto]: https://buf.build/connectrpc/conformance/file/main:connectrpc/conformance/v1/suite.proto
