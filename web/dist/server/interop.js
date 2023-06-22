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
var interop_exports = {};
__export(interop_exports, {
  interop: () => interop
});
module.exports = __toCommonJS(interop_exports);
var import_messages_pb = require("../gen/proto/connect-web/grpc/testing/messages_pb.js");
var import_connect = require("@bufbuild/connect");
const interop = {
  nonASCIIErrMsg: "soir\xE9e \u{1F389}",
  errorDetail: new import_messages_pb.ErrorDetail({
    reason: "soir\xE9e \u{1F389}",
    domain: "connect-crosstest"
  }),
  leadingMetadataKey: "x-grpc-test-echo-initial",
  trailingMetadataKey: "x-grpc-test-echo-trailing-bin",
  makeServerPayload(payloadType, size) {
    switch (payloadType) {
      case import_messages_pb.PayloadType.COMPRESSABLE:
        return new import_messages_pb.Payload({
          body: new Uint8Array(size),
          type: import_messages_pb.PayloadType.COMPRESSABLE
        });
      default:
        throw new Error(`unsupported payload type: ${payloadType}`);
    }
  },
  corsOptions: {
    origin: true,
    methods: [...import_connect.cors.allowedMethods],
    allowedHeaders: [
      ...import_connect.cors.allowedHeaders,
      "X-Grpc-Test-Echo-Initial",
      "X-Grpc-Test-Echo-Trailing-Bin",
      "Request-Protocol",
      "Get-Request"
    ],
    exposedHeaders: [
      ...import_connect.cors.exposedHeaders,
      "X-Grpc-Test-Echo-Initial",
      "X-Grpc-Test-Echo-Trailing-Bin",
      "Trailer-X-Grpc-Test-Echo-Trailing-Bin",
      "Request-Protocol",
      "Get-Request"
    ]
  }
};
