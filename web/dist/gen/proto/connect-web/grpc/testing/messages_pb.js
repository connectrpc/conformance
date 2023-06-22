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
var messages_pb_exports = {};
__export(messages_pb_exports, {
  BoolValue: () => BoolValue,
  ClientConfigureRequest: () => ClientConfigureRequest,
  ClientConfigureRequest_Metadata: () => ClientConfigureRequest_Metadata,
  ClientConfigureRequest_RpcType: () => ClientConfigureRequest_RpcType,
  ClientConfigureResponse: () => ClientConfigureResponse,
  EchoStatus: () => EchoStatus,
  ErrorDetail: () => ErrorDetail,
  ErrorStatus: () => ErrorStatus,
  GrpclbRouteType: () => GrpclbRouteType,
  LoadBalancerAccumulatedStatsRequest: () => LoadBalancerAccumulatedStatsRequest,
  LoadBalancerAccumulatedStatsResponse: () => LoadBalancerAccumulatedStatsResponse,
  LoadBalancerAccumulatedStatsResponse_MethodStats: () => LoadBalancerAccumulatedStatsResponse_MethodStats,
  LoadBalancerStatsRequest: () => LoadBalancerStatsRequest,
  LoadBalancerStatsResponse: () => LoadBalancerStatsResponse,
  LoadBalancerStatsResponse_RpcsByPeer: () => LoadBalancerStatsResponse_RpcsByPeer,
  Payload: () => Payload,
  PayloadType: () => PayloadType,
  ReconnectInfo: () => ReconnectInfo,
  ReconnectParams: () => ReconnectParams,
  ResponseParameters: () => ResponseParameters,
  SimpleRequest: () => SimpleRequest,
  SimpleResponse: () => SimpleResponse,
  StreamingInputCallRequest: () => StreamingInputCallRequest,
  StreamingInputCallResponse: () => StreamingInputCallResponse,
  StreamingOutputCallRequest: () => StreamingOutputCallRequest,
  StreamingOutputCallResponse: () => StreamingOutputCallResponse
});
module.exports = __toCommonJS(messages_pb_exports);
var import_protobuf = require("@bufbuild/protobuf");
var PayloadType = /* @__PURE__ */ ((PayloadType2) => {
  PayloadType2[PayloadType2["COMPRESSABLE"] = 0] = "COMPRESSABLE";
  return PayloadType2;
})(PayloadType || {});
import_protobuf.proto3.util.setEnumType(PayloadType, "grpc.testing.PayloadType", [
  { no: 0, name: "COMPRESSABLE" }
]);
var GrpclbRouteType = /* @__PURE__ */ ((GrpclbRouteType2) => {
  GrpclbRouteType2[GrpclbRouteType2["UNKNOWN"] = 0] = "UNKNOWN";
  GrpclbRouteType2[GrpclbRouteType2["FALLBACK"] = 1] = "FALLBACK";
  GrpclbRouteType2[GrpclbRouteType2["BACKEND"] = 2] = "BACKEND";
  return GrpclbRouteType2;
})(GrpclbRouteType || {});
import_protobuf.proto3.util.setEnumType(GrpclbRouteType, "grpc.testing.GrpclbRouteType", [
  { no: 0, name: "GRPCLB_ROUTE_TYPE_UNKNOWN" },
  { no: 1, name: "GRPCLB_ROUTE_TYPE_FALLBACK" },
  { no: 2, name: "GRPCLB_ROUTE_TYPE_BACKEND" }
]);
class BoolValue extends import_protobuf.Message {
  value = false;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.BoolValue";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 8 }
  ]);
  static fromBinary(bytes, options) {
    return new BoolValue().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new BoolValue().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new BoolValue().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(BoolValue, a, b);
  }
}
class Payload extends import_protobuf.Message {
  type = 0 /* COMPRESSABLE */;
  body = new Uint8Array(0);
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.Payload";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "type", kind: "enum", T: import_protobuf.proto3.getEnumType(PayloadType) },
    { no: 2, name: "body", kind: "scalar", T: 12 }
  ]);
  static fromBinary(bytes, options) {
    return new Payload().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new Payload().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new Payload().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(Payload, a, b);
  }
}
class EchoStatus extends import_protobuf.Message {
  code = 0;
  message = "";
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.EchoStatus";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "code", kind: "scalar", T: 5 },
    { no: 2, name: "message", kind: "scalar", T: 9 }
  ]);
  static fromBinary(bytes, options) {
    return new EchoStatus().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new EchoStatus().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new EchoStatus().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(EchoStatus, a, b);
  }
}
class SimpleRequest extends import_protobuf.Message {
  responseType = 0 /* COMPRESSABLE */;
  responseSize = 0;
  payload;
  fillUsername = false;
  fillOauthScope = false;
  responseCompressed;
  responseStatus;
  expectCompressed;
  fillServerId = false;
  fillGrpclbRouteType = false;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.SimpleRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "response_type", kind: "enum", T: import_protobuf.proto3.getEnumType(PayloadType) },
    { no: 2, name: "response_size", kind: "scalar", T: 5 },
    { no: 3, name: "payload", kind: "message", T: Payload },
    { no: 4, name: "fill_username", kind: "scalar", T: 8 },
    { no: 5, name: "fill_oauth_scope", kind: "scalar", T: 8 },
    { no: 6, name: "response_compressed", kind: "message", T: BoolValue },
    { no: 7, name: "response_status", kind: "message", T: EchoStatus },
    { no: 8, name: "expect_compressed", kind: "message", T: BoolValue },
    { no: 9, name: "fill_server_id", kind: "scalar", T: 8 },
    { no: 10, name: "fill_grpclb_route_type", kind: "scalar", T: 8 }
  ]);
  static fromBinary(bytes, options) {
    return new SimpleRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new SimpleRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new SimpleRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(SimpleRequest, a, b);
  }
}
class SimpleResponse extends import_protobuf.Message {
  payload;
  username = "";
  oauthScope = "";
  serverId = "";
  grpclbRouteType = 0 /* UNKNOWN */;
  hostname = "";
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.SimpleResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "payload", kind: "message", T: Payload },
    { no: 2, name: "username", kind: "scalar", T: 9 },
    { no: 3, name: "oauth_scope", kind: "scalar", T: 9 },
    { no: 4, name: "server_id", kind: "scalar", T: 9 },
    { no: 5, name: "grpclb_route_type", kind: "enum", T: import_protobuf.proto3.getEnumType(GrpclbRouteType) },
    { no: 6, name: "hostname", kind: "scalar", T: 9 }
  ]);
  static fromBinary(bytes, options) {
    return new SimpleResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new SimpleResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new SimpleResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(SimpleResponse, a, b);
  }
}
class StreamingInputCallRequest extends import_protobuf.Message {
  payload;
  expectCompressed;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.StreamingInputCallRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "payload", kind: "message", T: Payload },
    { no: 2, name: "expect_compressed", kind: "message", T: BoolValue }
  ]);
  static fromBinary(bytes, options) {
    return new StreamingInputCallRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new StreamingInputCallRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new StreamingInputCallRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(StreamingInputCallRequest, a, b);
  }
}
class StreamingInputCallResponse extends import_protobuf.Message {
  aggregatedPayloadSize = 0;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.StreamingInputCallResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "aggregated_payload_size", kind: "scalar", T: 5 }
  ]);
  static fromBinary(bytes, options) {
    return new StreamingInputCallResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new StreamingInputCallResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new StreamingInputCallResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(StreamingInputCallResponse, a, b);
  }
}
class ResponseParameters extends import_protobuf.Message {
  size = 0;
  intervalUs = 0;
  compressed;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ResponseParameters";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "size", kind: "scalar", T: 5 },
    { no: 2, name: "interval_us", kind: "scalar", T: 5 },
    { no: 3, name: "compressed", kind: "message", T: BoolValue }
  ]);
  static fromBinary(bytes, options) {
    return new ResponseParameters().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ResponseParameters().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ResponseParameters().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ResponseParameters, a, b);
  }
}
class StreamingOutputCallRequest extends import_protobuf.Message {
  responseType = 0 /* COMPRESSABLE */;
  responseParameters = [];
  payload;
  responseStatus;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.StreamingOutputCallRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "response_type", kind: "enum", T: import_protobuf.proto3.getEnumType(PayloadType) },
    { no: 2, name: "response_parameters", kind: "message", T: ResponseParameters, repeated: true },
    { no: 3, name: "payload", kind: "message", T: Payload },
    { no: 7, name: "response_status", kind: "message", T: EchoStatus }
  ]);
  static fromBinary(bytes, options) {
    return new StreamingOutputCallRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new StreamingOutputCallRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new StreamingOutputCallRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(StreamingOutputCallRequest, a, b);
  }
}
class StreamingOutputCallResponse extends import_protobuf.Message {
  payload;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.StreamingOutputCallResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "payload", kind: "message", T: Payload }
  ]);
  static fromBinary(bytes, options) {
    return new StreamingOutputCallResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new StreamingOutputCallResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new StreamingOutputCallResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(StreamingOutputCallResponse, a, b);
  }
}
class ReconnectParams extends import_protobuf.Message {
  maxReconnectBackoffMs = 0;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ReconnectParams";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "max_reconnect_backoff_ms", kind: "scalar", T: 5 }
  ]);
  static fromBinary(bytes, options) {
    return new ReconnectParams().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ReconnectParams().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ReconnectParams().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ReconnectParams, a, b);
  }
}
class ReconnectInfo extends import_protobuf.Message {
  passed = false;
  backoffMs = [];
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ReconnectInfo";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "passed", kind: "scalar", T: 8 },
    { no: 2, name: "backoff_ms", kind: "scalar", T: 5, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new ReconnectInfo().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ReconnectInfo().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ReconnectInfo().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ReconnectInfo, a, b);
  }
}
class LoadBalancerStatsRequest extends import_protobuf.Message {
  numRpcs = 0;
  timeoutSec = 0;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerStatsRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "num_rpcs", kind: "scalar", T: 5 },
    { no: 2, name: "timeout_sec", kind: "scalar", T: 5 }
  ]);
  static fromBinary(bytes, options) {
    return new LoadBalancerStatsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerStatsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerStatsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerStatsRequest, a, b);
  }
}
class LoadBalancerStatsResponse extends import_protobuf.Message {
  rpcsByPeer = {};
  numFailures = 0;
  rpcsByMethod = {};
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerStatsResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "rpcs_by_peer", kind: "map", K: 9, V: { kind: "scalar", T: 5 } },
    { no: 2, name: "num_failures", kind: "scalar", T: 5 },
    { no: 3, name: "rpcs_by_method", kind: "map", K: 9, V: { kind: "message", T: LoadBalancerStatsResponse_RpcsByPeer } }
  ]);
  static fromBinary(bytes, options) {
    return new LoadBalancerStatsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerStatsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerStatsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerStatsResponse, a, b);
  }
}
class LoadBalancerStatsResponse_RpcsByPeer extends import_protobuf.Message {
  rpcsByPeer = {};
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerStatsResponse.RpcsByPeer";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "rpcs_by_peer", kind: "map", K: 9, V: { kind: "scalar", T: 5 } }
  ]);
  static fromBinary(bytes, options) {
    return new LoadBalancerStatsResponse_RpcsByPeer().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerStatsResponse_RpcsByPeer().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerStatsResponse_RpcsByPeer().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerStatsResponse_RpcsByPeer, a, b);
  }
}
class LoadBalancerAccumulatedStatsRequest extends import_protobuf.Message {
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerAccumulatedStatsRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new LoadBalancerAccumulatedStatsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerAccumulatedStatsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerAccumulatedStatsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerAccumulatedStatsRequest, a, b);
  }
}
class LoadBalancerAccumulatedStatsResponse extends import_protobuf.Message {
  numRpcsStartedByMethod = {};
  numRpcsSucceededByMethod = {};
  numRpcsFailedByMethod = {};
  statsPerMethod = {};
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerAccumulatedStatsResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "num_rpcs_started_by_method", kind: "map", K: 9, V: { kind: "scalar", T: 5 } },
    { no: 2, name: "num_rpcs_succeeded_by_method", kind: "map", K: 9, V: { kind: "scalar", T: 5 } },
    { no: 3, name: "num_rpcs_failed_by_method", kind: "map", K: 9, V: { kind: "scalar", T: 5 } },
    { no: 4, name: "stats_per_method", kind: "map", K: 9, V: { kind: "message", T: LoadBalancerAccumulatedStatsResponse_MethodStats } }
  ]);
  static fromBinary(bytes, options) {
    return new LoadBalancerAccumulatedStatsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerAccumulatedStatsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerAccumulatedStatsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerAccumulatedStatsResponse, a, b);
  }
}
class LoadBalancerAccumulatedStatsResponse_MethodStats extends import_protobuf.Message {
  rpcsStarted = 0;
  result = {};
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.LoadBalancerAccumulatedStatsResponse.MethodStats";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "rpcs_started", kind: "scalar", T: 5 },
    { no: 2, name: "result", kind: "map", K: 5, V: { kind: "scalar", T: 5 } }
  ]);
  static fromBinary(bytes, options) {
    return new LoadBalancerAccumulatedStatsResponse_MethodStats().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new LoadBalancerAccumulatedStatsResponse_MethodStats().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new LoadBalancerAccumulatedStatsResponse_MethodStats().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(LoadBalancerAccumulatedStatsResponse_MethodStats, a, b);
  }
}
class ClientConfigureRequest extends import_protobuf.Message {
  types = [];
  metadata = [];
  timeoutSec = 0;
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ClientConfigureRequest";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "types", kind: "enum", T: import_protobuf.proto3.getEnumType(ClientConfigureRequest_RpcType), repeated: true },
    { no: 2, name: "metadata", kind: "message", T: ClientConfigureRequest_Metadata, repeated: true },
    { no: 3, name: "timeout_sec", kind: "scalar", T: 5 }
  ]);
  static fromBinary(bytes, options) {
    return new ClientConfigureRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ClientConfigureRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ClientConfigureRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ClientConfigureRequest, a, b);
  }
}
var ClientConfigureRequest_RpcType = /* @__PURE__ */ ((ClientConfigureRequest_RpcType2) => {
  ClientConfigureRequest_RpcType2[ClientConfigureRequest_RpcType2["EMPTY_CALL"] = 0] = "EMPTY_CALL";
  ClientConfigureRequest_RpcType2[ClientConfigureRequest_RpcType2["UNARY_CALL"] = 1] = "UNARY_CALL";
  return ClientConfigureRequest_RpcType2;
})(ClientConfigureRequest_RpcType || {});
import_protobuf.proto3.util.setEnumType(ClientConfigureRequest_RpcType, "grpc.testing.ClientConfigureRequest.RpcType", [
  { no: 0, name: "EMPTY_CALL" },
  { no: 1, name: "UNARY_CALL" }
]);
class ClientConfigureRequest_Metadata extends import_protobuf.Message {
  type = 0 /* EMPTY_CALL */;
  key = "";
  value = "";
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ClientConfigureRequest.Metadata";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "type", kind: "enum", T: import_protobuf.proto3.getEnumType(ClientConfigureRequest_RpcType) },
    { no: 2, name: "key", kind: "scalar", T: 9 },
    { no: 3, name: "value", kind: "scalar", T: 9 }
  ]);
  static fromBinary(bytes, options) {
    return new ClientConfigureRequest_Metadata().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ClientConfigureRequest_Metadata().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ClientConfigureRequest_Metadata().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ClientConfigureRequest_Metadata, a, b);
  }
}
class ClientConfigureResponse extends import_protobuf.Message {
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ClientConfigureResponse";
  static fields = import_protobuf.proto3.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new ClientConfigureResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ClientConfigureResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ClientConfigureResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ClientConfigureResponse, a, b);
  }
}
class ErrorDetail extends import_protobuf.Message {
  reason = "";
  domain = "";
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ErrorDetail";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "reason", kind: "scalar", T: 9 },
    { no: 2, name: "domain", kind: "scalar", T: 9 }
  ]);
  static fromBinary(bytes, options) {
    return new ErrorDetail().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ErrorDetail().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ErrorDetail().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ErrorDetail, a, b);
  }
}
class ErrorStatus extends import_protobuf.Message {
  code = 0;
  message = "";
  details = [];
  constructor(data) {
    super();
    import_protobuf.proto3.util.initPartial(data, this);
  }
  static runtime = import_protobuf.proto3;
  static typeName = "grpc.testing.ErrorStatus";
  static fields = import_protobuf.proto3.util.newFieldList(() => [
    { no: 1, name: "code", kind: "scalar", T: 5 },
    { no: 2, name: "message", kind: "scalar", T: 9 },
    { no: 3, name: "details", kind: "message", T: import_protobuf.Any, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new ErrorStatus().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new ErrorStatus().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new ErrorStatus().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return import_protobuf.proto3.util.equals(ErrorStatus, a, b);
  }
}
