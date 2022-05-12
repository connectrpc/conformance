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
  makePromiseClient,
  StatusCode,
} from "@bufbuild/connect-web";
import { ClientInterceptor } from "@bufbuild/connect-web/dist/types/client-interceptor";
import {
  TestService,
  UnimplementedService,
} from "../gen/proto/connect-web/grpc/testing/test_connectweb";
import { Empty } from "../gen/proto/connect-web/grpc/testing/empty_pb";
import {
  SimpleRequest,
  StreamingOutputCallRequest,
} from "../gen/proto/connect-web/grpc/testing/messages_pb";

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
  // TODO: enable this test when we have a fix on connect-web
  xit("empty_stream", async function () {
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
    // TODO: adjust this test once we land on the API for reading response headers and trailers
    const size = 314159;
    const ECHO_INITIAL_KEY = "x-grpc-test-echo-initial";
    const ECHO_INITIAL_VALUE = "test_initial_metadata_value";
    const ECHO_TRAILING_KEY = "x-grpc-test-echo-trailing-bin";
    const ECHO_TRAILING_VALUE = 0xababab;

    const req = new SimpleRequest({
      responseSize: size,
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
    });
    const metadata = {
      headers: {
        [ECHO_INITIAL_KEY]: ECHO_INITIAL_VALUE,
        [ECHO_TRAILING_KEY]: ECHO_TRAILING_VALUE.toString(),
      },
    };

    const interceptor: ClientInterceptor = (
      service,
      method,
      options,
      request,
      response
    ) => {
      return [
        request,
        {
          receive(handler) {
            response.receive({
              onHeader(header) {
                expect(header.has(ECHO_INITIAL_KEY)).toBeTrue();
                expect(header.get(ECHO_INITIAL_KEY)).toEqual(
                  ECHO_INITIAL_VALUE
                );
                handler.onHeader?.(header);
              },
              onMessage(message) {
                handler.onMessage(message);
              },
              onTrailer(trailer) {
                expect(trailer.has(ECHO_TRAILING_KEY)).toBeTrue();
                expect(trailer.get(ECHO_TRAILING_KEY)).toEqual(
                  ECHO_TRAILING_VALUE.toString()
                );
                handler.onTrailer?.(trailer);
              },
              onClose(error) {
                handler.onClose(error);
              },
            });
          },
        },
      ];
    };
    const transportWithInterceptor = createConnectTransport({
      baseUrl: `https://${host}:${port}`,
      interceptors: [interceptor],
    });
    const clientWithInterceptor = makePromiseClient(
      TestService,
      transportWithInterceptor
    );
    const response = await clientWithInterceptor.unaryCall(req, metadata);
    expect(response.payload).toBeDefined();
    expect(response.payload?.body.length).toEqual(size);
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
      expect(e.code).toEqual(StatusCode.Unknown);
      expect(e.rawMessage).toEqual(TEST_STATUS_MESSAGE);
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
      expect(e.code).toEqual(StatusCode.Unknown);
      expect(e.rawMessage).toEqual(TEST_STATUS_MESSAGE);
    }
  });
  // TODO: enable this test when we have a fix on connect-go
  xit("timeout_on_sleeping_server", async function () {
    const request = new StreamingOutputCallRequest({
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
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
        [StatusCode.Unknown, StatusCode.DeadlineExceeded].includes(e.code)
      ).toBeTrue();
    }
  });
  it("unimplemented_method", async function () {
    try {
      await client.unimplementedCall({});
      fail("expected to catch an error");
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect(e.code).toEqual(StatusCode.Unimplemented);
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
        [StatusCode.Unimplemented, StatusCode.NotFound].includes(e.code)
      ).toBeTrue();
    }
  });
  it("fail_unary", async function () {
    try {
      await client.failUnaryCall({});
    } catch (e) {
      expect(e).toBeInstanceOf(ConnectError);
      expect(e.code).toEqual(StatusCode.ResourceExhausted);
      expect(e.rawMessage).toEqual("soirée 🎉");
    }
  });
});
