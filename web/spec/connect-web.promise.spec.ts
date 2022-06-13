// Copyright 2022 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {
  ConnectError,
  createConnectTransport,
  decodeBinaryHeader,
  encodeBinaryHeader,
  makePromiseClient,
  StatusCode,
} from "@bufbuild/connect-web";
import {
  TestService,
  UnimplementedService,
} from "../gen/proto/connect-web/grpc/testing/test_connectweb";
import { Empty } from "../gen/proto/connect-web/grpc/testing/empty_pb";
import {
  SimpleRequest,
  StreamingOutputCallRequest,
} from "../gen/proto/connect-web/grpc/testing/messages_pb";

// Unfortunately there's no typing for the `__karma__` variable. Just declare it as any.
// eslint-disable-next-line no-underscore-dangle, @typescript-eslint/no-explicit-any
declare const __karma__: any;

describe("connect_web_promise_client", function () {
  const host = __karma__.config.host;
  const port = __karma__.config.port;
  const transport = createConnectTransport({
    baseUrl: `https://${host}:${port}`,
  });
  const client = makePromiseClient(TestService, transport);
  it("empty_unary", async function () {
    const response = await client.emptyCall({});
    expect(response).toEqual(new Empty());
  });
  it("empty_unary_with_timeout", async function () {
    const deadlineMs = 1000; // 1 second
    const response = await client.emptyCall({}, { timeout: deadlineMs });
    expect(response).toEqual(new Empty());
  });
  it("large_unary", async function () {
    const size = 314159;
    const req = new SimpleRequest({
      responseSize: size,
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
    });
    const response = await client.unaryCall(req);
    expect(response.payload).toBeDefined();
    expect(response.payload?.body.length).toEqual(size);
  });
  it("server_streaming", async function () {
    const sizes = [31415, 9, 2653, 58979];
    const responseParams = sizes.map((size, index) => {
      return {
        size: size,
        intervalUs: index * 10,
      };
    });
    let responseCount = 0;
    for await (const response of await client.streamingOutputCall({
      responseParameters: responseParams,
    })) {
      expect(response.payload).toBeDefined();
      expect(response.payload?.body.length).toEqual(sizes[responseCount]);
      responseCount++;
    }
    expect(responseCount).toEqual(sizes.length);
  });
  it("empty_stream", async function () {
    try {
      for await (const response of await client.streamingOutputCall({
        responseParameters: [],
      })) {
        fail(`expecting no response in the empty stream, got: ${response}`);
      }
    } catch (e) {
      fail(`expecting no error in the empty stream, got: ${e}`);
    }
  });
  it("custom_metadata", async function () {
    const size = 314159;
    const ECHO_LEADING_KEY = "x-grpc-test-echo-initial";
    const ECHO_LEADING_VALUE = "test_initial_metadata_value";
    const ECHO_TRAILING_KEY = "x-grpc-test-echo-trailing-bin";
    const ECHO_TRAILING_VALUE = new Uint8Array([0xab, 0xab, 0xab]);

    const req = new SimpleRequest({
      responseSize: size,
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
    });
    const response = await client.unaryCall(req, {
      headers: {
        [ECHO_LEADING_KEY]: ECHO_LEADING_VALUE,
        [ECHO_TRAILING_KEY]: encodeBinaryHeader(ECHO_TRAILING_VALUE),
      },
      onHeader(header) {
        expect(header.has(ECHO_LEADING_KEY)).toBeTrue();
        expect(header.get(ECHO_LEADING_KEY)).toEqual(ECHO_LEADING_VALUE);
      },
      onTrailer(trailer) {
        expect(trailer.has(ECHO_TRAILING_KEY)).toBeTrue();
        expect(decodeBinaryHeader(trailer.get(ECHO_TRAILING_KEY)||"")).toEqual(ECHO_TRAILING_VALUE);
      },
    });
    expect(response.payload).toBeDefined();
    expect(response.payload?.body.length).toEqual(size);
  });
  it("custom_metadata_server_streaming", async function () {
    const ECHO_LEADING_KEY = "x-grpc-test-echo-initial";
    const ECHO_LEADING_VALUE = "test_initial_metadata_value";
    const ECHO_TRAILING_KEY = "x-grpc-test-echo-trailing-bin";
    const ECHO_TRAILING_VALUE = new Uint8Array([0xab, 0xab, 0xab]);

    const size = 31415;
    const responseParams = [{
      size: size,
    }]
    for await (const response of await client.streamingOutputCall({
      responseParameters: responseParams,
    }, {
      headers: {
        [ECHO_LEADING_KEY]: ECHO_LEADING_VALUE,
        [ECHO_TRAILING_KEY]: encodeBinaryHeader(ECHO_TRAILING_VALUE),
      },
      onHeader(header) {
        expect(header.has(ECHO_LEADING_KEY)).toBeTrue();
        expect(header.get(ECHO_LEADING_KEY)).toEqual(ECHO_LEADING_VALUE);
      },
      onTrailer(trailer) {
        expect(trailer.has(ECHO_TRAILING_KEY)).toBeTrue();
        expect(decodeBinaryHeader(trailer.get(ECHO_TRAILING_KEY)||"")).toEqual(ECHO_TRAILING_VALUE);
      },
    })) {
      expect(response.payload).toBeDefined();
      expect(response.payload?.body.length).toEqual(size);
    }
  });
  it("status_code_and_message", async function () {
    const TEST_STATUS_MESSAGE = "test status message";
    const req = new SimpleRequest({
      responseStatus: {
        code: StatusCode.Unknown,
        message: TEST_STATUS_MESSAGE,
      },
    });
    try {
      await client.unaryCall(req);
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError).code).toEqual(StatusCode.Unknown);
      expect((e as ConnectError).rawMessage).toEqual(TEST_STATUS_MESSAGE);
    }
  });
  it("special_status", async function () {
    const TEST_STATUS_MESSAGE = `\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n`;
    const req = new SimpleRequest({
      responseStatus: {
        code: StatusCode.Unknown,
        message: TEST_STATUS_MESSAGE,
      },
    });
    try {
      await client.unaryCall(req);
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError).code).toEqual(StatusCode.Unknown);
      expect((e as ConnectError).rawMessage).toEqual(TEST_STATUS_MESSAGE);
    }
  });
  it("timeout_on_sleeping_server", async function () {
    const request = new StreamingOutputCallRequest({
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
      responseParameters: [
        {
          size: 31415,
          intervalUs: 5000,
        },
      ],
    });
    try {
      for await (const response of await client.streamingOutputCall(request, {
        timeout: 1,
      })) {
        fail(`expecting no response from sleeping server, got: ${response}`);
      }
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      // We expect this to be DEADLINE_EXCEEDED, however envoy is monitoring the stream timeout
      // and will return an HTTP status code 408 when stream max duration time reached, which
      // cannot be translated to a connect error code, so connect-web client throws an Unknown.
      expect(
        [StatusCode.Unknown, StatusCode.DeadlineExceeded].includes(
          (e as ConnectError).code
        )
      ).toBeTrue();
    }
  });
  it("unimplemented_method", async function () {
    try {
      await client.unimplementedCall({});
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError).code).toEqual(StatusCode.Unimplemented);
    }
  });
  it("unimplemented_server_streaming_method", async function () {
    try {
      for await (const response of await client.unimplementedStreamingOutputCall({})) {
        fail(`expecting no response from fail server streaming, got: ${response}`);
      }
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError).code).toEqual(StatusCode.Unimplemented);
    }
  });
  it("unimplemented_service", async function () {
    const badClient = makePromiseClient(UnimplementedService, transport);
    try {
      await badClient.unimplementedCall({});
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      // We expect this to be either Unimplemented or NotFound, depending on the implementation.
      // In order to support a consistent behaviour for this case, the backend would need to
      // own the router and all fallback behaviours. Both statuses are valid returns for this
      // case and the client should not retry on either status.
      expect(
        [StatusCode.Unimplemented, StatusCode.NotFound].includes(
          (e as ConnectError).code
        )
      ).toBeTrue();
    }
  });
  it("fail_unary", async function () {
    try {
      await client.failUnaryCall({});
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError).code).toEqual(StatusCode.ResourceExhausted);
      expect((e as ConnectError).rawMessage).toEqual("soirée 🎉");
    }
  });
  it("fail_server_streaming", async function () {
    const sizes = [31415, 9, 2653, 58979];
    const responseParams = sizes.map((size, index) => {
      return {
        size: size,
        intervalUs: index * 10,
      };
    });
    try {
      for await (const response of await client.failStreamingOutputCall({
        responseParameters: responseParams,
      })) {
        fail(`expecting no response from fail server streaming, got: ${response}`);
      }
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect((e as ConnectError)?.code).toEqual(StatusCode.ResourceExhausted);
      expect((e as ConnectError)?.rawMessage).toEqual("soirée 🎉");
    }
  });
});
