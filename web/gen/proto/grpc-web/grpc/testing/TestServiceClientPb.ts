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


import * as grpcWeb from 'grpc-web';

import * as grpc_testing_messages_pb from '../../grpc/testing/messages_pb';
import * as grpc_testing_empty_pb from '../../grpc/testing/empty_pb';


export class TestServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorEmptyCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/EmptyCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  emptyCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  emptyCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  emptyCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.TestService/EmptyCall',
        request,
        metadata || {},
        this.methodDescriptorEmptyCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.TestService/EmptyCall',
    request,
    metadata || {},
    this.methodDescriptorEmptyCall);
  }

  methodDescriptorUnaryCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/UnaryCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.SimpleRequest,
    grpc_testing_messages_pb.SimpleResponse,
    (request: grpc_testing_messages_pb.SimpleRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.SimpleResponse.deserializeBinary
  );

  unaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.SimpleResponse>;

  unaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.SimpleResponse>;

  unaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.TestService/UnaryCall',
        request,
        metadata || {},
        this.methodDescriptorUnaryCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.TestService/UnaryCall',
    request,
    metadata || {},
    this.methodDescriptorUnaryCall);
  }

  methodDescriptorFailUnaryCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/FailUnaryCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.SimpleRequest,
    grpc_testing_messages_pb.SimpleResponse,
    (request: grpc_testing_messages_pb.SimpleRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.SimpleResponse.deserializeBinary
  );

  failUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.SimpleResponse>;

  failUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.SimpleResponse>;

  failUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.TestService/FailUnaryCall',
        request,
        metadata || {},
        this.methodDescriptorFailUnaryCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.TestService/FailUnaryCall',
    request,
    metadata || {},
    this.methodDescriptorFailUnaryCall);
  }

  methodDescriptorCacheableUnaryCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/CacheableUnaryCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.SimpleRequest,
    grpc_testing_messages_pb.SimpleResponse,
    (request: grpc_testing_messages_pb.SimpleRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.SimpleResponse.deserializeBinary
  );

  cacheableUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.SimpleResponse>;

  cacheableUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.SimpleResponse>;

  cacheableUnaryCall(
    request: grpc_testing_messages_pb.SimpleRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.SimpleResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.TestService/CacheableUnaryCall',
        request,
        metadata || {},
        this.methodDescriptorCacheableUnaryCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.TestService/CacheableUnaryCall',
    request,
    metadata || {},
    this.methodDescriptorCacheableUnaryCall);
  }

  methodDescriptorStreamingOutputCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/StreamingOutputCall',
    grpcWeb.MethodType.SERVER_STREAMING,
    grpc_testing_messages_pb.StreamingOutputCallRequest,
    grpc_testing_messages_pb.StreamingOutputCallResponse,
    (request: grpc_testing_messages_pb.StreamingOutputCallRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.StreamingOutputCallResponse.deserializeBinary
  );

  streamingOutputCall(
    request: grpc_testing_messages_pb.StreamingOutputCallRequest,
    metadata?: grpcWeb.Metadata): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.StreamingOutputCallResponse> {
    return this.client_.serverStreaming(
      this.hostname_ +
        '/grpc.testing.TestService/StreamingOutputCall',
      request,
      metadata || {},
      this.methodDescriptorStreamingOutputCall);
  }

  methodDescriptorFailStreamingOutputCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/FailStreamingOutputCall',
    grpcWeb.MethodType.SERVER_STREAMING,
    grpc_testing_messages_pb.StreamingOutputCallRequest,
    grpc_testing_messages_pb.StreamingOutputCallResponse,
    (request: grpc_testing_messages_pb.StreamingOutputCallRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.StreamingOutputCallResponse.deserializeBinary
  );

  failStreamingOutputCall(
    request: grpc_testing_messages_pb.StreamingOutputCallRequest,
    metadata?: grpcWeb.Metadata): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.StreamingOutputCallResponse> {
    return this.client_.serverStreaming(
      this.hostname_ +
        '/grpc.testing.TestService/FailStreamingOutputCall',
      request,
      metadata || {},
      this.methodDescriptorFailStreamingOutputCall);
  }

  methodDescriptorUnimplementedCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.TestService/UnimplementedCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.TestService/UnimplementedCall',
        request,
        metadata || {},
        this.methodDescriptorUnimplementedCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.TestService/UnimplementedCall',
    request,
    metadata || {},
    this.methodDescriptorUnimplementedCall);
  }

}

export class UnimplementedServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorUnimplementedCall = new grpcWeb.MethodDescriptor(
    '/grpc.testing.UnimplementedService/UnimplementedCall',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  unimplementedCall(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.UnimplementedService/UnimplementedCall',
        request,
        metadata || {},
        this.methodDescriptorUnimplementedCall,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.UnimplementedService/UnimplementedCall',
    request,
    metadata || {},
    this.methodDescriptorUnimplementedCall);
  }

}

export class ReconnectServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorStart = new grpcWeb.MethodDescriptor(
    '/grpc.testing.ReconnectService/Start',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.ReconnectParams,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_messages_pb.ReconnectParams) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  start(
    request: grpc_testing_messages_pb.ReconnectParams,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  start(
    request: grpc_testing_messages_pb.ReconnectParams,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  start(
    request: grpc_testing_messages_pb.ReconnectParams,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.ReconnectService/Start',
        request,
        metadata || {},
        this.methodDescriptorStart,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.ReconnectService/Start',
    request,
    metadata || {},
    this.methodDescriptorStart);
  }

  methodDescriptorStop = new grpcWeb.MethodDescriptor(
    '/grpc.testing.ReconnectService/Stop',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_messages_pb.ReconnectInfo,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.ReconnectInfo.deserializeBinary
  );

  stop(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.ReconnectInfo>;

  stop(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.ReconnectInfo) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.ReconnectInfo>;

  stop(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.ReconnectInfo) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.ReconnectService/Stop',
        request,
        metadata || {},
        this.methodDescriptorStop,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.ReconnectService/Stop',
    request,
    metadata || {},
    this.methodDescriptorStop);
  }

}

export class LoadBalancerStatsServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorGetClientStats = new grpcWeb.MethodDescriptor(
    '/grpc.testing.LoadBalancerStatsService/GetClientStats',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.LoadBalancerStatsRequest,
    grpc_testing_messages_pb.LoadBalancerStatsResponse,
    (request: grpc_testing_messages_pb.LoadBalancerStatsRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.LoadBalancerStatsResponse.deserializeBinary
  );

  getClientStats(
    request: grpc_testing_messages_pb.LoadBalancerStatsRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.LoadBalancerStatsResponse>;

  getClientStats(
    request: grpc_testing_messages_pb.LoadBalancerStatsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.LoadBalancerStatsResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.LoadBalancerStatsResponse>;

  getClientStats(
    request: grpc_testing_messages_pb.LoadBalancerStatsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.LoadBalancerStatsResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.LoadBalancerStatsService/GetClientStats',
        request,
        metadata || {},
        this.methodDescriptorGetClientStats,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientStats',
    request,
    metadata || {},
    this.methodDescriptorGetClientStats);
  }

  methodDescriptorGetClientAccumulatedStats = new grpcWeb.MethodDescriptor(
    '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest,
    grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse,
    (request: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse.deserializeBinary
  );

  getClientAccumulatedStats(
    request: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse>;

  getClientAccumulatedStats(
    request: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse>;

  getClientAccumulatedStats(
    request: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.LoadBalancerAccumulatedStatsResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
        request,
        metadata || {},
        this.methodDescriptorGetClientAccumulatedStats,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.LoadBalancerStatsService/GetClientAccumulatedStats',
    request,
    metadata || {},
    this.methodDescriptorGetClientAccumulatedStats);
  }

}

export class XdsUpdateHealthServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorSetServing = new grpcWeb.MethodDescriptor(
    '/grpc.testing.XdsUpdateHealthService/SetServing',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  setServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  setServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  setServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.XdsUpdateHealthService/SetServing',
        request,
        metadata || {},
        this.methodDescriptorSetServing,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetServing',
    request,
    metadata || {},
    this.methodDescriptorSetServing);
  }

  methodDescriptorSetNotServing = new grpcWeb.MethodDescriptor(
    '/grpc.testing.XdsUpdateHealthService/SetNotServing',
    grpcWeb.MethodType.UNARY,
    grpc_testing_empty_pb.Empty,
    grpc_testing_empty_pb.Empty,
    (request: grpc_testing_empty_pb.Empty) => {
      return request.serializeBinary();
    },
    grpc_testing_empty_pb.Empty.deserializeBinary
  );

  setNotServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_empty_pb.Empty>;

  setNotServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void): grpcWeb.ClientReadableStream<grpc_testing_empty_pb.Empty>;

  setNotServing(
    request: grpc_testing_empty_pb.Empty,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_empty_pb.Empty) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.XdsUpdateHealthService/SetNotServing',
        request,
        metadata || {},
        this.methodDescriptorSetNotServing,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.XdsUpdateHealthService/SetNotServing',
    request,
    metadata || {},
    this.methodDescriptorSetNotServing);
  }

}

export class XdsUpdateClientConfigureServiceClient {
  client_: grpcWeb.AbstractClientBase;
  hostname_: string;
  credentials_: null | { [index: string]: string; };
  options_: null | { [index: string]: any; };

  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; }) {
    if (!options) options = {};
    if (!credentials) credentials = {};
    options['format'] = 'binary';

    this.client_ = new grpcWeb.GrpcWebClientBase(options);
    this.hostname_ = hostname;
    this.credentials_ = credentials;
    this.options_ = options;
  }

  methodDescriptorConfigure = new grpcWeb.MethodDescriptor(
    '/grpc.testing.XdsUpdateClientConfigureService/Configure',
    grpcWeb.MethodType.UNARY,
    grpc_testing_messages_pb.ClientConfigureRequest,
    grpc_testing_messages_pb.ClientConfigureResponse,
    (request: grpc_testing_messages_pb.ClientConfigureRequest) => {
      return request.serializeBinary();
    },
    grpc_testing_messages_pb.ClientConfigureResponse.deserializeBinary
  );

  configure(
    request: grpc_testing_messages_pb.ClientConfigureRequest,
    metadata: grpcWeb.Metadata | null): Promise<grpc_testing_messages_pb.ClientConfigureResponse>;

  configure(
    request: grpc_testing_messages_pb.ClientConfigureRequest,
    metadata: grpcWeb.Metadata | null,
    callback: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.ClientConfigureResponse) => void): grpcWeb.ClientReadableStream<grpc_testing_messages_pb.ClientConfigureResponse>;

  configure(
    request: grpc_testing_messages_pb.ClientConfigureRequest,
    metadata: grpcWeb.Metadata | null,
    callback?: (err: grpcWeb.RpcError,
               response: grpc_testing_messages_pb.ClientConfigureResponse) => void) {
    if (callback !== undefined) {
      return this.client_.rpcCall(
        this.hostname_ +
          '/grpc.testing.XdsUpdateClientConfigureService/Configure',
        request,
        metadata || {},
        this.methodDescriptorConfigure,
        callback);
    }
    return this.client_.unaryCall(
    this.hostname_ +
      '/grpc.testing.XdsUpdateClientConfigureService/Configure',
    request,
    metadata || {},
    this.methodDescriptorConfigure);
  }

}

