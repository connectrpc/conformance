// Copyright 2023-2024 The Connect Authors
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

import { createRegistry } from "@bufbuild/protobuf";
import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
} from "./gen/proto/es/connectrpc/conformance/v1/client_compat_pb.js";
import invoke from "./impl.js";
import {
  ServerStreamRequest,
  UnaryRequest,
  ClientStreamRequest,
  BidiStreamRequest,
  ConformancePayload_RequestInfo,
  UnimplementedRequest,
} from "./gen/proto/es/connectrpc/conformance/v1/service_pb.js";

declare global {
  // @ts-expect-error asd
  // eslint-disable-next-line no-shadow-restricted-names
  const arguments: [
    string, // The request JSON
    (
      res: { type: "data"; data: string } | { type: "error"; error: string },
    ) => void, // The done callback
  ];
}

const typeRegistry = createRegistry(
  UnaryRequest,
  ServerStreamRequest,
  ClientStreamRequest,
  BidiStreamRequest,
  ConformancePayload_RequestInfo,
  UnimplementedRequest,
  ClientCompatResponse,
  ClientResponseResult,
);

const req = ClientCompatRequest.fromJsonString(arguments[0], {
  typeRegistry,
});
const done = arguments[1];
void invoke(req).then(
  (result) => {
    done({
      type: "data",
      data: result.toJsonString({ typeRegistry }),
    });
  },
  (err) => {
    done({ type: "error", error: `${err}` });
  },
);
