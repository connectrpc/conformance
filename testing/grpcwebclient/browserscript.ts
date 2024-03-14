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
  ClientCompatResponse,
  ClientErrorResult,
  ClientResponseResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
import { Code } from "./gen/proto/connectrpc/conformance/v1/config_pb.js";
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
import {
  Metadata,
  RpcError,
  StatusCode,
  Status as GrpcWebStatus,
} from "grpc-web";

// The main entry point into the browser code running in Puppeteer/headless Chrome.
// This function is invoked by the page.evalulate call in grpcwebclient.
async function runTestCase(data: number[]): Promise<number[]> {
  const req = ClientCompatRequest.deserializeBinary(new Uint8Array(data));

  logTestNameToPage(req.getTestName())
  let res = new ClientCompatResponse();
  res.setTestName(req.getTestName())
  try {
    const result = await invoke(req);
    res.setResponse(result);
  } catch (e) {
    const err = new ClientErrorResult();
    err.setMessage((e as Error).message);
    res.setError(err);
  }

  return Array.from(res.serializeBinary());
}

function initPage() {
  let body = document.getElementsByTagName("body");
  const header = document.createElement("h1");
  header.id = "title";
  header.innerText = "Running tests...";
  body[0].appendChild(header);
}

function testsComplete() {
  const header = document.getElementById("title");
  if (header) {
    header.innerText = "Tests complete!";
  }
}

function logTestNameToPage(name: string) {
  const body = document.getElementsByTagName("body");
  const text = document.createElement("b")
  text.innerText = name + ":";
  body[0].appendChild(text);
  body[0].appendChild(document.createElement("br"));
}

function logToPage(message: string) {
  const body = document.getElementsByTagName("body");
  const text = document.createElement("span")
  text.innerText = message;
  body[0].appendChild(text);
  body[0].appendChild(document.createElement("br"));
}

function invoke(req: ClientCompatRequest) {
  const client = createClient(req);
  switch (req.getMethod()) {
    case "Unary":
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

function convertStatusCodeToCode(code: StatusCode): Code {
  switch (code) {
    case StatusCode.ABORTED:
      return Code.CODE_ABORTED;
    case StatusCode.ALREADY_EXISTS:
      return Code.CODE_ALREADY_EXISTS;
    case StatusCode.CANCELLED:
      return Code.CODE_CANCELED;
    case StatusCode.DATA_LOSS:
      return Code.CODE_DATA_LOSS;
    case StatusCode.DEADLINE_EXCEEDED:
      return Code.CODE_DEADLINE_EXCEEDED;
    case StatusCode.FAILED_PRECONDITION:
      return Code.CODE_FAILED_PRECONDITION;
    case StatusCode.INTERNAL:
      return Code.CODE_INTERNAL;
    case StatusCode.INVALID_ARGUMENT:
      return Code.CODE_INVALID_ARGUMENT;
    case StatusCode.NOT_FOUND:
      return Code.CODE_NOT_FOUND;
    case StatusCode.OUT_OF_RANGE:
      return Code.CODE_OUT_OF_RANGE;
    case StatusCode.PERMISSION_DENIED:
      return Code.CODE_PERMISSION_DENIED;
    case StatusCode.RESOURCE_EXHAUSTED:
      return Code.CODE_RESOURCE_EXHAUSTED;
    case StatusCode.UNAUTHENTICATED:
      return Code.CODE_UNAUTHENTICATED;
    case StatusCode.UNAVAILABLE:
      return Code.CODE_UNAVAILABLE;
    case StatusCode.UNIMPLEMENTED:
      return Code.CODE_UNIMPLEMENTED;
    default:
      return Code.CODE_UNKNOWN;
  }
}

function convertGrpcToProtoError(rpcErr: RpcError): ProtoError {
  const err = new ProtoError();
  err.setCode(convertStatusCodeToCode(rpcErr.code));
  err.setMessage(rpcErr.message);

  let value : string | undefined;
  const md = rpcErr.metadata;
  if (md !== undefined) {
    value = md["grpc-status-details-bin"];
  }
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
  if (req.getRequestMessagesList().length !== 1) {
    throw new Error("Unary method requires exactly one request message");
  }
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
  let isResolved = false;
  prom.then(() => { isResolved = true; });

  const resp = new ClientResponseResult();
  resp.setResponseHeadersList([]);
  resp.setResponseTrailersList([]);
  resp.setPayloadsList([]);
  resp.setError(undefined);

  const metadata: Metadata = buildMetadata(req);
  logToPage("Sending request");
  const result = client.unary(
    uReq,
    metadata,
    (err: RpcError, response: UnaryResponse) => {
      if (err !== null) {
        if (!isResolved) {
          logToPage("RPC error received");
          resp.setError(convertGrpcToProtoError(err));
          const md = err.metadata;
          if (md !== undefined) {
            resp.setResponseTrailersList(convertMetadataToHeader(md));
          }
          res(resp)
        }
      } else {
        logToPage("Message received");
        const payload = response.getPayload();
        if (payload !== undefined) {
          resp.addPayloads(payload);
        }
      }
    },
  );

  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  result.on("metadata", (md: Metadata) => {
    logToPage("Headers received");
    if (md !== undefined) {
      resp.setResponseHeadersList(convertMetadataToHeader(md));
    }
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  result.on("status", (status: GrpcWebStatus) => {
    if (!isResolved) {
      logToPage("Status/trailers received");
      const md = status.metadata;
      if (md !== undefined) {
        resp.setResponseTrailersList(convertMetadataToHeader(md));
      }
      res(resp)
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

  logToPage("Sending request");
  const stream = client.serverStream(uReq, metadata);

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });
  let isResolved = false;
  prom.then(() => { isResolved = true; });

  stream.on("data", (response: ServerStreamResponse) => {
    logToPage("Message received");
    const payload = response.getPayload();
    if (payload !== undefined) {
      resp.addPayloads(payload);
    }
  });
  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  stream.on("metadata", (md: Metadata) => {
    logToPage("Headers received");
    if (md !== undefined) {
      resp.setResponseHeadersList(convertMetadataToHeader(md));
    }
  });
  stream.on("error", (err: RpcError) => {
    if (!isResolved) {
      logToPage("RPC error received");
      resp.setError(convertGrpcToProtoError(err));
    }
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  stream.on("status", (status: GrpcWebStatus) => {
    if (!isResolved) {
      logToPage("Status/trailers received");
      const md = status.metadata;
      if (md !== undefined) {
        resp.setResponseTrailersList(convertMetadataToHeader(md));
      }
      res(resp);
    }
  });

  stream.on("end", function () {
    if (!isResolved) {
      logToPage("RPC ended");
      res(resp);
    }
  });

  return prom;
}

async function clientStream(): Promise<ClientResponseResult> {
  throw new Error("Client Streaming is not supported in gRPC-web");
}

async function bidiStream(): Promise<ClientResponseResult> {
  throw new Error("Client Streaming is not supported in gRPC-web");
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
  logToPage("Sending request")
  client.unimplemented(uReq, metadata, (err) => {
    logToPage("RPC error received") // what if we get back a message and not an error? ¯\_(ツ)_/¯
    const result = new ClientResponseResult();
    result.setError(convertGrpcToProtoError(err));
    res(result);
  });

  return prom;
}

// @ts-ignore
window.runTestCase = runTestCase;
// @ts-ignore
window.initPage = initPage;
// @ts-ignore
window.testsComplete = testsComplete;
