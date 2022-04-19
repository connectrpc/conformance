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
  createConnectTransport,
  makePromiseClient,
} from "@bufbuild/connect-web";
import { TestService } from "../gen/proto/connect-web/grpc/testing/test_connectweb";
import { Empty } from "../gen/proto/connect-web/grpc/testing/empty_pb";
// import { ResponseParameters } from "../gen/proto/connect-web/grpc/testing/messages_pb";
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
        name="empty_unary_with_deadline"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
      <TestCase
        name="large_unary"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
      <TestCase
        name="server_stream"
        testFunc={async () => {
          const sizes = [31415, 9, 2653, 58979];
          const responseParams = sizes.map((size, index) => {
            return {
              size: size,
              intervalUs: index * 10,
            }
          });
          let responseCount = 0;
          for await (const response of await client.streamingOutputCall({
            responseParameters: responseParams,
          })) {
            if (response === undefined) {
              throw "response is undefined"
            }
            if (response.payload === undefined) {
              throw "response.payload is undefined"
            }
            if (response.payload.body.length !== sizes[responseCount]) {
              throw "response.payload.body is not the same size as requested"
            }
            responseCount++
          }
          if (responseCount !== sizes.length) {
            throw "not enough response received"
          }
          return "success";
        }}
      />
      <TestCase
        name="custom_metadata"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
      <TestCase
        name="status_code_and_message"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
      />
      <TestCase
        name="special_status"
        // TODO: fill in test case using `client`
        testFunc={async () => "success"}
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
