var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);
var routes_exports = {};
__export(routes_exports, {
  default: () => routes_default
});
module.exports = __toCommonJS(routes_exports);
var import_connect = require("@bufbuild/connect");
var import_test_connect = require("../gen/proto/connect-web/grpc/testing/test_connect.js");
var import_interop = require("./interop.js");
var routes_default = (router) => {
  router.service(import_test_connect.TestService, testService);
  router.service(import_test_connect.UnimplementedService, unimplementedService);
};
const unimplementedService = {
  unimplementedCall() {
    throw new import_connect.ConnectError("unimplemented", import_connect.Code.Unimplemented);
  },
  async *unimplementedStreamingOutputCall() {
    throw new import_connect.ConnectError("unimplemented", import_connect.Code.Unimplemented);
  }
};
const testService = {
  emptyCall() {
    return {};
  },
  unaryCall(request, context) {
    echoMetadata(context);
    maybeRaiseError(request.responseStatus);
    return {
      payload: import_interop.interop.makeServerPayload(request.responseType, request.responseSize)
    };
  },
  failUnaryCall() {
    throw new import_connect.ConnectError(import_interop.interop.nonASCIIErrMsg, import_connect.Code.ResourceExhausted, {}, [
      import_interop.interop.errorDetail
    ]);
  },
  cacheableUnaryCall(request, context) {
    if (context.requestMethod == "GET") {
      context.responseHeader.set("get-request", "true");
    }
    return this.unaryCall(request, context);
  },
  async *streamingOutputCall(request, context) {
    echoMetadata(context);
    for (const param of request.responseParameters) {
      await maybeDelayResponse(param);
      context.signal.throwIfAborted();
      yield {
        payload: import_interop.interop.makeServerPayload(request.responseType, param.size)
      };
    }
    maybeRaiseError(request.responseStatus);
  },
  async *failStreamingOutputCall(request, context) {
    echoMetadata(context);
    for (const param of request.responseParameters) {
      await maybeDelayResponse(param);
      context.signal.throwIfAborted();
      yield {
        payload: import_interop.interop.makeServerPayload(request.responseType, param.size)
      };
    }
    throw new import_connect.ConnectError(import_interop.interop.nonASCIIErrMsg, import_connect.Code.ResourceExhausted, {}, [
      import_interop.interop.errorDetail
    ]);
  },
  async streamingInputCall(requests, context) {
    echoMetadata(context);
    let total = 0;
    for await (const req of requests) {
      total += req.payload?.body.length ?? 0;
    }
    return {
      aggregatedPayloadSize: total
    };
  },
  async *fullDuplexCall(requests, context) {
    echoMetadata(context);
    for await (const req of requests) {
      for (const param of req.responseParameters) {
        await maybeDelayResponse(param);
        context.signal.throwIfAborted();
        yield {
          payload: import_interop.interop.makeServerPayload(req.responseType, param.size)
        };
      }
      maybeRaiseError(req.responseStatus);
    }
  },
  async *halfDuplexCall(requests, context) {
    echoMetadata(context);
    const buffer = [];
    for await (const req of requests) {
      buffer.push(req);
    }
    for await (const req of buffer) {
      for (const param of req.responseParameters) {
        await maybeDelayResponse(param);
        context.signal.throwIfAborted();
        yield {
          payload: import_interop.interop.makeServerPayload(req.responseType, param.size)
        };
      }
      maybeRaiseError(req.responseStatus);
    }
  },
  unimplementedCall() {
    throw new import_connect.ConnectError("grpc.testing.TestService.UnimplementedCall is not implemented", import_connect.Code.Unimplemented);
  },
  async *unimplementedStreamingOutputCall() {
    throw new import_connect.ConnectError("grpc.testing.TestService.UnimplementedStreamingOutputCall is not implemented", import_connect.Code.Unimplemented);
  }
};
async function maybeDelayResponse(param) {
  if (param.intervalUs > 0) {
    await new Promise((resolve) => {
      setTimeout(resolve, param.intervalUs / 1e3);
    });
  }
}
function maybeRaiseError(status) {
  if (!status || status.code <= 0) {
    return;
  }
  throw new import_connect.ConnectError(status.message, status.code);
}
function echoMetadata(ctx) {
  const hdrs = ctx.requestHeader.get(import_interop.interop.leadingMetadataKey);
  if (hdrs) {
    ctx.responseHeader.append(import_interop.interop.leadingMetadataKey, hdrs);
  }
  const trailer = ctx.requestHeader.get(import_interop.interop.trailingMetadataKey);
  if (trailer) {
    const vals = trailer.split(",");
    vals.forEach((hdr) => {
      const decoded = (0, import_connect.decodeBinaryHeader)(hdr);
      ctx.responseTrailer.append(import_interop.interop.trailingMetadataKey, (0, import_connect.encodeBinaryHeader)(decoded));
    });
  }
}
