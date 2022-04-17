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
import * as React from "react";

interface TestCasesProps {
  host: string;
  port: string;
}

const TestCases: React.FC<TestCasesProps> = (props: TestCasesProps) => {
  const { host, port } = props;
  // TODO(doria): this should probably be a hook right?
  const transport = createConnectTransport({
    baseUrl: `http://${host}:${port}`,
  });
  const client = makePromiseClient(TestService, transport);
  return (
    <table>
      <TestCase
        name="empty_unary"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="empty_unary_with_deadline"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="large_unary"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="server_stream"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="custom_metadata"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="status_code_and_message"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="special_status"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="unimplemented_method"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="unimplemented_service"
        testFunc={() =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
    </table>
  );
};

export default TestCases;
