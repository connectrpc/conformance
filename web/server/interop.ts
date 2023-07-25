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
  ErrorDetail,
  Payload,
  PayloadType,
} from "../gen/proto/connect-web/connect/conformance/messages_pb.js";
import { cors as connectCors } from "@bufbuild/connect";

export const interop = {
  /**
   * readable non-ASCII
   */
  nonASCIIErrMsg: "soirée 🎉",

  /**
   * An error detail to be included in an error.
   */
  errorDetail: new ErrorDetail({
    reason: "soirée 🎉",
    domain: "connect-crosstest",
  }),

  leadingMetadataKey: "x-grpc-test-echo-initial",
  trailingMetadataKey: "x-grpc-test-echo-trailing-bin",

  makeServerPayload(payloadType: PayloadType, size: number): Payload {
    switch (payloadType) {
      case PayloadType.COMPRESSABLE:
        return new Payload({
          body: new Uint8Array(size),
          type: PayloadType.COMPRESSABLE,
        });
      default:
        // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
        throw new Error(`unsupported payload type: ${payloadType}`);
    }
  },

  corsOptions: {
    // Reflects the request origin. This should only be used for development.
    // Production should explicitly specify an origin
    origin: true,
    methods: [...connectCors.allowedMethods],
    allowedHeaders: [
      ...connectCors.allowedHeaders,
      "X-Grpc-Test-Echo-Initial",
      "X-Grpc-Test-Echo-Trailing-Bin",
      "Request-Protocol",
      "Get-Request",
    ],
    exposedHeaders: [
      ...connectCors.exposedHeaders,
      "X-Grpc-Test-Echo-Initial",
      "X-Grpc-Test-Echo-Trailing-Bin",
      "Trailer-X-Grpc-Test-Echo-Trailing-Bin", // unary trailer in Connect
      "Request-Protocol",
      "Get-Request",
    ],
  },
};
