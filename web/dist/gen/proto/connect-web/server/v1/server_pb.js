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
var server_pb_exports = {};
__export(server_pb_exports, {
  HTTPVersion: () => HTTPVersion,
  Protocol: () => Protocol,
  ProtocolSupport: () => ProtocolSupport,
  ServerMetadata: () => ServerMetadata
});
module.exports = __toCommonJS(server_pb_exports);
var import_protobuf = require("@bufbuild/protobuf");
var Protocol = /* @__PURE__ */ ((Protocol2) => {
  Protocol2[Protocol2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  Protocol2[Protocol2["GRPC"] = 1] = "GRPC";
  Protocol2[Protocol2["GRPC_WEB"] = 2] = "GRPC_WEB";
  return Protocol2;
})(Protocol || {});
import_protobuf.proto3.util.setEnumType(Protocol, "server.v1.Protocol", [
  { no: 0, name: "PROTOCOL_UNSPECIFIED" },
  { no: 1, name: "PROTOCOL_GRPC" },
  { no: 2, name: "PROTOCOL_GRPC_WEB" }
]);
class ServerMetadata extends import_protobuf.Message {
  host = "";
  protocols = [];
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "server.v1.ServerMetadata";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "host", kind: "scalar", T: 9 },
    { no: 2, name: "protocols", kind: "message", T: ProtocolSupport, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new ServerMetadata().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ServerMetadata().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ServerMetadata().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ServerMetadata, a, b);
  }
}
class ProtocolSupport extends import_protobuf.Message {
  protocol = 0 /* UNSPECIFIED */;
  httpVersions = [];
  port = "";
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "server.v1.ProtocolSupport";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "protocol", kind: "enum", T: import_protobuf.proto3.getEnumType(Protocol) },
    { no: 2, name: "http_versions", kind: "message", T: HTTPVersion, repeated: true },
    { no: 3, name: "port", kind: "scalar", T: 9 }
  ]);
  static fromBinary(bytes, options) {
    return new ProtocolSupport().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ProtocolSupport().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ProtocolSupport().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ProtocolSupport, a, b);
  }
}
class HTTPVersion extends import_protobuf.Message {
  major = 0;
  minor = 0;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "server.v1.HTTPVersion";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "major", kind: "scalar", T: 5 },
    { no: 2, name: "minor", kind: "scalar", T: 5 }
  ]);
  static fromBinary(bytes, options) {
    return new HTTPVersion().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new HTTPVersion().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new HTTPVersion().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(HTTPVersion, a, b);
  }
}
