# Configuring and Running Tests

To run the suite of conformance tests against a Connect, gRPC, or gRPC-Web
implementation, there are several steps.

1. Create a configuration file that describes the relevant features of the
   implementation. This is how the test runner knows what test cases to run.
2. Create an implementation under test.
   * If you are testing a client implementation, then you will need to implement
     a conformance client. This is a program that uses the implementation under
     test to send RPCs to a reference server. This program then records all of the
     results of the RPC (response data, errors, metadata, etc). The test runner
     analyzes those results, along with other feedback provided by the reference
     server, to decide on the correctness of your client implementation.
   * Similarly, if you are testing a server implementation, then you will need to
     implement a conformance server. This is a program that uses the implementation
     under test to provide an RPC service. This program merely responds to RPCs in
     a particular fashion. The RPCs will be sent by a reference client, and, like
     other conformance clients, it will report the results of each RPC. The test
     runner analyzes those results, along with other feedback provided by the
     reference client, to decide on the correctness of your server implementation.
3. Run the test suite! There will likely be some troubleshooting when the
   conformance client or server is initially written.
   * This typically involves running
     tests repeatedly and fixing issues, some of which may be due to bugs in the
     conformance client or server code, and some of which may be a result of bugs in
     the implementation under test.
   * You may choose to not fix 100% of test cases, in which case you will create
     separate files that configure "known failing" and/or "known flaky" test cases.
4. Configure continuous integration of the conformance test suite, so you can know
   whenever a change in your implementation might inadvertently cause it to no
   longer conform to the RPC specs.

The sections below discuss all but the second bullet. That is the meatiest topic, so
there are separate guides for how to implement the conformance client and server.

## Configuration Files

When running conformance tests, you will first create a YAML file that describes
the features that are supported and need to be tested. The schema for this YAML file
is defined by the [`Config`][config-proto] Protobuf message. The representation of a Protobuf
message in YAML is the same as its [JSON format][json-docs].

This message is defined in the same module as messages that describe test suites and
the RPC service that a conformance client uses and that a conformance server provides.
They are all found in the [connectrpc/conformance module][connectrpc-repo] in the BSR.

The file has three top-level keys:
1. `features`: This is a set of supported capabilities of the implementation under test.
   These features are used to create a set of "config cases". They are like axes in a
   table, and each cell in the table is a config case. So the supported features define
   which config cases are relevant to your implementation. For example, if your
   implementation supports Connect and gRPC protocols (but not gRPC-Web), HTTP 1.1 and
   HTTP/2, and both "proto" and "json" message encoding, then a config case for the
   Connect protocol over HTTP 1.1 using "json" applies to your implementation. But a
   config case for the gRPC-Web protocol over HTTP/3 does not.
2. `include_cases`: This is a set of config cases whose test cases should be run against
   your implementation, _in addition to_ config cases implied by the set of supported
   features.
3. `exclude_cases`: This is a set of config cases whose test cases should **not** be run
   against your implementation, even if they otherwise match the set of supported
   features.

### Features

The YAML config stanza named "features" represents a [`Features`][features-proto]
Protobuf message and has the following fields, each of which configures a single
feature of the implementation under test.

* `versions`: This configures which HTTP protocol versions that the implementation
  supports. Valid options are `HTTP_VERSION_1`, for HTTP 1.1, `HTTP_VERSION_2` for
  HTTP/2, and `HTTP_VERSION_3` for HTTP/3. Note that HTTP/3 _requires_ TLS. If not
  configured, support is assumed for HTTP 1.1 and HTTP/2.
* `protocols`: This configures which RPC protocol that the implementation supports.
  The options are `PROTOCOL_CONNECT`, `PROTOCOL_GRPC`, and `PROTOCOL_GRPC_WEB`. Note
  that gRPC _requires_ HTTP/2. The other two (Connect and gRPC-Web) can work with
  any version of HTTP. If not configured, support is assumed for all three.
* `codecs`: This configures which codecs, or message formats, that the implementation
  supports. The options are `CODEC_PROTO` (which corresponds to the sub-format "proto",
  which is the Protobuf binary format) and `CODEC_JSON` (sub-format "json"). If not
  configured, support is assumed for both.
* `compressions`: This configures which compression encodings are supported by the
  implementation. `COMPRESSION_IDENTITY`, the "identity" encoding, means no compression.
  `COMPRESSION_GZIP` indicates "gzip" compression. If not configured, support for these
  two is assumed. Other valid options include `COMPRESSION_BR`, `COMPRESSION_ZSTD`,
  `COMPRESSION_DEFLATE`, and `COMPRESSION_SNAPPY`. (These comprise all of the compression
  schemes indicated in the protocols specs for [Connect][connect-protocol] and
  [gRPC][grpc-protocol].)
* `stream_types`: This configures which stream types the implementation supports. If not
  configured, support is assumed for all types. The valid options are
  `STREAM_TYPE_UNARY`, `STREAM_TYPE_CLIENT_STREAM`, `STREAM_TYPE_SERVER_STREAM`,
  `STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`, and `STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`. The
  latter two both use bidirectional streams. The former (half-duplex) is technically
  compatible with HTTP 1.1, while the other (full-duplex) _requires_ HTTP/2 or HTTP/3.
* `supports_h2c`: This is a simple flag that indicates whether or not the implementation
  supports "H2C" or "HTTP/2 over Clear text". This indicates the use of HTTP/2 without
  TLS. Normally, support for HTTP/2 is negotiated between clients and servers during the
  TLS handshake. But H2C requires prior knowledge that the server supports HTTP/2. If not
  configured, it is assumed that the implementation _does_ support H2C.
* `supports_tls`: This flag indicates whether the implementation supports TLS. If not
  configured, it is assumed that the implementation _does_ support TLS. If TLS is not
  supported, HTTP/3 cannot be supported and HTTP/2 can only be supported if the
  implementation supports H2C.
* `supports_tls_client_certs`: This flag indicates whether the implementation supports
  TLS client certificates. For servers, this means that the server will require that
  clients present trusted certificates during the TLS handshake. For clients, this means
  that they can present certificates to the server during the handshake. If not
  configured, it is assumed that implementations do _not_ support client certificates.
* `supports_trailers`: This flag indicates whether the implementation can support the
  use of HTTP trailers. If false, then the gRPC protocol cannot be supported. If not
  configured, it is assumed that the implementation _does_ support trailers.
* `supports_half_duplex_bidi_over_http1`: This flag indicates whether the implementation
  supports bidirectional streams over HTTP 1.1. Only half-duplex streams can be supported
  this way. If false, bidirectional streams (regardless of whether they are half- or
  full-duplex) are only supported over HTTP/2 or HTTP/3. If not configured, it is
  assumed that the implementation does _not_ support bidirectional streams over HTTP 1.1.
* `supports_connect_get`: This flag indicates whether the implementation supports the use
  of the "GET" HTTP method with the Connect protocol. This allows certain methods (those
  that are side-effect-free) to be invoked using the GET method, with the request
  parameters encoded in the query string, which can enable infrastructure support for
  things like retries, caching, and conditional queries. If not configured, it is
  assumed that the implementation _does_ support the GET method with the Connect protocol.
  This does not need to be configured if the implementation only supports the gRPC and/or
  gRPC-Web protocols.
* `supports_message_receive_limit`: This flag indicates whether the implementation
  supports configuration to limit the size of messages that it receives. Such configuration
  is intended to aid operations, to prevent excessive memory usage and resource consumption
  in workloads. For servers, this means an option that limits the size of request data; for
  clients, this option limits the size of response data. The limit is applied on a per
  message basis, so it does not limit the total amount of data transferred in a streaming
  operation. The limit should apply to the _uncompressed_ size of a message. So even if the
  message is smaller than the limit on the wire, when compressed, it should be rejected if
  it would exceed the limit when uncompressed. If not configured, it is assumed that the
  implementation _does_ support a limit.

### Config Cases

A single "config case" is represented by the [`ConfigCase`][configcase-proto] Protobuf
message. The result of applying features is a matrix of config cases. You can opt into
additional config cases in the config file, and you can also opt out of some cases that
are otherwise implied by the features.

A single case is defined by the following properties:
* `version`: A single HTTP version -- HTTP 1.1, HTTP/2, or HTTP/3.
* `protocol`: An RPC protocol -- Connect, gRPC, or gRPC-Web.
* `codec`: A codec, such as "proto" or "json".
* `compression`: A compression encoding.
* `stream_type`: A stream type.
* `use_tls`: Whether TLS is in use. If false, plain-text connections are used.
* `use_tls_client_certs`: Whether TLS client certificates are in use. Must be false if
  `use_tls` is false.
* `use_message_receive_limit`: Whether a message receive limit is in use.

A single set of features is expanded into one or more (usually many more) config cases.
For example, if the features support HTTP 1.1 and HTTP/2, all three protocols, all
stream types, identity and gzip encoding, and TLS, that results in 2×3×5×2×2 = 120
combinations. Some of those combinations may not be valid (such as full-duplex
bidirectional streams over HTTP 1.1, or gRPC over HTTP 1.1), so the total number of
config cases would be close to 120 but not quite.

Config cases can be described in the config YAML, to indicate config cases to test in
addition to those computed just from the combination of features or to opt out of some
config cases that are otherwise implied by the features.

When in a YAML file, any of the above properties may be omitted, in which case that
property is treated as a wildcard. Take the following config case for example:
```yaml
version: HTTP_VERSION_2
codec: CODEC_PROTO
stream_type: STREAM_TYPE_SERVER_STREAM
```
Since the protocol is not specified, this effectively expands to config cases with
the above properties for all three protocols. Similarly, since it does not indicate
whether the case is for TLS or not, it expands into config cases that represent TLS
and those that do not.

## Running Tests

Running the tests is done using the `connectconformance` binary. This binary can be
downloaded from the "Assets" section of a [GitHub release][releases]. It can also be
built using the `go` tool, if you have a Go SDK:

```shell
go install connectrpc.com/conformance/cmd/connectconformance@latest
```

Once downloaded/installed, you can run the tests like so:
```shell
> connectconformance \
    --conf ./path/to/client/config.yaml \
    --mode client \
    -- \
    ./path/to/client/program --some-flag-for-client-program
```

Let's dissect this line-by-line:
* `connectconformance`: The executable program we are running, of course.
* `--conf ./path/to/client/config.yaml`: The path to the YAML file that defines the
  features of our implementation under test as well as any other config cases to
  opt into or out of.
* `--mode client`: The test mode. In this case, we are testing a **client** implementation
  so the mode is "client". If we want to run conformance tests against a server
  implementation, the mode would be "server".
* `--`: This is just a separator; everything after is a _positional_ argument. This is
  not strictly needed, but if we want to pass options to the program under test, and
  we don't want the test runner to confuse them for its own options, we use this
  separator. This tells the test runner that there are no more options after this and
  anything that _looks_ like an option, is actually a positional argument.
* `./path/to/client/program --some-flag-for-client-program`: The positional arguments
  represent the command to invoke in order to run the client under test. The first
  token must be the path to the executable. Any subsequent arguments are passed as
  arguments to that executable. So in this case, `--some-flag-for-client-program` is
  an option that our client under test understands.

Common reasons to pass arguments to the client or server under test are:
1. To control verbosity of log output. When troubleshooting an implementation, it
   can be useful to enable verbose output in the program under test, to provide
   visibility into what it's doing (and what might be going wrong).
2. For clients, to control parallelism. Clients under test can issue RPCs to the
   server concurrently in order to speed up the test run. When a large number of
   test case permutations apply to a particular implementation, it can take time to
   issue all of the RPCs if done serially from a single thread. So a client may choose
   to send RPCs in parallel to speed things up. Since RPCs involve network I/O, they
   are typically amenable to high parallelism. Timeout test cases can involve the
   client and server delaying or sleeping between actions, which makes them even more
   amenable to aggressive parallelism. (Care must be taken that results are recorded
   sequentially or else the output stream may be corrupted and unprocessable by the 
   test runner.)
3. To control other implementation-specific aspects.
   * For example: an implementation may allow users to swap between different underlying
     HTTP libraries. So to get conformance test coverage with both libraries, instead of
     creating a separate program for each, create a single program that accepts an
     option that tells it which library to use.
   * Another example: an implementation my provide alternate code generation to provide
     sync/blocking vs. async paradigms for invoking RPCs. Instead of creating separate
     programs for each, an option could be used to have a single program select which
     form of generated code to use under-the-hood.

The `connectconformance` program will exit with a zero status code to indicate success.
It will exit with a non-zero code when there are errors. By default, the program does not
print any output except at the very end, printing a list of failing test cases and a
summary.

### Test Output

When test cases fail, the test runner prints a `FAILED` banner with the _full name_ of the
test case permutation and the errors that were observed with the RPC. If the `--trace` option
is provided to the test runner, then it may also show a full diagnostic trace of the HTTP
request and response. Here's an example (with an HTTP trace example, too):
```text
FAILED: Client Cancellation/HTTPVersion:1/Protocol:PROTOCOL_GRPC_WEB/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/server-stream/cancel-after-responses:
	expecting an error but received none
	expecting 1 response messages but instead got 2
---- HTTP Trace ----
 request>     0.000ms POST http://.../connectrpc.conformance.v1.ConformanceService/ServerStream HTTP/1.1
 request>             Accept: */*
 request>             Accept-Encoding: gzip, deflate, br
 request>             Accept-Language: en-US,en;q=0.9
 request>             Connection: keep-alive
 request>             Content-Length: 40
 request>             Content-Type: application/grpc-web+proto
 request>             Origin: null
 request>             Sec-Ch-Ua: "Chromium";v="119", "Not?A_Brand";v="24"
 request>             Sec-Ch-Ua-Mobile: ?0
 request>             Sec-Ch-Ua-Platform: "macOS"
 request>             Sec-Fetch-Dest: empty
 request>             Sec-Fetch-Mode: cors
 request>             Sec-Fetch-Site: cross-site
 request>             User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/119.0.0.0 Safari/537.36
 request>             X-Expect-Codec: 1
 request>             X-Expect-Compression: 1
 request>             X-Expect-Http-Method: POST
 request>             X-Expect-Http-Version: 1
 request>             X-Expect-Protocol: 3
 request>             X-Expect-Tls: false
 request>             X-Grpc-Web: 1
 request>             X-Test-Case-Name: Client Cancellation/HTTPVersion:1/Protocol:PROTOCOL_GRPC_WEB/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/server-stream/cancel-after-responses
 request>             X-User-Agent: grpc-web-javascript/0.1
 request>
 request>     0.020ms message #1: prefix: flags=0, len=35
 request>             message #1: data: 35/35 bytes
 request>     0.020ms body end
response<   201.121ms 200 OK
response<             Access-Control-Allow-Origin: *
response<             Access-Control-Expose-Headers: *
response<             Content-Type: application/grpc-web+proto
response<             Grpc-Accept-Encoding: zstd,snappy,deflate,br,gzip
response<             Server: connectconformance-referenceserver/v1.0.0-rc3
response<             Vary: Origin
response<
response<   201.123ms message #1: prefix: flags=0, len=1072
response<             message #1: data: 1072/1072 bytes
response<   402.243ms message #2: prefix: flags=0, len=17
response<             message #2: data: 17/17 bytes
response<   402.301ms message #3: prefix: flags=128, len=32
response<             message #3: data: 32/32 bytes
response<               eos: grpc-message: 
response<               eos: grpc-status: 0
response<               eos: 
response<   402.359ms body end
--------------------
```

The top-line in the example output above shows the full name of the test case. The first and final
components of the name tell us that it is in the "Client Cancellation" test suite and is a test case
that is named "server-stream/cancel-after-responses". The rest of the name identifies which
_permutation_ of the test failed.

The trace shows the full HTTP headers (and trailers, if any) and summarizes the request and response
body data in the form of messages exchanged. Each message is typically sent with a five-byte prefix
and then the actual message payload. It also shows timestamps, relative to the start of the RPC (so
the first event, the request line and request headers, is always at time zero). The final message in
Connect streaming RPCs and gRPC-Web is a special "end of stream" message, which is also shown in the
trace, with an "eos:" prefix before each line.

If a test cases fails that is **known** to fail, it is printed with an `INFO` banner, to remind
you that there are failing test cases, even if the test run is successful.

After printing the above information for any failed test cases, the test runner then prints a
summary like so:
```text
Total cases: 602
598 passed, 0 failed
(4 failed as expected due to being known failures.)
```
The above states that 598 out of 602 test cases were successful. The remaining four are "known
failing" cases. So they aren't count as failures, and therefore this run of the conformance
tests was successful.

If you provide a `-v` option to the test runner, it will print some other messages as it is
running:
```text
Computed 44 config case permutations.
Loaded 8 test suite(s), 97 test case template(s).
Loaded 1 known failing test case pattern(s) that match 4 test case permutation(s).
Computed 602 test case permutation(s) across 10 server configuration(s).
Running 47 tests with reference server for server config {HTTP_VERSION_1, PROTOCOL_CONNECT, TLS:false}...
Running 47 tests with reference server for server config {HTTP_VERSION_1, PROTOCOL_CONNECT, TLS:true}...
Running 46 tests with reference server for server config {HTTP_VERSION_1, PROTOCOL_GRPC_WEB, TLS:false}...
Running 46 tests with reference server for server config {HTTP_VERSION_1, PROTOCOL_GRPC_WEB, TLS:true}...
Running 47 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_CONNECT, TLS:false}...
Running 47 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_CONNECT, TLS:true}...
Running 46 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_GRPC, TLS:false}...
Running 46 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_GRPC, TLS:true}...
Running 46 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_GRPC_WEB, TLS:false}...
Running 46 tests with reference server for server config {HTTP_VERSION_2, PROTOCOL_GRPC_WEB, TLS:true}...
Running 46 tests with reference server (grpc) for server config {HTTP_VERSION_1, PROTOCOL_GRPC_WEB, TLS:false}...
Running 46 tests with reference server (grpc) for server config {HTTP_VERSION_2, PROTOCOL_GRPC, TLS:false}...
Running 46 tests with reference server (grpc) for server config {HTTP_VERSION_2, PROTOCOL_GRPC_WEB, TLS:false}...
```
This shows a summary of the config as it is loaded and processed, telling us the total number of
[config cases](#config-cases) that apply to the current configuration (44), the total number of test suites (8),
and the total number of test cases across those suites (97). It then shows the number of patterns
provided to identify "known failing" cases (1), and the number of test cases that matched the "known
failing" patterns (4). The next line shows us that it has used the 44 relevant config cases and 97
test case templates to compute a total of 602 [test case permutations](#test-case-permutations). This
means that the client under test will be invoking 602 RPCs.

The remaining lines in the example output above are printed as each test server is started. Each server config
represents a different RPC server, started with the given configuration (since we are running the tests using
`--mode client`, it is starting a reference server). In the above example, 47 test case permutations apply
to the Connect protocol; 46 apply to the gRPC and gRPC-Web protocols. If you add up all of those numbers
(47+47+46+46+...), the result is 602: the total number of test case permutations being run.

### Test Case Permutations

As mentioned above, a single test case can turn into multiple permutations, where the same RPC is used
with multiple configuration cases.

The full name of the example failing test above follows:
```
Client Cancellation/HTTPVersion:1/Protocol:PROTOCOL_GRPC_WEB/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/server-stream/cancel-after-responses
```
We can decipher this name by looking at the various _components_ of the name (components are separated
by slashes `/`).

* `Client Cancellation`: The first component is the name of the test suite that defines this test case.
* `HTTPVersion:1`: Path elements with a colon represent one axis of a permutation. For this one, the
  test case is relevant to all HTTP versions, so it will run against all versions supported by the
  implementation under test. This component of the name tells us that the failure occurred when run
  using version 1 (which is really HTTP 1.1).
* `Protocol:PROTOCOL_GRPC_WEB`: Similar to above, this test case is relevant to all protocols. This
  failure occurred when using the gRPC-Web protocol.
* `Codec:CODEC_PROTO`: The failure occurred using the "proto" codec (the Protobuf binary format).
* `Compression:COMPRESSION_IDENTITY`: The failure occurred using "identity" encoding, which actually
  means no compression.
* `TLS:false`: The failure occurred over a plain-text HTTP connection, not a secure one.
* `server-stream/cancel-after-responses`: After the permutation properties, the final element(s) are
  the actual test case name, as it is appears in the test suite YAML. As in this example, this name
  can contain a slash and therefore represent more than just the final name component. This naming
  convention is amenable to wildcards in test case patterns (useful for selecting which tests to
  run or marking test cases as known to fail).

Test cases that are run against a [gRPC implementation](#grpc-implementations) will have an additional
component in their full name: `(grpc server impl)` or `(grpc client impl)`. The former is used in "client"
mode, when the client under test sends an RPC to the gRPC server implementation; the latter is used in
"server" mode, when it's the gRPC client implementation that is sending the RPC to the server under test.

### gRPC Implementations

The lines in the example verbose output above that say "reference server (grpc)" are for test cases using
a different server implementation -- one that was written with an official [gRPC][grpc] implementation
instead of using a [ConnectRPC][connectrpc] implementation. When your client or server supports the gRPC
protocol, it gets run against both reference implementations, as a way of further verifying
interoperability with the ecosystem. (Note that the standard reference implementations _also_ support
the gRPC protocol; so the gRPC test cases are repeated with a different server.)

### Selecting Test Cases

The `connectconformance` test runner supports four different options for selecting which test cases to
run. Two of them are intended for interactive runs and the other two are suggested for use with test
runs during continuous integration testing.

1. `--known-failing`: The most likely option you may need to use is to mark tests that are known to
   fail. When an implementation is not 100% compliant, this can be used to opt out of the test cases
   that do not pass. With this flag, these test cases will be run, but expected to fail. In fact, if
   they happen to pass, the test run will be considered a failure since it means that the known
   failing configuration is stale and needs to be updated to match recent fixes.
2. `--known-flaky`: This is similar to above, but is for tests that do not consistently fail. This
   is often indicative of timing-related and concurrency bugs, where the sequence of execution can
   be non-deterministic, so sometimes a test passes and sometimes it fails. In this mode, the test
   cases are still run, but allowed to fail. Whether the test case passes or fails does not cause
   the whole test run to pass or fail. But if it does fail, it will be logged in the test output.
3. `--run`: This option is intended for interactive runs, like when troubleshooting particular
   test cases. Instead of running the entire suite, you can run just select test cases.
4. `--skip`: This option is also intended for interactive runs. It is the  opposite of `--run`
   and allows you to skip execution of certain test cases. When combined with `--run`, you can
   use the `--run` flag to enumerate which cases are considered, and then `--skip` refines
   the set, skipping some of the test cases that otherwise match a `--run` pattern.

All four of these options accept two kinds of values:
1. A _test case pattern_. This is like a full test case name but it can have wildcards. An
   asterisk (`*`) matches one name component. A double-asterisk (`**`) matches zero or more
   components. A name component is the portion of the name between `/` separators. Wildcards
   can**not** be used to match partial name components. For example `foo/bar*` is not valid
   since the wildcard is expected to only match a suffix. If such a pattern is used, the
   asterisk is matched exactly instead of being treated as a wildcard.
2. If prefixed with an at-sign (`@`), the rest of the value is the path of a file that
   contains test case patterns, one per line. When the file is processed, leading and trailing
   whitespace is discarded from each line, blank lines are ignored, and lines that start with
   a pound-sign (`#`) are treated as comments and ignored.

All four of these options can be provided multiple times on the command-line, to provide
multiple test case patterns, refer to multiple files, or both.

It is strongly recommended to only use `--known-failing` in CI configurations. For legitimately
flaky test cases, use `--known-flaky` (instead of `--skip`). Use of `--run` or `--skip` in CI
configurations is discouraged. It should instead be possibly to correctly filter the set of tests
to run just based on config YAML files.

One reason one might need to use `--skip` in a CI configuration is if a bug in the implementation
under test causes the client or server to crash or to deadlock. Since such bugs could prevent the
conformance suite from ever completing successfully (even if such tests are marked as "known
failing"), it may be necessary to temporarily skip them in CI until those bugs are fixed.

## Configuring CI

The easiest way to run conformance tests as part of CI is to do so from a container that has the
Go SDK. The conformance tests currently requires Go 1.20 or higher. The `connectconformance` program
that runs the tests can be installed via a simple invocation of the Go tool:
```shell
go install connectrpc.com/conformance/cmd/connectconformance@latest
```
This binary _embeds_ all the test cases in the repo and also embeds the code for the reference
clients and servers. So this single binary is all that is needed to run the entire suite against
either a client or server under test.

If CI uses a container that does _not_ have the Go SDK, then you can download a self-contained
binary from the GitHub release artifacts. The artifacts are named so that the output of `uname`
can be used to compute the URL of the file to download. Running `uname -s` will provide the OS
and then `uname -m` provides the machine architecture, or ARCH. The full URL of a release artifact
then is the following, where VERSION indicates which release version of the Conformance tests to
download:

```text
https://github.com/connectrpc/conformance/releases/download/${VERSION}/connectconformance-${VERSION}-${OS}-${ARCH}.tar.gz
```

The above can easily fetched using `curl` or `wget` and can also be downloaded using HTTP client
capabilities available in most programming and scripting languages. The downloaded file is a compressed
TAR archive that contains an executable file named `connectconformance`. The following example shows
how this could be downloaded and then run from a shell script:

```bash
#!/bin/bash

OS=`uname -s`
ARCH=`uname -m`
VERSION=v1.0.0-rc3
URL=https://github.com/connectrpc/conformance/releases/download/${VERSION}/connectconformance-${VERSION}-${OS}-${ARCH}.tar.gz
mkdir tmp
# Download the artifact into the tmp directory
curl -sSL $URL -o - \
    | tar -x -z --directory tmp
# Now we can run the tests!
./tmp/connectconformance --mode client \
    --conf path/to/config.yaml \
    path/to/test/client
```

You will want the config YAML checked into source control. It is also recommended to put any known
failing test cases into a "known-failing.txt" that is also checked into source control and referenced
in the command that runs the test.

If you have multiple test programs, such as both a client and a server, or even a client with
different sets of arguments, you should name the relevant config YAML and known-failing files so
it is clear to which invocation they apply.

## Upgrading

When a new version of the conformance suite is released, ideally, you could simply update
the version number you are using and everything just works. We aim for
backwards-compatibility between releases to maximize the chances of this ideal outcome. But
there are a number of things that can happen in a release that make the process a little
more laborious:

* As a matter of hygiene/maintenance, we may rename and re-organize test suites and test
  cases. This means that any test case patterns that are part of your configuration (like
  known-failing files) may need to be updated. We don't expect this to happen often, but
  when it does, we will include information in the release notes to aid in updating your
  configuration.
* The new version may contain new/updated test cases that require some changes in the
  behavior/logic of your implementations under test. This might be for testing new
  functionality that requires new fields in the conformance protocol messages. Without
  changes in your client or server under test, the new test cases will likely fail.
* The new version may contain new/updated test cases that reveal previously undetected
  conformance failures.

To minimize disruption when upgrading, we recommend a process that looks like so:
1. Update to the new release of the conformance suite.
2. Update test case patterns (like in known-failing configurations) if necessary to match
   any changes to test case names and organization.
3. Update/add known-failing configurations for any new failures resulting from new/updated
   test cases.
4. **Commit/merge the upgrade.**
5. File bugs for the new failures.
6. As the bugs are fixed, update the known-failing configurations as you go.

By simply marking all new failures as "known failing" and filing bugs for them, it should
allow you to upgrade to a new release quickly. You can then decide on the urgency of fixing
the new failures and prioritize accordingly.

[config-proto]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.Config
[configcase-proto]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.ConfigCase
[connect-protocol]: https://connectrpc.com/docs/protocol/
[connectrpc]: https://connectrpc.com
[connectrpc-repo]: https://buf.build/connectrpc/conformance
[features-proto]: https://buf.build/connectrpc/conformance/docs/main:connectrpc.conformance.v1#connectrpc.conformance.v1.Features
[grpc]: https://grpc.io
[grpc-protocol]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
[grpc-web-protocol]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md
[json-docs]: https://protobuf.dev/programming-guides/proto3/#json
[releases]: https://github.com/connectrpc/conformance/releases
