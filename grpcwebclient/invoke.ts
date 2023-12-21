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
import { UnaryRequest } from "./gen/proto/es/connectrpc/conformance/v1/service_pb.js";
import {
  UnaryRequest as UR,
  UnaryResponseDefinition,
} from "./gen/proto/connectrpc/conformance/v1/service_pb.js";

async function unary(
  client: ConformanceServiceClient,
  req: ClientCompatRequest,
): Promise<ClientResponseResult> {
  const msg = req.requestMessages[0];
  const uReq = new UnaryRequest();
  if (!msg.unpackTo(uReq)) {
    throw new Error("Could not unpack request message to unary request");
  }

  const ur = new UR();
  const def = new UnaryResponseDefinition();
  def.setResponseData("goosh");
  ur.setResponseDefinition(def);

  console.log("GOOOOSH");
  console.log(ur);

  try {
    const result = await client.unary(ur);

    const resp = new ClientResponseResult();

    return resp;
  } catch (e) {
    throw new Error(e + JSON.stringify(ur.toObject()));
  }
  // console.log(result);
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
