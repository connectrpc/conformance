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

/**
 * @fileoverview gRPC-Web generated client stub for grpc.testing
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var grpc_testing_empty_pb = require('../../grpc/testing/empty_pb.js')

var grpc_testing_messages_pb = require('../../grpc/testing/messages_pb.js')
const proto = {};
proto.grpc = {};
proto.grpc.testing = require('./test_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.TestServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.TestServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_TestService_EmptyCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/EmptyCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.emptyCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.TestService/EmptyCall',
      request,
      metadata || {},
      methodDescriptor_TestService_EmptyCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.TestServicePromiseClient.prototype.emptyCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.TestService/EmptyCall',
      request,
      metadata || {},
      methodDescriptor_TestService_EmptyCall);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.SimpleRequest,
 *   !proto.grpc.testing.SimpleResponse>}
 */
const methodDescriptor_TestService_UnaryCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/UnaryCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.SimpleRequest,
  grpc_testing_messages_pb.SimpleResponse,
  /**
   * @param {!proto.grpc.testing.SimpleRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.SimpleResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.SimpleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.SimpleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.unaryCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.TestService/UnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_UnaryCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.SimpleResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.TestServicePromiseClient.prototype.unaryCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.TestService/UnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_UnaryCall);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.SimpleRequest,
 *   !proto.grpc.testing.SimpleResponse>}
 */
const methodDescriptor_TestService_FailUnaryCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/FailUnaryCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.SimpleRequest,
  grpc_testing_messages_pb.SimpleResponse,
  /**
   * @param {!proto.grpc.testing.SimpleRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.SimpleResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.SimpleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.SimpleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.failUnaryCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.TestService/FailUnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_FailUnaryCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.SimpleResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.TestServicePromiseClient.prototype.failUnaryCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.TestService/FailUnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_FailUnaryCall);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.SimpleRequest,
 *   !proto.grpc.testing.SimpleResponse>}
 */
const methodDescriptor_TestService_CacheableUnaryCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/CacheableUnaryCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.SimpleRequest,
  grpc_testing_messages_pb.SimpleResponse,
  /**
   * @param {!proto.grpc.testing.SimpleRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.SimpleResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.SimpleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.SimpleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.cacheableUnaryCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.TestService/CacheableUnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_CacheableUnaryCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.SimpleRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.SimpleResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.TestServicePromiseClient.prototype.cacheableUnaryCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.TestService/CacheableUnaryCall',
      request,
      metadata || {},
      methodDescriptor_TestService_CacheableUnaryCall);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.StreamingOutputCallRequest,
 *   !proto.grpc.testing.StreamingOutputCallResponse>}
 */
const methodDescriptor_TestService_StreamingOutputCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/StreamingOutputCall',
  grpc.web.MethodType.SERVER_STREAMING,
  grpc_testing_messages_pb.StreamingOutputCallRequest,
  grpc_testing_messages_pb.StreamingOutputCallResponse,
  /**
   * @param {!proto.grpc.testing.StreamingOutputCallRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.StreamingOutputCallResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.StreamingOutputCallRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.StreamingOutputCallResponse>}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.streamingOutputCall =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/grpc.testing.TestService/StreamingOutputCall',
      request,
      metadata || {},
      methodDescriptor_TestService_StreamingOutputCall);
};


/**
 * @param {!proto.grpc.testing.StreamingOutputCallRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.StreamingOutputCallResponse>}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServicePromiseClient.prototype.streamingOutputCall =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/grpc.testing.TestService/StreamingOutputCall',
      request,
      metadata || {},
      methodDescriptor_TestService_StreamingOutputCall);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_TestService_UnimplementedCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.TestService/UnimplementedCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.TestServiceClient.prototype.unimplementedCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.TestService/UnimplementedCall',
      request,
      metadata || {},
      methodDescriptor_TestService_UnimplementedCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.TestServicePromiseClient.prototype.unimplementedCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.TestService/UnimplementedCall',
      request,
      metadata || {},
      methodDescriptor_TestService_UnimplementedCall);
};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.UnimplementedServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.UnimplementedServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_UnimplementedService_UnimplementedCall = new grpc.web.MethodDescriptor(
  '/grpc.testing.UnimplementedService/UnimplementedCall',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.UnimplementedServiceClient.prototype.unimplementedCall =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.UnimplementedService/UnimplementedCall',
      request,
      metadata || {},
      methodDescriptor_UnimplementedService_UnimplementedCall,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.UnimplementedServicePromiseClient.prototype.unimplementedCall =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.UnimplementedService/UnimplementedCall',
      request,
      metadata || {},
      methodDescriptor_UnimplementedService_UnimplementedCall);
};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.ReconnectServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.ReconnectServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.ReconnectParams,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_ReconnectService_Start = new grpc.web.MethodDescriptor(
  '/grpc.testing.ReconnectService/Start',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.ReconnectParams,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.ReconnectParams} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.ReconnectParams} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.ReconnectServiceClient.prototype.start =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.ReconnectService/Start',
      request,
      metadata || {},
      methodDescriptor_ReconnectService_Start,
      callback);
};


/**
 * @param {!proto.grpc.testing.ReconnectParams} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.ReconnectServicePromiseClient.prototype.start =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.ReconnectService/Start',
      request,
      metadata || {},
      methodDescriptor_ReconnectService_Start);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.ReconnectInfo>}
 */
const methodDescriptor_ReconnectService_Stop = new grpc.web.MethodDescriptor(
  '/grpc.testing.ReconnectService/Stop',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_messages_pb.ReconnectInfo,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.ReconnectInfo.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.ReconnectInfo)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.ReconnectInfo>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.ReconnectServiceClient.prototype.stop =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.ReconnectService/Stop',
      request,
      metadata || {},
      methodDescriptor_ReconnectService_Stop,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.ReconnectInfo>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.ReconnectServicePromiseClient.prototype.stop =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.ReconnectService/Stop',
      request,
      metadata || {},
      methodDescriptor_ReconnectService_Stop);
};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.LoadBalancerStatsServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.LoadBalancerStatsServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.LoadBalancerStatsRequest,
 *   !proto.grpc.testing.LoadBalancerStatsResponse>}
 */
const methodDescriptor_LoadBalancerStatsService_GetClientStats = new grpc.web.MethodDescriptor(
  '/grpc.testing.LoadBalancerStatsService/GetClientStats',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.LoadBalancerStatsRequest,
  grpc_testing_messages_pb.LoadBalancerStatsResponse,
  /**
   * @param {!proto.grpc.testing.LoadBalancerStatsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.LoadBalancerStatsResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.LoadBalancerStatsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.LoadBalancerStatsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.LoadBalancerStatsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.LoadBalancerStatsServiceClient.prototype.getClientStats =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientStats',
      request,
      metadata || {},
      methodDescriptor_LoadBalancerStatsService_GetClientStats,
      callback);
};


/**
 * @param {!proto.grpc.testing.LoadBalancerStatsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.LoadBalancerStatsResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.LoadBalancerStatsServicePromiseClient.prototype.getClientStats =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientStats',
      request,
      metadata || {},
      methodDescriptor_LoadBalancerStatsService_GetClientStats);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.LoadBalancerAccumulatedStatsRequest,
 *   !proto.grpc.testing.LoadBalancerAccumulatedStatsResponse>}
 */
const methodDescriptor_LoadBalancerStatsService_GetClientAccumulatedStats = new grpc.web.MethodDescriptor(
  '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest,
  grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse,
  /**
   * @param {!proto.grpc.testing.LoadBalancerAccumulatedStatsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.LoadBalancerAccumulatedStatsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.LoadBalancerAccumulatedStatsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.LoadBalancerAccumulatedStatsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.LoadBalancerStatsServiceClient.prototype.getClientAccumulatedStats =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
      request,
      metadata || {},
      methodDescriptor_LoadBalancerStatsService_GetClientAccumulatedStats,
      callback);
};


/**
 * @param {!proto.grpc.testing.LoadBalancerAccumulatedStatsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.LoadBalancerAccumulatedStatsResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.LoadBalancerStatsServicePromiseClient.prototype.getClientAccumulatedStats =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
      request,
      metadata || {},
      methodDescriptor_LoadBalancerStatsService_GetClientAccumulatedStats);
};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.XdsUpdateHealthServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.XdsUpdateHealthServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_XdsUpdateHealthService_SetServing = new grpc.web.MethodDescriptor(
  '/grpc.testing.XdsUpdateHealthService/SetServing',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.XdsUpdateHealthServiceClient.prototype.setServing =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetServing',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateHealthService_SetServing,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.XdsUpdateHealthServicePromiseClient.prototype.setServing =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetServing',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateHealthService_SetServing);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.Empty,
 *   !proto.grpc.testing.Empty>}
 */
const methodDescriptor_XdsUpdateHealthService_SetNotServing = new grpc.web.MethodDescriptor(
  '/grpc.testing.XdsUpdateHealthService/SetNotServing',
  grpc.web.MethodType.UNARY,
  grpc_testing_empty_pb.Empty,
  grpc_testing_empty_pb.Empty,
  /**
   * @param {!proto.grpc.testing.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.XdsUpdateHealthServiceClient.prototype.setNotServing =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetNotServing',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateHealthService_SetNotServing,
      callback);
};


/**
 * @param {!proto.grpc.testing.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.Empty>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.XdsUpdateHealthServicePromiseClient.prototype.setNotServing =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetNotServing',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateHealthService_SetNotServing);
};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.XdsUpdateClientConfigureServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.grpc.testing.XdsUpdateClientConfigureServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.grpc.testing.ClientConfigureRequest,
 *   !proto.grpc.testing.ClientConfigureResponse>}
 */
const methodDescriptor_XdsUpdateClientConfigureService_Configure = new grpc.web.MethodDescriptor(
  '/grpc.testing.XdsUpdateClientConfigureService/Configure',
  grpc.web.MethodType.UNARY,
  grpc_testing_messages_pb.ClientConfigureRequest,
  grpc_testing_messages_pb.ClientConfigureResponse,
  /**
   * @param {!proto.grpc.testing.ClientConfigureRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  grpc_testing_messages_pb.ClientConfigureResponse.deserializeBinary
);


/**
 * @param {!proto.grpc.testing.ClientConfigureRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.grpc.testing.ClientConfigureResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.grpc.testing.ClientConfigureResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.grpc.testing.XdsUpdateClientConfigureServiceClient.prototype.configure =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/grpc.testing.XdsUpdateClientConfigureService/Configure',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateClientConfigureService_Configure,
      callback);
};


/**
 * @param {!proto.grpc.testing.ClientConfigureRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.grpc.testing.ClientConfigureResponse>}
 *     Promise that resolves to the response
 */
proto.grpc.testing.XdsUpdateClientConfigureServicePromiseClient.prototype.configure =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/grpc.testing.XdsUpdateClientConfigureService/Configure',
      request,
      metadata || {},
      methodDescriptor_XdsUpdateClientConfigureService_Configure);
};


module.exports = proto.grpc.testing;

