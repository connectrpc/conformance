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
var TestServiceClientPb_exports = {};
__export(TestServiceClientPb_exports, {
  LoadBalancerStatsServiceClient: () => LoadBalancerStatsServiceClient,
  ReconnectServiceClient: () => ReconnectServiceClient,
  TestServiceClient: () => TestServiceClient,
  UnimplementedServiceClient: () => UnimplementedServiceClient,
  XdsUpdateClientConfigureServiceClient: () => XdsUpdateClientConfigureServiceClient,
  XdsUpdateHealthServiceClient: () => XdsUpdateHealthServiceClient
});
module.exports = __toCommonJS(TestServiceClientPb_exports);
var grpcWeb = __toESM(require("grpc-web"));
var grpc_testing_messages_pb = __toESM(require("../../grpc/testing/messages_pb"));
var grpc_testing_empty_pb = __toESM(require("../../grpc/testing/empty_pb"));
class TestServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorEmptyCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/EmptyCall", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  emptyCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.TestService/EmptyCall", request, metadata || {}, this.methodDescriptorEmptyCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.TestService/EmptyCall", request, metadata || {}, this.methodDescriptorEmptyCall);
  }
  methodDescriptorUnaryCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/UnaryCall", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.SimpleRequest, grpc_testing_messages_pb.SimpleResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.SimpleResponse.deserializeBinary);
  unaryCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.TestService/UnaryCall", request, metadata || {}, this.methodDescriptorUnaryCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.TestService/UnaryCall", request, metadata || {}, this.methodDescriptorUnaryCall);
  }
  methodDescriptorFailUnaryCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/FailUnaryCall", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.SimpleRequest, grpc_testing_messages_pb.SimpleResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.SimpleResponse.deserializeBinary);
  failUnaryCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.TestService/FailUnaryCall", request, metadata || {}, this.methodDescriptorFailUnaryCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.TestService/FailUnaryCall", request, metadata || {}, this.methodDescriptorFailUnaryCall);
  }
  methodDescriptorCacheableUnaryCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/CacheableUnaryCall", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.SimpleRequest, grpc_testing_messages_pb.SimpleResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.SimpleResponse.deserializeBinary);
  cacheableUnaryCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.TestService/CacheableUnaryCall", request, metadata || {}, this.methodDescriptorCacheableUnaryCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.TestService/CacheableUnaryCall", request, metadata || {}, this.methodDescriptorCacheableUnaryCall);
  }
  methodDescriptorStreamingOutputCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/StreamingOutputCall", grpcWeb.MethodType.SERVER_STREAMING, grpc_testing_messages_pb.StreamingOutputCallRequest, grpc_testing_messages_pb.StreamingOutputCallResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.StreamingOutputCallResponse.deserializeBinary);
  streamingOutputCall(request, metadata) {
    return this.client_.serverStreaming(this.hostname_ + "/grpc.testing.TestService/StreamingOutputCall", request, metadata || {}, this.methodDescriptorStreamingOutputCall);
  }
  methodDescriptorFailStreamingOutputCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/FailStreamingOutputCall", grpcWeb.MethodType.SERVER_STREAMING, grpc_testing_messages_pb.StreamingOutputCallRequest, grpc_testing_messages_pb.StreamingOutputCallResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.StreamingOutputCallResponse.deserializeBinary);
  failStreamingOutputCall(request, metadata) {
    return this.client_.serverStreaming(this.hostname_ + "/grpc.testing.TestService/FailStreamingOutputCall", request, metadata || {}, this.methodDescriptorFailStreamingOutputCall);
  }
  methodDescriptorUnimplementedCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/UnimplementedCall", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  unimplementedCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.TestService/UnimplementedCall", request, metadata || {}, this.methodDescriptorUnimplementedCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.TestService/UnimplementedCall", request, metadata || {}, this.methodDescriptorUnimplementedCall);
  }
  methodDescriptorUnimplementedStreamingOutputCall = new grpcWeb.MethodDescriptor("/grpc.testing.TestService/UnimplementedStreamingOutputCall", grpcWeb.MethodType.SERVER_STREAMING, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  unimplementedStreamingOutputCall(request, metadata) {
    return this.client_.serverStreaming(this.hostname_ + "/grpc.testing.TestService/UnimplementedStreamingOutputCall", request, metadata || {}, this.methodDescriptorUnimplementedStreamingOutputCall);
  }
}
class UnimplementedServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorUnimplementedCall = new grpcWeb.MethodDescriptor("/grpc.testing.UnimplementedService/UnimplementedCall", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  unimplementedCall(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.UnimplementedService/UnimplementedCall", request, metadata || {}, this.methodDescriptorUnimplementedCall, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.UnimplementedService/UnimplementedCall", request, metadata || {}, this.methodDescriptorUnimplementedCall);
  }
  methodDescriptorUnimplementedStreamingOutputCall = new grpcWeb.MethodDescriptor("/grpc.testing.UnimplementedService/UnimplementedStreamingOutputCall", grpcWeb.MethodType.SERVER_STREAMING, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  unimplementedStreamingOutputCall(request, metadata) {
    return this.client_.serverStreaming(this.hostname_ + "/grpc.testing.UnimplementedService/UnimplementedStreamingOutputCall", request, metadata || {}, this.methodDescriptorUnimplementedStreamingOutputCall);
  }
}
class ReconnectServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorStart = new grpcWeb.MethodDescriptor("/grpc.testing.ReconnectService/Start", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.ReconnectParams, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  start(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.ReconnectService/Start", request, metadata || {}, this.methodDescriptorStart, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.ReconnectService/Start", request, metadata || {}, this.methodDescriptorStart);
  }
  methodDescriptorStop = new grpcWeb.MethodDescriptor("/grpc.testing.ReconnectService/Stop", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_messages_pb.ReconnectInfo, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.ReconnectInfo.deserializeBinary);
  stop(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.ReconnectService/Stop", request, metadata || {}, this.methodDescriptorStop, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.ReconnectService/Stop", request, metadata || {}, this.methodDescriptorStop);
  }
}
class LoadBalancerStatsServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorGetClientStats = new grpcWeb.MethodDescriptor("/grpc.testing.LoadBalancerStatsService/GetClientStats", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.LoadBalancerStatsRequest, grpc_testing_messages_pb.LoadBalancerStatsResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.LoadBalancerStatsResponse.deserializeBinary);
  getClientStats(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.LoadBalancerStatsService/GetClientStats", request, metadata || {}, this.methodDescriptorGetClientStats, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.LoadBalancerStatsService/GetClientStats", request, metadata || {}, this.methodDescriptorGetClientStats);
  }
  methodDescriptorGetClientAccumulatedStats = new grpcWeb.MethodDescriptor("/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest, grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse.deserializeBinary);
  getClientAccumulatedStats(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats", request, metadata || {}, this.methodDescriptorGetClientAccumulatedStats, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats", request, metadata || {}, this.methodDescriptorGetClientAccumulatedStats);
  }
}
class XdsUpdateHealthServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorSetServing = new grpcWeb.MethodDescriptor("/grpc.testing.XdsUpdateHealthService/SetServing", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  setServing(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.XdsUpdateHealthService/SetServing", request, metadata || {}, this.methodDescriptorSetServing, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.XdsUpdateHealthService/SetServing", request, metadata || {}, this.methodDescriptorSetServing);
  }
  methodDescriptorSetNotServing = new grpcWeb.MethodDescriptor("/grpc.testing.XdsUpdateHealthService/SetNotServing", grpcWeb.MethodType.UNARY, grpc_testing_empty_pb.Empty, grpc_testing_empty_pb.Empty, (request) => {
    return request.serializeBinary();
  }, grpc_testing_empty_pb.Empty.deserializeBinary);
  setNotServing(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.XdsUpdateHealthService/SetNotServing", request, metadata || {}, this.methodDescriptorSetNotServing, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.XdsUpdateHealthService/SetNotServing", request, metadata || {}, this.methodDescriptorSetNotServing);
  }
}
class XdsUpdateClientConfigureServiceClient {
  client_;
  hostname_;
  credentials_;
  options_;
  constructor(hostname, credentials, options) {
    if (!options)
      options = {};
    if (!credentials)
      credentials = {};
    options["format"] = "binary";
    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname.replace(/\/+$/, "");
    this.credentials_ = credentials;
    this.options_ = options;
  }
  methodDescriptorConfigure = new grpcWeb.MethodDescriptor("/grpc.testing.XdsUpdateClientConfigureService/Configure", grpcWeb.MethodType.UNARY, grpc_testing_messages_pb.ClientConfigureRequest, grpc_testing_messages_pb.ClientConfigureResponse, (request) => {
    return request.serializeBinary();
  }, grpc_testing_messages_pb.ClientConfigureResponse.deserializeBinary);
  configure(request, metadata, callback) {
    if (callback !== void 0) {
      return this.client_.rpcCall(this.hostname_ + "/grpc.testing.XdsUpdateClientConfigureService/Configure", request, metadata || {}, this.methodDescriptorConfigure, callback);
    }
    return this.client_.unaryCall(this.hostname_ + "/grpc.testing.XdsUpdateClientConfigureService/Configure", request, metadata || {}, this.methodDescriptorConfigure);
  }
}
