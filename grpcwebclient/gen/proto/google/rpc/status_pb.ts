// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @generated by protoc-gen-es v1.6.0 with parameter "target=ts"
// @generated from file google/rpc/status.proto (package google.rpc, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Any, Message, proto3 } from "@bufbuild/protobuf";

/**
 * The `Status` type defines a logical error model that is suitable for
 * different programming environments, including REST APIs and RPC APIs. It is
 * used by [gRPC](https://github.com/grpc). Each `Status` message contains
 * three pieces of data: error code, error message, and error details.
 *
 * You can find out more about this error model and how to work with it in the
 * [API Design Guide](https://cloud.google.com/apis/design/errors).
 *
 * @generated from message google.rpc.Status
 */
export class Status extends Message<Status> {
  /**
   * The status code, which should be an enum value of
   * [google.rpc.Code][google.rpc.Code].
   *
   * @generated from field: int32 code = 1;
   */
  code = 0;

  /**
   * A developer-facing error message, which should be in English. Any
   * user-facing error message should be localized and sent in the
   * [google.rpc.Status.details][google.rpc.Status.details] field, or localized
   * by the client.
   *
   * @generated from field: string message = 2;
   */
  message = "";

  /**
   * A list of messages that carry the error details.  There is a common set of
   * message types for APIs to use.
   *
   * @generated from field: repeated google.protobuf.Any details = 3;
   */
  details: Any[] = [];

  constructor(data?: PartialMessage<Status>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.rpc.Status";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "code", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 2, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "details", kind: "message", T: Any, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Status {
    return new Status().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Status {
    return new Status().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Status {
    return new Status().fromJsonString(jsonString, options);
  }

  static equals(a: Status | PlainMessage<Status> | undefined, b: Status | PlainMessage<Status> | undefined): boolean {
    return proto3.util.equals(Status, a, b);
  }
}

