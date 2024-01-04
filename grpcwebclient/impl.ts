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
  UnimplementedRequest,
  ServerStreamRequest,
} from "./gen/proto/es/connectrpc/conformance/v1/service_pb.js";
import {
  UnaryRequest as UnaryRequestGoog,
  UnaryResponse as UnaryResponseGoog,
  UnimplementedRequest as UnimplementedRequestGoog,
  ServerStreamRequest as ServerStreamRequestGoog,
  ServerStreamResponse as ServerStreamResponseGoog,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";
import { Status } from "@buf/googleapis_googleapis.bufbuild_es/google/rpc/status_pb.js";
import { Metadata, RpcError, Status as GrpcWebStatus } from "grpc-web";

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

function buildMetadata(req: ClientCompatRequest): Metadata {
  const metadata: Metadata = {};
  req.requestHeaders.forEach((hdr: Header) => {
    const s = hdr.value.join(",");
    metadata[hdr.name] = s;
  });

  if (req.timeoutMs !== undefined && req.timeoutMs > 0) {
    let deadline = new Date();
    deadline.setMilliseconds(deadline.getMilliseconds() + req.timeoutMs);
    metadata.deadline = deadline.getTime().toString();
  }
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

function getCancelTiming(req: ClientCompatRequest): number {
  if (req.cancel == undefined) {
    return -1;
  }
  const cancelTiming = req.cancel.cancelTiming;
  switch (cancelTiming.case) {
    case "afterNumResponses":
      return cancelTiming.value;
    default:
      // Since server streaming is the only streaming method supported for grpc-web,
      // we can just ignore all values other than afterNumResponses since they
      // apply to client and bidi
      return -1;
  }
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
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  const resp = new ClientResponseResult({
    responseHeaders: [],
    responseTrailers: [],
    payloads: [],
    error: undefined,
  });

  const metadata: Metadata = buildMetadata(req);
  const result = client.unary(
    ur,
    metadata,
    (err: RpcError, response: UnaryResponseGoog) => {
      if (err !== null) {
        resp.error = convertGrpcToProtoError(err);
      } else {
        const payload = response.getPayload();
        if (payload !== undefined) {
          resp.payloads.push(
            ConformancePayload.fromBinary(payload.serializeBinary()),
          );
        }
      }
    },
  );

  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  result.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.responseHeaders = convertMetadataToHeader(md);
    }
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  result.on("status", (status: GrpcWebStatus) => {
    const md = status.metadata;
    if (md !== undefined) {
      resp.responseTrailers = convertMetadataToHeader(md);
      res(resp);
    }
  });

  return prom;
}

async function serverStream(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.requestMessages[0];
  const uReq = new ServerStreamRequest();
  if (!msg.unpackTo(uReq)) {
    throw new Error(
      "Could not unpack request message to server stream request",
    );
  }

  // Convert from Protobuf-ES into the gRPC-web compatible library
  const ssr = ServerStreamRequestGoog.deserializeBinary(uReq.toBinary());

  const resp = new ClientResponseResult({
    responseHeaders: [],
    responseTrailers: [],
    payloads: [],
    error: undefined,
  });

  const cancelAfterNumResponses = getCancelTiming(req);
  const metadata: Metadata = buildMetadata(req);

  const stream = client.serverStream(ssr, metadata);

  // If the cancel timing specifies after 0 responses, then cancel before
  // receiving anything
  if (cancelAfterNumResponses === 0) {
    stream.cancel();
    return new Promise<ClientResponseResult>((resolve) => {
      resolve(resp);
    });
  }

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  let totalRcvd = 0;
  stream.on("data", (response: ServerStreamResponseGoog) => {
    const payload = response.getPayload();
    if (payload !== undefined) {
      resp.payloads.push(
        ConformancePayload.fromBinary(payload.serializeBinary()),
      );
      totalRcvd += 1;
      // If afterNumResponses is specified, it will be a number > 0 here.
      // If it wasn't specified, it will be -1, which means the totalRcvd
      // will never be equal and we won't cancel.
      if (totalRcvd === cancelAfterNumResponses) {
        stream.cancel();
        res(resp);
      }
    }
  });
  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  stream.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.responseHeaders = convertMetadataToHeader(md);
    }
    res(resp);
  });
  stream.on("error", (err: RpcError) => {
    resp.error = convertGrpcToProtoError(err);
    res(resp);
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  stream.on("status", (status: GrpcWebStatus) => {
    const md = status.metadata;
    if (md !== undefined) {
      resp.responseTrailers = convertMetadataToHeader(md);
    }
    res(resp);
  });

  stream.on("end", function () {
    res(resp);
  });

  return prom;
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
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.requestMessages[0];
  const uReq = new UnimplementedRequest();
  if (!msg.unpackTo(uReq)) {
    throw new Error(
      "Could not unpack request message to unimplemented request",
    );
  }
  // Convert from Protobuf-ES into the gRPC-web compatible library
  const ur = UnimplementedRequestGoog.deserializeBinary(uReq.toBinary());

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  const metadata: Metadata = buildMetadata(req);
  client.unimplemented(ur, metadata, (err) => {
    res(
      new ClientResponseResult({
        error: convertGrpcToProtoError(err),
      }),
    );
  });

  return prom;
}

function createClient(req: ClientCompatRequest) {
  // let scheme = "http://";
  // if (req.serverTlsCert.length > 0) {
  //   scheme = "https://";
  // }
  // const baseUrl = `${scheme}${req.host}:${req.port}`;
  const baseUrl = "http://127.0.0.1:23457";
  return new ConformanceServiceClient(baseUrl);
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
