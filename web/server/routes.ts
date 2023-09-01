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

// TODO(TCN-1771) keep in sync with `connect-web-test`'s `test-routes` until we share code between test packages
import type {
  ConnectRouter,
  HandlerContext,
  ServiceImpl,
} from "@bufbuild/connect";
import {
  Code,
  ConnectError,
  decodeBinaryHeader,
  encodeBinaryHeader,
} from "@bufbuild/connect";
import type {
  EchoStatus,
  ResponseParameters,
  SimpleRequest,
  StreamingInputCallRequest,
  StreamingOutputCallRequest,
} from "../gen/proto/connect-web/connectrpc/conformance/v1/messages_pb.js";
import {
  TestService,
  UnimplementedService,
} from "../gen/proto/connect-web/connectrpc/conformance/v1/test_connect.js";
import { interop } from "./interop.js";

export default (router: ConnectRouter) => {
  router.service(TestService, testService);
  router.service(UnimplementedService, unimplementedService);
};

const unimplementedService: ServiceImpl<typeof UnimplementedService> = {
  unimplementedCall() {
    throw new ConnectError("unimplemented", Code.Unimplemented);
  },
  // eslint-disable-next-line require-yield
  async *unimplementedStreamingOutputCall() {
    throw new ConnectError("unimplemented", Code.Unimplemented);
  },
};

const testService: ServiceImpl<typeof TestService> = {
  emptyCall() {
    return {};
  },

  unaryCall(request: SimpleRequest, context: HandlerContext) {
    echoMetadata(context);
    maybeRaiseError(request.responseStatus);
    return {
      payload: interop.makeServerPayload(
        request.responseType,
        request.responseSize
      ),
    };
  },

  failUnaryCall() {
    throw new ConnectError(interop.nonASCIIErrMsg, Code.ResourceExhausted, {}, [
      interop.errorDetail,
    ]);
  },

  cacheableUnaryCall(request: SimpleRequest, context: HandlerContext) {
    if (context.requestMethod == "GET") {
      context.responseHeader.set("get-request", "true");
    }
    return this.unaryCall(request, context);
  },

  async *streamingOutputCall(
    request: StreamingOutputCallRequest,
    context: HandlerContext
  ) {
    echoMetadata(context);
    for (const param of request.responseParameters) {
      await maybeDelayResponse(param);
      context.signal.throwIfAborted();
      yield {
        payload: interop.makeServerPayload(request.responseType, param.size),
      };
    }
    maybeRaiseError(request.responseStatus);
  },

  async *failStreamingOutputCall(
    request: StreamingOutputCallRequest,
    context: HandlerContext
  ) {
    echoMetadata(context);
    for (const param of request.responseParameters) {
      await maybeDelayResponse(param);
      context.signal.throwIfAborted();
      yield {
        payload: interop.makeServerPayload(request.responseType, param.size),
      };
    }
    throw new ConnectError(interop.nonASCIIErrMsg, Code.ResourceExhausted, {}, [
      interop.errorDetail,
    ]);
  },

  async streamingInputCall(
    requests: AsyncIterable<StreamingInputCallRequest>,
    context: HandlerContext
  ) {
    echoMetadata(context);
    let total = 0;
    for await (const req of requests) {
      total += req.payload?.body.length ?? 0;
    }
    return {
      aggregatedPayloadSize: total,
    };
  },

  async *fullDuplexCall(
    requests: AsyncIterable<StreamingOutputCallRequest>,
    context: HandlerContext
  ) {
    echoMetadata(context);
    for await (const req of requests) {
      for (const param of req.responseParameters) {
        await maybeDelayResponse(param);
        context.signal.throwIfAborted();
        yield {
          payload: interop.makeServerPayload(req.responseType, param.size),
        };
      }
      maybeRaiseError(req.responseStatus);
    }
  },

  async *halfDuplexCall(
    requests: AsyncIterable<StreamingOutputCallRequest>,
    context: HandlerContext
  ) {
    echoMetadata(context);
    const buffer: StreamingOutputCallRequest[] = [];
    for await (const req of requests) {
      buffer.push(req);
    }
    for await (const req of buffer) {
      for (const param of req.responseParameters) {
        await maybeDelayResponse(param);
        context.signal.throwIfAborted();
        yield {
          payload: interop.makeServerPayload(req.responseType, param.size),
        };
      }
      maybeRaiseError(req.responseStatus);
    }
  },

  unimplementedCall(/*request*/) {
    throw new ConnectError(
      "grpc.testing.TestService.UnimplementedCall is not implemented",
      Code.Unimplemented
    );
  },

  // eslint-disable-next-line @typescript-eslint/require-await,require-yield
  async *unimplementedStreamingOutputCall(/*requests*/) {
    throw new ConnectError(
      "grpc.testing.TestService.UnimplementedStreamingOutputCall is not implemented",
      Code.Unimplemented
    );
  },
};

async function maybeDelayResponse(param: ResponseParameters) {
  if (param.intervalUs > 0) {
    await new Promise<void>((resolve) => {
      setTimeout(resolve, param.intervalUs / 1000);
    });
  }
}

function maybeRaiseError(status: EchoStatus | undefined): void {
  if (!status || status.code <= 0) {
    return;
  }
  throw new ConnectError(status.message, status.code);
}

function echoMetadata(ctx: HandlerContext) {
  const hdrs = ctx.requestHeader.get(interop.leadingMetadataKey);
  if (hdrs) {
    ctx.responseHeader.append(interop.leadingMetadataKey, hdrs);
  }
  const trailer = ctx.requestHeader.get(interop.trailingMetadataKey);
  if (trailer) {
    const vals = trailer.split(",");
    vals.forEach((hdr) => {
      const decoded = decodeBinaryHeader(hdr);
      ctx.responseTrailer.append(
        interop.trailingMetadataKey,
        encodeBinaryHeader(decoded)
      );
    });
  }
}
