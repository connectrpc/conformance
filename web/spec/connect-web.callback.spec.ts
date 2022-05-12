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
  makeCallbackClient,
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

function multiDone(done: DoneFn, count: number) {
  return function () {
    count -= 1;
    if (count <= 0) {
      done();
    }
  };
}

describe("connect_web_callback_client", function () {
  const host = __karma__.config.host;
  const port = __karma__.config.port;
  const transport = createConnectTransport({
    baseUrl: `https://${host}:${port}`,
  });
  const client = makeCallbackClient(TestService, transport);
  it("empty_unary", function (done) {
    client.emptyCall({}, (err, response) => {
      expect(err).toBeUndefined();
      expect(response).toEqual(new Empty());
      done();
    });
  });
  it("empty_unary_with_timeout", function (done) {
    const deadlineMs = 1000; // 1 second
    client.emptyCall(
      {},
      (err, response) => {
        expect(err).toBeUndefined();
        expect(response).toEqual(new Empty());
        done();
      },
      { timeout: deadlineMs }
    );
  });
  it("large_unary", function (done) {
    const size = 314159;
    const req = new SimpleRequest({
      responseSize: size,
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
    });
    client.unaryCall(req, (err, response) => {
      expect(err).toBeUndefined();
      expect(response.payload).toBeDefined();
      expect(response.payload?.body.length).toEqual(size);
      done();
    });
  });
  it("server_streaming", function (done) {
    const sizes = [31415, 9, 2653, 58979];
    const doneFn = multiDone(done, sizes.length);
    const responseParams = sizes.map((size, index) => {
      return {
        size: size,
        intervalUs: index * 10,
      };
    });
    let responseCount = 0;
    client.streamingOutputCall(
      {
        responseParameters: responseParams,
      },
      (response) => {
        expect(response.payload).toBeDefined();
        expect(response.payload?.body.length).toEqual(sizes[responseCount]);
        responseCount++;
        doneFn();
      },
      (err) => {
        expect(err).toBeUndefined();
      }
    );
  });
  // TODO: enable this test when we have a fix on connect-web
  xit("empty_stream", function (done) {
    client.streamingOutputCall(
      {
        responseParameters: [],
      },
      () => {
        fail("expecting no response in the empty stream");
      },
      (err) => {
        expect(err).toBeUndefined();
        done();
      }
    );
  });
  it("custom_metadata", function (done) {
    const doneFn = multiDone(done, 3);
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
                doneFn();
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
                doneFn();
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
    const clientWithInterceptor = makeCallbackClient(
      TestService,
      transportWithInterceptor
    );
    clientWithInterceptor.unaryCall(
      req,
      (err, response) => {
        expect(err).toBeUndefined();
        expect(response.payload).toBeDefined();
        expect(response.payload?.body.length).toEqual(size);
        doneFn();
      },
      metadata
    );
  });
  it("status_code_and_message", function (done) {
    const TEST_STATUS_MESSAGE = "test status message";
    const req = new SimpleRequest({
      responseStatus: {
        code: StatusCode.Unknown,
        message: TEST_STATUS_MESSAGE,
      },
    });
    client.unaryCall(req, (err) => {
      expect(err).toBeInstanceOf(ConnectError);
      expect(err.code).toEqual(StatusCode.Unknown);
      expect(err.rawMessage).toEqual(TEST_STATUS_MESSAGE);
      done();
    });
  });
  it("special_status", function (done) {
    const TEST_STATUS_MESSAGE = `\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n`;
    const req = new SimpleRequest({
      responseStatus: {
        code: StatusCode.Unknown,
        message: TEST_STATUS_MESSAGE,
      },
    });
    client.unaryCall(req, (err) => {
      expect(err).toBeInstanceOf(ConnectError);
      expect(err.code).toEqual(StatusCode.Unknown);
      expect(err.rawMessage).toEqual(TEST_STATUS_MESSAGE);
      done();
    });
  });
  // TODO: enable this test when we have a fix on connect-go
  xit("timeout_on_sleeping_server", function (done) {
    const request = new StreamingOutputCallRequest({
      payload: {
        body: new Uint8Array(271828).fill(0),
      },
    });
    client.streamingOutputCall(
      request,
      (response) => {
        fail(`expecting no response from sleeping server, got: ${response}`);
      },
      (err) => {
        expect(err).toBeDefined();
        expect(err).toBeInstanceOf(ConnectError);
        // We expect this to be DEADLINE_EXCEEDED, however envoy is monitoring the stream timeout
        // and will return an HTTP status code 408 when stream max duration time reached, which
        // cannot be translated to a connect error code, so connect-web client throws an Unknown.
        expect(
          [StatusCode.Unknown, StatusCode.DeadlineExceeded].includes(err.code)
        ).toBeTrue();
        done();
      },
      {
        timeout: 1, // 1ms
      }
    );
  });
  it("unimplemented_method", function (done) {
    client.unimplementedCall({}, (err) => {
      expect(err).toBeInstanceOf(ConnectError);
      expect(err.code).toEqual(StatusCode.Unimplemented);
      done();
    });
  });
  it("unimplemented_service", function (done) {
    const badClient = makeCallbackClient(UnimplementedService, transport);
    badClient.unimplementedCall({}, (err) => {
      expect(err).toBeInstanceOf(ConnectError);
      // We expect this to be either Unimplemented or NotFound, depending on the implementation.
      // In order to support a consistent behaviour for this case, the backend would need to
      // own the router and all fallback behaviours. Both statuses are valid returns for this
      // case and the client should not retry on either status.
      expect(
        [StatusCode.Unimplemented, StatusCode.NotFound].includes(err.code)
      ).toBeTrue();
      done();
    });
  });
  it("fail_unary", function (done) {
    client.failUnaryCall({}, (err) => {
      expect(err).toBeInstanceOf(ConnectError);
      expect(err.code).toEqual(StatusCode.ResourceExhausted);
      expect(err.rawMessage).toEqual("soirée 🎉");
      done();
    });
  });
});
