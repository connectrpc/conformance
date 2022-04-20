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
          assert(response.equals({}), `unexpected response: ${response}`);

          return "success";
        }}
      />
      <TestCase
        name="empty_unary_with_timeout"
        testFunc={async () => {
          const deadlineMs = 1000; // 1 second
          const response = await client.emptyCall({}, { timeout: deadlineMs });
          assert(response.equals({}), `unexpected response: ${response}`);

          return "success";
        }}
      />
      <TestCase
        name="large_unary"
        testFunc={async () => {
          const size = 314159;
          const req = new SimpleRequest({
            responseSize: size,
            payload: {
              body: new Uint8Array(271828).fill(0),
            },
          });

          const response = await client.unaryCall(req);
          assert(
            response.payload !== undefined,
            "response payload is undefined"
          );
          assert(
            response.payload.body.length === size,
            "response payload body length not match"
          );

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
            assert(response !== undefined, "response is undefined");
            assert(
              response.payload !== undefined,
              "response.payload is undefined"
            );
            assert(
              response.payload.body.length === sizes[responseCount],
              "response.payload.body is not the same size as requested"
            );
            responseCount++;
          }
          assert(
            responseCount === sizes.length,
            "not enough response received"
          );
          return "success";
        }}
      />
      <TestCase
        name="custom_metadata"
        testFunc={async () => {
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

          const call = await client.unaryCall(req, metadata);
          // TODO: assert the response header

          return "success";
        }}
      />
      <TestCase
        name="status_code_and_message"
        testFunc={async () => {
          const TEST_STATUS_MESSAGE = "test status message";
          const req = new SimpleRequest({
            responseStatus: {
              code: StatusCode.Unknown,
              message: TEST_STATUS_MESSAGE,
            },
          });

          try {
            const response = await client.unaryCall(req);
            throw "unexpected successful call";
          } catch (e) {
            assert(
              e instanceof ConnectError,
              `error is not a ConnectError: ${e}`
            );
            assert(
              e.code === StatusCode.Unknown,
              `unexpected error code: ${e.code}`
            );
            assert(
              e.message === `[${StatusCode[e.code]}] ${TEST_STATUS_MESSAGE}`,
              `unexpected error message: ${e.message}`
            );
          }

          return "success";
        }}
      />
      <TestCase
        name="special_status"
        testFunc={async () => {
          const TEST_STATUS_MESSAGE = `\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n`;
          const req = new SimpleRequest({
            responseStatus: {
              code: StatusCode.Unknown,
              message: TEST_STATUS_MESSAGE,
            },
          });

          try {
            const response = await client.unaryCall(req);
            throw "unexpected successful call";
          } catch (e) {
            assert(
              e instanceof ConnectError,
              `error is not a ConnectError: ${e}`
            );
            assert(
              e.code === StatusCode.Unknown,
              `unexpected error code: ${e.code}`
            );
            assert(
              e.message === `[${StatusCode[e.code]}] ${TEST_STATUS_MESSAGE}`,
              `unexpected error message: ${e.message}`
            );
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

/**
 * Assert that condition is truthy or throw error (with message)
 */
export function assert(condition: unknown, msg?: string): asserts condition {
  // eslint-disable-next-line @typescript-eslint/strict-boolean-expressions -- we want the implicit conversion to boolean
  if (!condition) {
    throw new Error(msg);
  }
}

export default TestCases;
