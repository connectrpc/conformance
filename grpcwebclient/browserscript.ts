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

import { ConformanceServiceClient } from "./gen/proto/connectrpc/conformance/v1/ServiceServiceClientPb.js";
import {
  ClientCompatRequest,
  ClientResponseResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
import {
  Error as ProtoError,
  Header,
  UnaryRequest,
  UnimplementedRequest,
  ServerStreamRequest,
  ServerStreamResponse,
  UnaryResponse,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";
import { Status } from "@buf/googleapis_googleapis.protocolbuffers_js/google/rpc/status_pb.js";
import { Metadata, RpcError, Status as GrpcWebStatus } from "grpc-web";

// The main entry point into the browser code running in Puppeteer/headless Chrome.
// This function is invoked by the page.evalulate call in grpcwebclient.
async function runTestCase(data: number[]): Promise<number[]> {
  const request = ClientCompatRequest.deserializeBinary(new Uint8Array(data));

  const result = await invoke(request);

  return Array.from(result.serializeBinary());
}

function invoke(req: ClientCompatRequest) {
  const client = createClient(req);
  switch (req.getMethod()) {
    case "Unary":
      if (req.getRequestMessagesList().length !== 1) {
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
      throw new Error(`Unknown method: ${req.getMethod()}`);
  }
}

function createClient(req: ClientCompatRequest) {
  let scheme = "http://";
  if (req.getServerTlsCert().length > 0) {
    scheme = "https://";
  }
  const baseUrl = `${scheme}${req.getHost()}:${req.getPort()}`;
  return new ConformanceServiceClient(baseUrl);
}

function convertGrpcToProtoError(rpcErr: RpcError): ProtoError {
  const err = new ProtoError();
  err.setCode(rpcErr.code);
  err.setMessage(rpcErr.message);

  const value = rpcErr.metadata["grpc-status-details-bin"];
  if (value) {
    const status = Status.deserializeBinary(stringToUint8Array(atob(value)));
    err.setDetailsList(status.getDetailsList());
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
  req.getRequestHeadersList().forEach((hdr: Header) => {
    const s = hdr.getValueList().join(",");
    metadata[hdr.getName()] = s;
  });

  if (req.getTimeoutMs() !== undefined && req.getTimeoutMs() > 0) {
    let deadline = new Date();
    deadline.setMilliseconds(deadline.getMilliseconds() + req.getTimeoutMs());
    metadata.deadline = deadline.getTime().toString();
  }

  return metadata;
}

function convertMetadataToHeader(md: Metadata): Header[] {
  const hdrs: Header[] = [];
  for (const [name, value] of Object.entries(md)) {
    const hdr = new Header();
    hdr.setName(name);
    hdr.setValueList([value]);
    hdrs.push(hdr);
  }

  return hdrs;
}

async function unary(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.getRequestMessagesList()[0];
  const uReq = msg.unpack(
    UnaryRequest.deserializeBinary,
    "connectrpc.conformance.v1.UnaryRequest",
  );
  if (!uReq) {
    throw new Error("Could not unpack request message to unary request");
  }

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  const resp = new ClientResponseResult();
  resp.setResponseHeadersList([]);
  resp.setResponseTrailersList([]);
  resp.setPayloadsList([]);
  resp.setError(undefined);

  const metadata: Metadata = buildMetadata(req);
  const result = client.unary(
    uReq,
    metadata,
    (err: RpcError, response: UnaryResponse) => {
      if (err !== null) {
        resp.setError(convertGrpcToProtoError(err));
      } else {
        const payload = response.getPayload();
        if (payload !== undefined) {
          resp.addPayloads(payload);
        }
      }
    },
  );

  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  result.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.setResponseHeadersList(convertMetadataToHeader(md));
    }
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  result.on("status", (status: GrpcWebStatus) => {
    const md = status.metadata;
    if (md !== undefined) {
      resp.setResponseTrailersList(convertMetadataToHeader(md));
      res(resp);
    }
  });

  return prom;
}

async function serverStream(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.getRequestMessagesList()[0];
  const uReq = msg.unpack(
    ServerStreamRequest.deserializeBinary,
    "connectrpc.conformance.v1.ServerStreamRequest",
  );
  if (!uReq) {
    throw new Error(
      "Could not unpack request message to server stream request",
    );
  }

  const resp = new ClientResponseResult();
  resp.setResponseHeadersList([]);
  resp.setResponseTrailersList([]);
  resp.setPayloadsList([]);
  resp.setError(undefined);

  const metadata: Metadata = buildMetadata(req);

  const stream = client.serverStream(uReq, metadata);

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  stream.on("data", (response: ServerStreamResponse) => {
    const payload = response.getPayload();
    if (payload !== undefined) {
      resp.addPayloads(payload);
    }
  });
  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  stream.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.setResponseHeadersList(convertMetadataToHeader(md));
    }
  });
  stream.on("error", (err: RpcError) => {
    resp.setError(convertGrpcToProtoError(err));
    res(resp);
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  stream.on("status", (status: GrpcWebStatus) => {
    const md = status.metadata;
    if (md !== undefined) {
      resp.setResponseTrailersList(convertMetadataToHeader(md));
    }
  });

  stream.on("end", function () {
    res(resp);
  });

  return prom;
}

async function clientStream(): Promise<ClientResponseResult> {
  const result = new ClientResponseResult();
  const err = new ProtoError();
  err.setCode(12);
  err.setMessage("Client Streaming is not supported in gRPC-web");
  result.setError(err);

  return result;
}

async function bidiStream(): Promise<ClientResponseResult> {
  const result = new ClientResponseResult();
  const err = new ProtoError();
  err.setCode(12);
  err.setMessage("Bidi Streaming is not supported in gRPC-web");
  result.setError(err);

  return result;
}

async function unimplemented(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.getRequestMessagesList()[0];
  const uReq = msg.unpack(
    UnimplementedRequest.deserializeBinary,
    "connectrpc.conformance.v1.UnimplementedRequest",
  );
  if (!uReq) {
    throw new Error(
      "Could not unpack request message to unimplemented request",
    );
  }

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });

  const metadata: Metadata = buildMetadata(req);
  client.unimplemented(uReq, metadata, (err) => {
    const result = new ClientResponseResult();
    result.setError(convertGrpcToProtoError(err));

    res(result);
  });

  return prom;
}

// @ts-ignore
window.runTestCase = runTestCase;
