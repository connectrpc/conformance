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
            if (response instanceof Empty) throw "err is not null"
        }
        }
      />
      <TestCase
        name="empty_unary_with_deadline"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="large_unary"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="server_stream"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="custom_metadata"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="status_code_and_message"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="special_status"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="unimplemented_method"
        testFunc={async () =>
          void (
            {
              // TODO: fill in test case using `client`
            }
          )
        }
      />
      <TestCase
        name="unimplemented_service"
        testFunc={async () =>
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
