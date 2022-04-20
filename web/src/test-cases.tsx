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

import TestCase from "./test-case";
import {
  ConnectError,
  createConnectTransport,
  makePromiseClient,
  StatusCode,
} from "@bufbuild/connect-web";
import { TestService } from "../gen/proto/connect-web/grpc/testing/test_connectweb";
import { Empty } from "../gen/proto/connect-web/grpc/testing/empty_pb";
import {
  EchoStatus,
  Payload,
  SimpleRequest,
  SimpleResponse,
} from "../gen/proto/connect-web/grpc/testing/messages_pb";
import * as React from "react";

interface TestCasesProps {
  host: string;
  port: string;
}

const TestCases: React.FC<TestCasesProps> = (props: TestCasesProps) => {
  const { host, port } = props;
  const transport = createConnectTransport({
    baseUrl: `http://${host}:${port}`,
  });
  const client = makePromiseClient(TestService, transport);
  return (
    <table>
      <TestCase
        name="empty_unary"
        testFunc={async () => {
          const response = await client.emptyCall({});
          if (!(response instanceof Empty)) throw "response is not an Empty";
          return "success";
        }}
      />
      <TestCase
        name="empty_unary_with_timeout"
        testFunc={async () => {
          const deadlineMs = 1000; // 1 second
          const response = await client.emptyCall({}, { timeout: deadlineMs });
          if (!(response instanceof Empty)) throw "response is not an Empty";
          return "success";
        }}
      />
      <TestCase
        name="large_unary"
        testFunc={async () => {
          const req = new SimpleRequest();
          const size = 314159;

          req.responseSize = size;
          req.payload = new Payload();
          req.payload.body = new Uint8Array(271828).fill(0);

          const response = await client.unaryCall(req);
          if (!(response instanceof SimpleResponse))
            throw "response is not an SimpleResponse";
          if (response.payload === undefined)
            throw "response payload is undefined";
          if (response.payload.body.length !== size)
            throw "response payload body length not match";

          return "success";
        }}
      />
      <TestCase
        name="server_stream"
        testFunc={async () => {
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
            if (response === undefined) {
              throw "response is undefined";
            }
            if (response.payload === undefined) {
              throw "response.payload is undefined";
            }
            if (response.payload.body.length !== sizes[responseCount]) {
              throw "response.payload.body is not the same size as requested";
            }
            responseCount++;
          }
          if (responseCount !== sizes.length) {
            throw "not enough response received";
          }
          return "success";
        }}
      />
      <TestCase
        name="custom_metadata"
        testFunc={async () => {
          const req = new SimpleRequest();
          const size = 314159;
          const ECHO_INITIAL_KEY = "x-grpc-test-echo-initial";
          const ECHO_INITIAL_VALUE = "test_initial_metadata_value";
          const ECHO_TRAILING_KEY = "x-grpc-test-echo-trailing-bin";
          const ECHO_TRAILING_VALUE = 0xababab;

          req.responseSize = size;
          req.payload = new Payload();
          req.payload.body = new Uint8Array(271828).fill(0);

          const metadata = {
            headers: {
              [ECHO_INITIAL_KEY]: ECHO_INITIAL_VALUE,
              [ECHO_TRAILING_KEY]: ECHO_TRAILING_VALUE.toString(),
            },
          };

          const call = await client.unaryCall(req, metadata);
          // TODO: assert the response header

          return "success";
        }}
      />
      <TestCase
        name="status_code_and_message"
        testFunc={async () => {
          const req = new SimpleRequest();
          const TEST_STATUS_MESSAGE = "test status message";

          req.responseStatus = new EchoStatus();
          req.responseStatus.code = 2;
          req.responseStatus.message = TEST_STATUS_MESSAGE;

          try {
            const response = await client.unaryCall(req);
            if (response instanceof SimpleResponse)
              throw "unexpected successful call";
          } catch (e) {
            if (!(e instanceof ConnectError)) throw e;
            if (e.code !== 2) throw "unexpected error code";
            if (e.message !== `[${StatusCode[e.code]}] ${TEST_STATUS_MESSAGE}`)
              throw "unexpected error message";
          }

          return "success";
        }}
      />
      <TestCase
        name="special_status"
        testFunc={async () => {
          const req = new SimpleRequest();
          const TEST_STATUS_MESSAGE = `\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n`;

          req.responseStatus = new EchoStatus();
          req.responseStatus.code = 2;
          req.responseStatus.message = TEST_STATUS_MESSAGE;

          try {
            const response = await client.unaryCall(req);
            if (response instanceof SimpleResponse)
              throw "unexpected successful call";
          } catch (e) {
            if (!(e instanceof ConnectError)) throw e;
            if (e.code !== 2) throw "unexpected error code";
            if (e.message !== `[${StatusCode[e.code]}] ${TEST_STATUS_MESSAGE}`)
              throw "unexpected error message";
          }

          return "success";
        }}
      />
      <TestCase
        name="unimplemented_method"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
      <TestCase
        name="unimplemented_service"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
    </table>
  );
};

export default TestCases;
