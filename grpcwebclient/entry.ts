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

import { ClientCompatRequest } from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
import invoke from "./impl.js";

// This file represents the entry point into the browser code being executed by
// Webdriver. The conformance-runner file builds all the browser code, using
// this file as its entry point.

declare global {
  // @ts-expect-error asd
  const arguments: [
    string, // The request binary data
    (
      res:
        | { type: "data"; data: Uint8Array }
        | { type: "error"; error: string },
    ) => void, // The done callback
  ];
}

// Read the arguments passed from the executeAsyncScript call
// The first argument is a binary string representing a ClientCompatRequest
// sent from the conformance runner binary. The second argument is a callback
// to be invoked with the ClientResponseResult returned from the gRPC-web client.
// These arguments are how the conformance runner code communicates with the
// code running in the Webdriver's headless browser shell.
const buffer = new Uint8Array(JSON.parse(arguments[0]));

const req = ClientCompatRequest.deserializeBinary(buffer);
const done = arguments[1];
void invoke(req).then(
  (result) => {
    done({
      type: "data",
      data: result.serializeBinary(),
    });
  },
  (err) => {
    done({ type: "error", error: `${err}` });
  },
);
