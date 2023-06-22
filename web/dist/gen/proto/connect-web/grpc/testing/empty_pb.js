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
var empty_pb_exports = {};
__export(empty_pb_exports, {
  Empty: () => Empty
});
module.exports = __toCommonJS(empty_pb_exports);
var import_protobuf = require("@bufbuild/protobuf");
class Empty extends import_protobuf.Message {
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.Empty";
  static fields = import_protobuf.proto3.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new Empty().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new Empty().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new Empty().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(Empty, a, b);
  }
}
