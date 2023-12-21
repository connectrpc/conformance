// Copyright 2023 The Connect Authors
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
  Header as GoogleHeader,
  ConformancePayload as GoogleConformancePayload,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";
import {
  Header,
  ConformancePayload,
} from "./gen/proto/es/connectrpc/conformance/v1/service_pb.js";

export function convertBufHeaderToGoogleHeader(hdr: Header): GoogleHeader {
  const bin = hdr.toBinary();

  const dest = GoogleHeader.deserializeBinary(bin);
  dest.setName(hdr.name);
  dest.setValueList(hdr.value);
  return dest;
}

export function convertGooglePayloadToBufPayload(
  src: GoogleConformancePayload | undefined,
): ConformancePayload {
  if (src === undefined) {
    return new ConformancePayload();
  }
  const bin = src.serializeBinary();
  return ConformancePayload.fromBinary(bin);
}
