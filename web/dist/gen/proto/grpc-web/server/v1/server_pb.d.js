var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
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
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target, mod));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);
var server_pb_d_exports = {};
__export(server_pb_d_exports, {
  HTTPVersion: () => HTTPVersion,
  Protocol: () => Protocol,
  ProtocolSupport: () => ProtocolSupport,
  ServerMetadata: () => ServerMetadata
});
module.exports = __toCommonJS(server_pb_d_exports);
var jspb = __toESM(require("google-protobuf"));
class ServerMetadata extends jspb.Message {
}
class ProtocolSupport extends jspb.Message {
}
class HTTPVersion extends jspb.Message {
}
var Protocol = /* @__PURE__ */ ((Protocol2) => {
  Protocol2[Protocol2["PROTOCOL_UNSPECIFIED"] = 0] = "PROTOCOL_UNSPECIFIED";
  Protocol2[Protocol2["PROTOCOL_GRPC"] = 1] = "PROTOCOL_GRPC";
  Protocol2[Protocol2["PROTOCOL_GRPC_WEB"] = 2] = "PROTOCOL_GRPC_WEB";
  return Protocol2;
})(Protocol || {});
