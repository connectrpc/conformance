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

import { ConformanceServiceClient } from "./gen/proto/connectrpc/conformance/v1/ServiceServiceClientPb.js";
import {
  ClientCompatRequest,
  ClientResponseResult,
} from "./gen/proto/es/connectrpc/conformance/v1/client_compat_pb.js";
import {
  ConformancePayload,
  Error as ProtoError,
  Header,
  UnaryRequest,
} from "./gen/proto/es/connectrpc/conformance/v1/service_pb.js";
import {
  ConformancePayload as ConformancePayloadGoog,
  UnaryRequest as UnaryRequestGoog,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";
import { Status } from "./gen/proto/google/rpc/status_pb.js";
import { Metadata, RpcError } from "grpc-web";

function convertGrpcToProtoError(rpcErr: RpcError): ProtoError {
  const err = new ProtoError({
    code: rpcErr.code,
    message: rpcErr.message,
  });
  for (const [name, value] of Object.entries(rpcErr.metadata)) {
    if (name === "grpc-status-details-bin") {
      const status = Status.fromBinary(stringToUint8Array(atob(value)));

      err.details = status.details;
      break;
    }
  }
  return err;
}

function stringToUint8Array(str: string): Uint8Array {
  const buf = new ArrayBuffer(str.length);
  const bufView = new Uint8Array(buf);
  for (let i = 0; i < str.length; i++) {
    bufView[i] = str.charCodeAt(i);
  }
  return bufView;
}

function convertMetadataToDetails(md: Metadata): Header[] {
  const hdrs: Header[] = [];
  for (const [name, value] of Object.entries(md)) {
    hdrs.push(
      new Header({
        name,
        value: [value],
      }),
    );
  }

  return hdrs;
}

function deets(err: RpcError) {
  for (const [name, value] of Object.entries(err.metadata)) {
    if (name === "grpc-status-details-bin") {
      const s = atob(value);
      const status = Status.fromBinary(stringToUint8Array(s));

      return status.details;
    }
  }
}

export function convertGooglePayloadToProtoPayload(
  src: ConformancePayloadGoog | undefined,
): ConformancePayload {
  if (src === undefined) {
    return new ConformancePayload();
  }
  const bin = src.serializeBinary();
  return ConformancePayload.fromBinary(bin);
}

function convertHeadersToMetadata(hdrs: Header[]): Metadata {
  const metadata: Metadata = {};
  hdrs.forEach((hdr: Header) => {
    const s = hdr.value.join(",");
    metadata[hdr.name] = s;
  });
  return metadata;
}
function convertMetadataToHeader(md: Metadata): Header[] {
  const hdrs: Header[] = [];
  for (const [name, value] of Object.entries(md)) {
    hdrs.push(
      new Header({
        name,
        value: [value],
      }),
    );
  }

  return hdrs;
}

async function unary(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.requestMessages[0];
  const uReq = new UnaryRequest();
  if (!msg.unpackTo(uReq)) {
    throw new Error("Could not unpack request message to unary request");
  }

  // Convert from Protobuf-ES into the gRPC-web compatible library
  const ur = UnaryRequestGoog.deserializeBinary(uReq.toBinary());

  let res: (result: ClientResponseResult) => void;
  let rej: (reason: any) => void;
  const prom = new Promise<ClientResponseResult>((resolve, reject) => {
    res = resolve;
    rej = reject;
  });

  const resp = new ClientResponseResult({
    responseHeaders: [],
    responseTrailers: [],
    payloads: [],
    error: undefined,
  });

  const metadata: Metadata = convertHeadersToMetadata(req.requestHeaders);
  const result = client.unary(ur, metadata, (err, response) => {
    if (err !== null) {
      resp.error = convertGrpcToProtoError(err);
    } else {
      resp.payloads.push(
        convertGooglePayloadToProtoPayload(response.getPayload()),
      );
    }
  });

  result.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.responseHeaders = convertMetadataToHeader(md);
    }
  });
  result.on("status", (status) => {
    const md = status.metadata;
    if (md !== undefined) {
      resp.responseTrailers = convertMetadataToHeader(md);
      res(resp);
    }
  });

  return prom;
}

async function serverStream(
  _: ConformanceServiceClient,
  ccr: ClientCompatRequest,
): Promise<ClientResponseResult> {
  console.log(ccr);
  return new ClientResponseResult();
}

async function clientStream(): Promise<ClientResponseResult> {
  return new ClientResponseResult({
    error: {
      code: 12,
      message: "Client Streaming is not supported in gRPC-web",
    },
  });
}

async function bidiStream(): Promise<ClientResponseResult> {
  return new ClientResponseResult({
    error: {
      code: 12,
      message: "Bidi Streaming is not supported in gRPC-web",
    },
  });
}

async function unimplemented(
  _: ConformanceServiceClient,
  ccr: ClientCompatRequest,
): Promise<ClientResponseResult> {
  console.log(ccr);
  return new ClientResponseResult();
}

function createClient(req: ClientCompatRequest) {
  let scheme = "http://";
  if (req.serverTlsCert.length > 0) {
    scheme = "https://";
  }
  const baseUrl = `${scheme}${req.host}:${req.port}`;
  return new ConformanceServiceClient(baseUrl);
  // return new ConformanceServiceClient("http://127.0.0.1:23457");
}

export default (req: ClientCompatRequest) => {
  const client = createClient(req);
  switch (req.method) {
    case "Unary":
      if (req.requestMessages.length !== 1) {
        throw new Error("Unary method requires exactly one request message");
      }
      return unary(client, req);
    case "ServerStream":
      return serverStream(client, req);
    case "ClientStream":
      return clientStream();
    case "BidiStream":
      return bidiStream();
    case "Unimplemented":
      return unimplemented(client, req);
    default:
      throw new Error(`Unknown method: ${req.method}`);
  }
};