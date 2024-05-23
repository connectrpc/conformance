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

import {ConformanceServiceClient} from "./gen/proto/connectrpc/conformance/v1/ServiceServiceClientPb.js";
import {ClientCompatRequest, ClientResponseResult,} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
import {Code} from "./gen/proto/connectrpc/conformance/v1/config_pb.js";
import {
  Error as ProtoError,
  Header,
  ServerStreamRequest,
  ServerStreamResponse,
  UnaryRequest,
  UnaryResponse,
  UnimplementedRequest,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";
import {Status} from "@buf/googleapis_googleapis.protocolbuffers_js/google/rpc/status_pb.js";
import {Metadata, RpcError, Status as GrpcWebStatus, StatusCode,} from "grpc-web";
import CancelTimingCase = ClientCompatRequest.Cancel.CancelTimingCase;

// The main entry point into the browser code running in Puppeteer/headless Chrome.
// This function is invoked by the page.evaluate call in grpcwebclient.
async function runTestCase(data: number[]): Promise<number[]> {
  const request = ClientCompatRequest.deserializeBinary(new Uint8Array(data));

  const rpcResult = invoke(request);
  const timeout = new Promise<ClientResponseResult>((_, reject) => {
    // Fail if we still don't have a response after 15 seconds
    // so that user can at least see exactly which test case timed out.
    setTimeout(() => {
      reject(new Error("promise never resolved after 15s!"));
    }, 15*1000);
  });
  const result = await Promise.race([rpcResult, timeout]);

  return Array.from(result.serializeBinary());
}

function addErrorListeners() {
  window.addEventListener("error", function (e) {
    // @ts-ignore
    window.log("ERROR: uncaught error in browser: " + e.error.filename + ":" + e.error.lineno + ": " + e.message);
    return false;
  })
  window.addEventListener("unhandledrejection", function (e) {
    // @ts-ignore
    window.log("ERROR: unhandled promise failure in browser: " + e.reason);
    return false;
  })
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

  let md = rpcErr.metadata;
  if (md !== undefined) {
    const value = md["grpc-status-details-bin"];
    if (value) {
      const status = Status.deserializeBinary(stringToUint8Array(atob(value)));
      err.setDetailsList(status.getDetailsList());
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
        let md = err.metadata;
        if (md !== undefined) {
          resp.setResponseTrailersList(convertMetadataToHeader(md));
        }
        // Ideally, we'd complete the promise from the "end" event. However,
        // most RPCs that result in an RPC error (as of 3/15/2024, 50 out of
        // 57 failed RPCs) do not produce an "end" event after the callback
        // is invoked with an error.
        res(resp);
      } else {
        const payload = response.getPayload();
        if (payload !== undefined) {
          resp.addPayloads(payload);
        }
      }
    },
  );

  if (req.getCancel()?.getCancelTimingCase() == CancelTimingCase.AFTER_CLOSE_SEND_MS) {
    setTimeout(() => {
      result.cancel()
      // Annoyingly, when we cancel the RPC, no other callbacks or events get triggered.
      // So we have to complete the promise here.
      let err = new ProtoError()
      err.setCode(Code.CODE_CANCELED)
      resp.setError(err)
      res(resp)
    }, req.getCancel()?.getAfterCloseSendMs())
  }

  // Response headers (i.e. initial metadata) are sent in the 'metadata' event
  result.on("metadata", (md: Metadata) => {
    if (md !== undefined) {
      resp.setResponseHeadersList(convertMetadataToHeader(md));
    }
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  result.on("status", (status: GrpcWebStatus) => {
    // One might expect that the "status" event is always delivered (since
    // consistency would make it much easier to implement interceptors or
    // decorators, to instrument all RPCs with cross-cutting concerns, like
    // metrics, logging, etc). But one would be wrong: as of 3/15/2024, there
    // are 2 cases where the "status" event is never delivered (both cases
    // are RPC failures).
    const md = status.metadata;
    if (md !== undefined) {
      resp.setResponseTrailersList(convertMetadataToHeader(md));
    }
  });

  result.on("end", () => {
    res(resp);
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
  let doCancel = () => {
    stream.cancel()
    // Annoyingly, when we cancel the RPC, no other callbacks or events get triggered.
    // So we have to complete the promise here.
    let err = new ProtoError()
    err.setCode(Code.CODE_CANCELED)
    resp.setError(err)
    res(resp)
  }

  let res: (result: ClientResponseResult) => void;
  const prom = new Promise<ClientResponseResult>((resolve) => {
    res = resolve;
  });
  if (req.getCancel()?.getCancelTimingCase() == CancelTimingCase.AFTER_CLOSE_SEND_MS) {
    setTimeout(doCancel, req.getCancel()?.getAfterCloseSendMs())
  }
  let numResps = 0;
  stream.on("data", (response: ServerStreamResponse) => {
    const payload = response.getPayload();
    if (payload !== undefined) {
      resp.addPayloads(payload);
    }
    numResps++;
    if (req.getCancel()?.getCancelTimingCase() == CancelTimingCase.AFTER_NUM_RESPONSES &&
        numResps === req.getCancel()?.getAfterNumResponses()) {
      doCancel()
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
    let md = err.metadata;
    if (md !== undefined) {
      resp.setResponseTrailersList(convertMetadataToHeader(md));
    }
    // Ideally, we'd complete the promise from the "end" event. However, there
    // are some RPCs that result in an RPC error (as of 3/15/2024, 3 out of 44
    // failed RPCs) that do not produce an "end" event after the "error" event.
    res(resp);
  });

  // Response trailers (i.e. trailing metadata) are sent in the 'status' event
  stream.on("status", (status: GrpcWebStatus) => {
    // One might expect that the "status" event is always delivered (since
    // consistency would make it much easier to implement interceptors or
    // decorators, to instrument all RPCs with cross-cutting concerns, like
    // metrics, logging, etc). But one would be wrong: as of 3/15/2024, there
    // is one case (out of 62 total RPCs) where the "status" event is never
    // delivered for a streaming call.
    const md = status.metadata;
    if (md !== undefined) {
      resp.setResponseTrailersList(convertMetadataToHeader(md));
    }
  });

  stream.on("end", () => {
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
// @ts-ignore
window.addErrorListeners = addErrorListeners;
