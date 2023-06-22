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
var messages_pb_d_exports = {};
__export(messages_pb_d_exports, {
  BoolValue: () => BoolValue,
  ClientConfigureRequest: () => ClientConfigureRequest,
  ClientConfigureResponse: () => ClientConfigureResponse,
  EchoStatus: () => EchoStatus,
  ErrorDetail: () => ErrorDetail,
  ErrorStatus: () => ErrorStatus,
  GrpclbRouteType: () => GrpclbRouteType,
  LoadBalancerAccumulatedStatsRequest: () => LoadBalancerAccumulatedStatsRequest,
  LoadBalancerAccumulatedStatsResponse: () => LoadBalancerAccumulatedStatsResponse,
  LoadBalancerStatsRequest: () => LoadBalancerStatsRequest,
  LoadBalancerStatsResponse: () => LoadBalancerStatsResponse,
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
module.exports = __toCommonJS(messages_pb_d_exports);
var jspb = __toESM(require("google-protobuf"));
class BoolValue extends jspb.Message {
}
class Payload extends jspb.Message {
}
class EchoStatus extends jspb.Message {
}
class SimpleRequest extends jspb.Message {
}
class SimpleResponse extends jspb.Message {
}
class StreamingInputCallRequest extends jspb.Message {
}
class StreamingInputCallResponse extends jspb.Message {
}
class ResponseParameters extends jspb.Message {
}
class StreamingOutputCallRequest extends jspb.Message {
}
class StreamingOutputCallResponse extends jspb.Message {
}
class ReconnectParams extends jspb.Message {
}
class ReconnectInfo extends jspb.Message {
}
class LoadBalancerStatsRequest extends jspb.Message {
}
class LoadBalancerStatsResponse extends jspb.Message {
}
((LoadBalancerStatsResponse2) => {
  class RpcsByPeer extends jspb.Message {
  }
  LoadBalancerStatsResponse2.RpcsByPeer = RpcsByPeer;
})(LoadBalancerStatsResponse || (LoadBalancerStatsResponse = {}));
class LoadBalancerAccumulatedStatsRequest extends jspb.Message {
}
class LoadBalancerAccumulatedStatsResponse extends jspb.Message {
}
((LoadBalancerAccumulatedStatsResponse2) => {
  class MethodStats extends jspb.Message {
  }
  LoadBalancerAccumulatedStatsResponse2.MethodStats = MethodStats;
})(LoadBalancerAccumulatedStatsResponse || (LoadBalancerAccumulatedStatsResponse = {}));
class ClientConfigureRequest extends jspb.Message {
}
((ClientConfigureRequest2) => {
  class Metadata extends jspb.Message {
  }
  ClientConfigureRequest2.Metadata = Metadata;
  let RpcType;
  ((RpcType2) => {
    RpcType2[RpcType2["EMPTY_CALL"] = 0] = "EMPTY_CALL";
    RpcType2[RpcType2["UNARY_CALL"] = 1] = "UNARY_CALL";
  })(RpcType = ClientConfigureRequest2.RpcType || (ClientConfigureRequest2.RpcType = {}));
})(ClientConfigureRequest || (ClientConfigureRequest = {}));
class ClientConfigureResponse extends jspb.Message {
}
class ErrorDetail extends jspb.Message {
}
class ErrorStatus extends jspb.Message {
}
var PayloadType = /* @__PURE__ */ ((PayloadType2) => {
  PayloadType2[PayloadType2["COMPRESSABLE"] = 0] = "COMPRESSABLE";
  return PayloadType2;
})(PayloadType || {});
var GrpclbRouteType = /* @__PURE__ */ ((GrpclbRouteType2) => {
  GrpclbRouteType2[GrpclbRouteType2["GRPCLB_ROUTE_TYPE_UNKNOWN"] = 0] = "GRPCLB_ROUTE_TYPE_UNKNOWN";
  GrpclbRouteType2[GrpclbRouteType2["GRPCLB_ROUTE_TYPE_FALLBACK"] = 1] = "GRPCLB_ROUTE_TYPE_FALLBACK";
  GrpclbRouteType2[GrpclbRouteType2["GRPCLB_ROUTE_TYPE_BACKEND"] = 2] = "GRPCLB_ROUTE_TYPE_BACKEND";
  return GrpclbRouteType2;
})(GrpclbRouteType || {});
