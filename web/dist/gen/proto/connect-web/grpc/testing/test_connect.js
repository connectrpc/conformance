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
var test_connect_exports = {};
__export(test_connect_exports, {
  LoadBalancerStatsService: () => LoadBalancerStatsService,
  ReconnectService: () => ReconnectService,
  TestService: () => TestService,
  UnimplementedService: () => UnimplementedService,
  XdsUpdateClientConfigureService: () => XdsUpdateClientConfigureService,
  XdsUpdateHealthService: () => XdsUpdateHealthService
});
module.exports = __toCommonJS(test_connect_exports);
var import_empty_pb = require("./empty_pb.js");
var import_protobuf = require("@bufbuild/protobuf");
var import_messages_pb = require("./messages_pb.js");
const TestService = {
  typeName: "grpc.testing.TestService",
  methods: {
    emptyCall: {
      name: "EmptyCall",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    },
    unaryCall: {
      name: "UnaryCall",
      I: import_messages_pb.SimpleRequest,
      O: import_messages_pb.SimpleResponse,
      kind: import_protobuf.MethodKind.Unary
    },
    failUnaryCall: {
      name: "FailUnaryCall",
      I: import_messages_pb.SimpleRequest,
      O: import_messages_pb.SimpleResponse,
      kind: import_protobuf.MethodKind.Unary
    },
    cacheableUnaryCall: {
      name: "CacheableUnaryCall",
      I: import_messages_pb.SimpleRequest,
      O: import_messages_pb.SimpleResponse,
      kind: import_protobuf.MethodKind.Unary,
      idempotency: import_protobuf.MethodIdempotency.NoSideEffects
    },
    streamingOutputCall: {
      name: "StreamingOutputCall",
      I: import_messages_pb.StreamingOutputCallRequest,
      O: import_messages_pb.StreamingOutputCallResponse,
      kind: import_protobuf.MethodKind.ServerStreaming
    },
    failStreamingOutputCall: {
      name: "FailStreamingOutputCall",
      I: import_messages_pb.StreamingOutputCallRequest,
      O: import_messages_pb.StreamingOutputCallResponse,
      kind: import_protobuf.MethodKind.ServerStreaming
    },
    streamingInputCall: {
      name: "StreamingInputCall",
      I: import_messages_pb.StreamingInputCallRequest,
      O: import_messages_pb.StreamingInputCallResponse,
      kind: import_protobuf.MethodKind.ClientStreaming
    },
    fullDuplexCall: {
      name: "FullDuplexCall",
      I: import_messages_pb.StreamingOutputCallRequest,
      O: import_messages_pb.StreamingOutputCallResponse,
      kind: import_protobuf.MethodKind.BiDiStreaming
    },
    halfDuplexCall: {
      name: "HalfDuplexCall",
      I: import_messages_pb.StreamingOutputCallRequest,
      O: import_messages_pb.StreamingOutputCallResponse,
      kind: import_protobuf.MethodKind.BiDiStreaming
    },
    unimplementedCall: {
      name: "UnimplementedCall",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    },
    unimplementedStreamingOutputCall: {
      name: "UnimplementedStreamingOutputCall",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.ServerStreaming
    }
  }
};
const UnimplementedService = {
  typeName: "grpc.testing.UnimplementedService",
  methods: {
    unimplementedCall: {
      name: "UnimplementedCall",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    },
    unimplementedStreamingOutputCall: {
      name: "UnimplementedStreamingOutputCall",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.ServerStreaming
    }
  }
};
const ReconnectService = {
  typeName: "grpc.testing.ReconnectService",
  methods: {
    start: {
      name: "Start",
      I: import_messages_pb.ReconnectParams,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    },
    stop: {
      name: "Stop",
      I: import_empty_pb.Empty,
      O: import_messages_pb.ReconnectInfo,
      kind: import_protobuf.MethodKind.Unary
    }
  }
};
const LoadBalancerStatsService = {
  typeName: "grpc.testing.LoadBalancerStatsService",
  methods: {
    getClientStats: {
      name: "GetClientStats",
      I: import_messages_pb.LoadBalancerStatsRequest,
      O: import_messages_pb.LoadBalancerStatsResponse,
      kind: import_protobuf.MethodKind.Unary
    },
    getClientAccumulatedStats: {
      name: "GetClientAccumulatedStats",
      I: import_messages_pb.LoadBalancerAccumulatedStatsRequest,
      O: import_messages_pb.LoadBalancerAccumulatedStatsResponse,
      kind: import_protobuf.MethodKind.Unary
    }
  }
};
const XdsUpdateHealthService = {
  typeName: "grpc.testing.XdsUpdateHealthService",
  methods: {
    setServing: {
      name: "SetServing",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    },
    setNotServing: {
      name: "SetNotServing",
      I: import_empty_pb.Empty,
      O: import_empty_pb.Empty,
      kind: import_protobuf.MethodKind.Unary
    }
  }
};
const XdsUpdateClientConfigureService = {
  typeName: "grpc.testing.XdsUpdateClientConfigureService",
  methods: {
    configure: {
      name: "Configure",
      I: import_messages_pb.ClientConfigureRequest,
      O: import_messages_pb.ClientConfigureResponse,
      kind: import_protobuf.MethodKind.Unary
    }
  }
};
