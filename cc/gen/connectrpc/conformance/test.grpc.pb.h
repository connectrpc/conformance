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

#ifndef GRPC_connectrpc_2fconformance_2ftest_2eproto__INCLUDED
#define GRPC_connectrpc_2fconformance_2ftest_2eproto__INCLUDED

#include "connectrpc/conformance/test.pb.h"

#include <functional>
#include <grpcpp/generic/async_generic_service.h>
#include <grpcpp/support/async_stream.h>
#include <grpcpp/support/async_unary_call.h>
#include <grpcpp/support/client_callback.h>
#include <grpcpp/client_context.h>
#include <grpcpp/completion_queue.h>
#include <grpcpp/support/message_allocator.h>
#include <grpcpp/support/method_handler.h>
#include <grpcpp/impl/proto_utils.h>
#include <grpcpp/impl/rpc_method.h>
#include <grpcpp/support/server_callback.h>
#include <grpcpp/impl/server_callback_handlers.h>
#include <grpcpp/server_context.h>
#include <grpcpp/impl/service_type.h>
#include <grpcpp/support/status.h>
#include <grpcpp/support/stub_options.h>
#include <grpcpp/support/sync_stream.h>

namespace connectrpc {
namespace conformance {

// A simple service to test the various types of RPCs and experiment with
// performance with various types of payload.
class TestService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.TestService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    // One empty request followed by one empty response.
    virtual ::grpc::Status EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncEmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncEmptyCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncEmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncEmptyCallRaw(context, request, cq));
    }
    // One request followed by one response.
    virtual ::grpc::Status UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> AsyncUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(AsyncUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncUnaryCallRaw(context, request, cq));
    }
    // One request followed by one response. This RPC always fails.
    virtual ::grpc::Status FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> AsyncFailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(AsyncFailUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncFailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncFailUnaryCallRaw(context, request, cq));
    }
    // One request followed by one response. Response has cache control
    // headers set such that a caching HTTP proxy (such as GFE) can
    // satisfy subsequent requests.
    virtual ::grpc::Status CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> AsyncCacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(AsyncCacheableUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncCacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncCacheableUnaryCallRaw(context, request, cq));
    }
    // One request followed by a sequence of responses (streamed download).
    // The server returns the payload with client desired type and sizes.
    std::unique_ptr< ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> StreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) {
      return std::unique_ptr< ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(StreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncStreamingOutputCallRaw(context, request, cq));
    }
    // One request followed by a sequence of responses (streamed download).
    // The server returns the payload with client desired type and sizes.
    // This RPC always responds with an error status.
    std::unique_ptr< ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> FailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) {
      return std::unique_ptr< ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(FailStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncFailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncFailStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncFailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncFailStreamingOutputCallRaw(context, request, cq));
    }
    // A sequence of requests followed by one response (streamed upload).
    // The server returns the aggregated size of client payload as the result.
    std::unique_ptr< ::grpc::ClientWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>> StreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response) {
      return std::unique_ptr< ::grpc::ClientWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>>(StreamingInputCallRaw(context, response));
    }
    std::unique_ptr< ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>> AsyncStreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>>(AsyncStreamingInputCallRaw(context, response, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>> PrepareAsyncStreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>>(PrepareAsyncStreamingInputCallRaw(context, response, cq));
    }
    // A sequence of requests with each request served by the server immediately.
    // As one request could lead to multiple responses, this interface
    // demonstrates the idea of full duplexing.
    std::unique_ptr< ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> FullDuplexCall(::grpc::ClientContext* context) {
      return std::unique_ptr< ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(FullDuplexCallRaw(context));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncFullDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncFullDuplexCallRaw(context, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncFullDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncFullDuplexCallRaw(context, cq));
    }
    // A sequence of requests followed by a sequence of responses.
    // The server buffers all the client requests and then serves them in order. A
    // stream of responses are returned to the client when the server starts with
    // first request.
    std::unique_ptr< ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> HalfDuplexCall(::grpc::ClientContext* context) {
      return std::unique_ptr< ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(HalfDuplexCallRaw(context));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncHalfDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncHalfDuplexCallRaw(context, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncHalfDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncHalfDuplexCallRaw(context, cq));
    }
    // The test server will not implement this method. It will be used
    // to test the behavior when clients call unimplemented methods.
    virtual ::grpc::Status UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedCallRaw(context, request, cq));
    }
    // The test server will not implement this method. It will be used
    // to test the behavior when clients call unimplemented streaming output methods.
    std::unique_ptr< ::grpc::ClientReaderInterface< ::google::protobuf::Empty>> UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) {
      return std::unique_ptr< ::grpc::ClientReaderInterface< ::google::protobuf::Empty>>(UnimplementedStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>> AsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>>(AsyncUnimplementedStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>> PrepareAsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedStreamingOutputCallRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      // One empty request followed by one empty response.
      virtual void EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // One request followed by one response.
      virtual void UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // One request followed by one response. This RPC always fails.
      virtual void FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // One request followed by one response. Response has cache control
      // headers set such that a caching HTTP proxy (such as GFE) can
      // satisfy subsequent requests.
      virtual void CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // One request followed by a sequence of responses (streamed download).
      // The server returns the payload with client desired type and sizes.
      virtual void StreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ClientReadReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* reactor) = 0;
      // One request followed by a sequence of responses (streamed download).
      // The server returns the payload with client desired type and sizes.
      // This RPC always responds with an error status.
      virtual void FailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ClientReadReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* reactor) = 0;
      // A sequence of requests followed by one response (streamed upload).
      // The server returns the aggregated size of client payload as the result.
      virtual void StreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::ClientWriteReactor< ::connectrpc::conformance::StreamingInputCallRequest>* reactor) = 0;
      // A sequence of requests with each request served by the server immediately.
      // As one request could lead to multiple responses, this interface
      // demonstrates the idea of full duplexing.
      virtual void FullDuplexCall(::grpc::ClientContext* context, ::grpc::ClientBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* reactor) = 0;
      // A sequence of requests followed by a sequence of responses.
      // The server buffers all the client requests and then serves them in order. A
      // stream of responses are returned to the client when the server starts with
      // first request.
      virtual void HalfDuplexCall(::grpc::ClientContext* context, ::grpc::ClientBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* reactor) = 0;
      // The test server will not implement this method. It will be used
      // to test the behavior when clients call unimplemented methods.
      virtual void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // The test server will not implement this method. It will be used
      // to test the behavior when clients call unimplemented streaming output methods.
      virtual void UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::grpc::ClientReadReactor< ::google::protobuf::Empty>* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncEmptyCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncEmptyCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* AsyncUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* AsyncFailUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncFailUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* AsyncCacheableUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncCacheableUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* StreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* FailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncFailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncFailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>* StreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response) = 0;
    virtual ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>* AsyncStreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncWriterInterface< ::connectrpc::conformance::StreamingInputCallRequest>* PrepareAsyncStreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* FullDuplexCallRaw(::grpc::ClientContext* context) = 0;
    virtual ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncFullDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncFullDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* HalfDuplexCallRaw(::grpc::ClientContext* context) = 0;
    virtual ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncHalfDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderWriterInterface< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncHalfDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderInterface< ::google::protobuf::Empty>* UnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>* AsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>* PrepareAsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncEmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncEmptyCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncEmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncEmptyCallRaw(context, request, cq));
    }
    ::grpc::Status UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> AsyncUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(AsyncUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncUnaryCallRaw(context, request, cq));
    }
    ::grpc::Status FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> AsyncFailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(AsyncFailUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncFailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncFailUnaryCallRaw(context, request, cq));
    }
    ::grpc::Status CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::connectrpc::conformance::SimpleResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> AsyncCacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(AsyncCacheableUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>> PrepareAsyncCacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>>(PrepareAsyncCacheableUnaryCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>> StreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) {
      return std::unique_ptr< ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(StreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncStreamingOutputCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>> FailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) {
      return std::unique_ptr< ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(FailStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncFailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncFailStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncFailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncFailStreamingOutputCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientWriter< ::connectrpc::conformance::StreamingInputCallRequest>> StreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response) {
      return std::unique_ptr< ::grpc::ClientWriter< ::connectrpc::conformance::StreamingInputCallRequest>>(StreamingInputCallRaw(context, response));
    }
    std::unique_ptr< ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>> AsyncStreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>>(AsyncStreamingInputCallRaw(context, response, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>> PrepareAsyncStreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>>(PrepareAsyncStreamingInputCallRaw(context, response, cq));
    }
    std::unique_ptr< ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> FullDuplexCall(::grpc::ClientContext* context) {
      return std::unique_ptr< ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(FullDuplexCallRaw(context));
    }
    std::unique_ptr<  ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncFullDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncFullDuplexCallRaw(context, cq, tag));
    }
    std::unique_ptr<  ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncFullDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncFullDuplexCallRaw(context, cq));
    }
    std::unique_ptr< ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> HalfDuplexCall(::grpc::ClientContext* context) {
      return std::unique_ptr< ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(HalfDuplexCallRaw(context));
    }
    std::unique_ptr<  ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> AsyncHalfDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(AsyncHalfDuplexCallRaw(context, cq, tag));
    }
    std::unique_ptr<  ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>> PrepareAsyncHalfDuplexCall(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>>(PrepareAsyncHalfDuplexCallRaw(context, cq));
    }
    ::grpc::Status UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientReader< ::google::protobuf::Empty>> UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) {
      return std::unique_ptr< ::grpc::ClientReader< ::google::protobuf::Empty>>(UnimplementedStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>> AsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>>(AsyncUnimplementedStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>> PrepareAsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedStreamingOutputCallRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void EmptyCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
      void UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) override;
      void UnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
      void FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) override;
      void FailUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
      void CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, std::function<void(::grpc::Status)>) override;
      void CacheableUnaryCall(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
      void StreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ClientReadReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* reactor) override;
      void FailStreamingOutputCall(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ClientReadReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* reactor) override;
      void StreamingInputCall(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::ClientWriteReactor< ::connectrpc::conformance::StreamingInputCallRequest>* reactor) override;
      void FullDuplexCall(::grpc::ClientContext* context, ::grpc::ClientBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* reactor) override;
      void HalfDuplexCall(::grpc::ClientContext* context, ::grpc::ClientBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* reactor) override;
      void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
      void UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::grpc::ClientReadReactor< ::google::protobuf::Empty>* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncEmptyCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncEmptyCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* AsyncUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* AsyncFailUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncFailUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* AsyncCacheableUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::SimpleResponse>* PrepareAsyncCacheableUnaryCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::SimpleRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>* StreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) override;
    ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReader< ::connectrpc::conformance::StreamingOutputCallResponse>* FailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request) override;
    ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncFailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReader< ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncFailStreamingOutputCallRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientWriter< ::connectrpc::conformance::StreamingInputCallRequest>* StreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response) override;
    ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>* AsyncStreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncWriter< ::connectrpc::conformance::StreamingInputCallRequest>* PrepareAsyncStreamingInputCallRaw(::grpc::ClientContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* FullDuplexCallRaw(::grpc::ClientContext* context) override;
    ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncFullDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncFullDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* HalfDuplexCallRaw(::grpc::ClientContext* context) override;
    ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* AsyncHalfDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* PrepareAsyncHalfDuplexCallRaw(::grpc::ClientContext* context, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReader< ::google::protobuf::Empty>* UnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) override;
    ::grpc::ClientAsyncReader< ::google::protobuf::Empty>* AsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReader< ::google::protobuf::Empty>* PrepareAsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_EmptyCall_;
    const ::grpc::internal::RpcMethod rpcmethod_UnaryCall_;
    const ::grpc::internal::RpcMethod rpcmethod_FailUnaryCall_;
    const ::grpc::internal::RpcMethod rpcmethod_CacheableUnaryCall_;
    const ::grpc::internal::RpcMethod rpcmethod_StreamingOutputCall_;
    const ::grpc::internal::RpcMethod rpcmethod_FailStreamingOutputCall_;
    const ::grpc::internal::RpcMethod rpcmethod_StreamingInputCall_;
    const ::grpc::internal::RpcMethod rpcmethod_FullDuplexCall_;
    const ::grpc::internal::RpcMethod rpcmethod_HalfDuplexCall_;
    const ::grpc::internal::RpcMethod rpcmethod_UnimplementedCall_;
    const ::grpc::internal::RpcMethod rpcmethod_UnimplementedStreamingOutputCall_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    // One empty request followed by one empty response.
    virtual ::grpc::Status EmptyCall(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response);
    // One request followed by one response.
    virtual ::grpc::Status UnaryCall(::grpc::ServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response);
    // One request followed by one response. This RPC always fails.
    virtual ::grpc::Status FailUnaryCall(::grpc::ServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response);
    // One request followed by one response. Response has cache control
    // headers set such that a caching HTTP proxy (such as GFE) can
    // satisfy subsequent requests.
    virtual ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response);
    // One request followed by a sequence of responses (streamed download).
    // The server returns the payload with client desired type and sizes.
    virtual ::grpc::Status StreamingOutputCall(::grpc::ServerContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* writer);
    // One request followed by a sequence of responses (streamed download).
    // The server returns the payload with client desired type and sizes.
    // This RPC always responds with an error status.
    virtual ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* writer);
    // A sequence of requests followed by one response (streamed upload).
    // The server returns the aggregated size of client payload as the result.
    virtual ::grpc::Status StreamingInputCall(::grpc::ServerContext* context, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* reader, ::connectrpc::conformance::StreamingInputCallResponse* response);
    // A sequence of requests with each request served by the server immediately.
    // As one request could lead to multiple responses, this interface
    // demonstrates the idea of full duplexing.
    virtual ::grpc::Status FullDuplexCall(::grpc::ServerContext* context, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* stream);
    // A sequence of requests followed by a sequence of responses.
    // The server buffers all the client requests and then serves them in order. A
    // stream of responses are returned to the client when the server starts with
    // first request.
    virtual ::grpc::Status HalfDuplexCall(::grpc::ServerContext* context, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* stream);
    // The test server will not implement this method. It will be used
    // to test the behavior when clients call unimplemented methods.
    virtual ::grpc::Status UnimplementedCall(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response);
    // The test server will not implement this method. It will be used
    // to test the behavior when clients call unimplemented streaming output methods.
    virtual ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::grpc::ServerWriter< ::google::protobuf::Empty>* writer);
  };
  template <class BaseClass>
  class WithAsyncMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_EmptyCall() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestEmptyCall(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_UnaryCall() {
      ::grpc::Service::MarkMethodAsync(1);
    }
    ~WithAsyncMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnaryCall(::grpc::ServerContext* context, ::connectrpc::conformance::SimpleRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::SimpleResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodAsync(2);
    }
    ~WithAsyncMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFailUnaryCall(::grpc::ServerContext* context, ::connectrpc::conformance::SimpleRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::SimpleResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(2, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodAsync(3);
    }
    ~WithAsyncMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestCacheableUnaryCall(::grpc::ServerContext* context, ::connectrpc::conformance::SimpleRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::SimpleResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(3, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodAsync(4);
    }
    ~WithAsyncMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStreamingOutputCall(::grpc::ServerContext* context, ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ServerAsyncWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(4, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodAsync(5);
    }
    ~WithAsyncMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFailStreamingOutputCall(::grpc::ServerContext* context, ::connectrpc::conformance::StreamingOutputCallRequest* request, ::grpc::ServerAsyncWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(5, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_StreamingInputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_StreamingInputCall() {
      ::grpc::Service::MarkMethodAsync(6);
    }
    ~WithAsyncMethod_StreamingInputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingInputCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* /*reader*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStreamingInputCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReader< ::connectrpc::conformance::StreamingInputCallResponse, ::connectrpc::conformance::StreamingInputCallRequest>* reader, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncClientStreaming(6, context, reader, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_FullDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_FullDuplexCall() {
      ::grpc::Service::MarkMethodAsync(7);
    }
    ~WithAsyncMethod_FullDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FullDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFullDuplexCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* stream, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncBidiStreaming(7, context, stream, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_HalfDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_HalfDuplexCall() {
      ::grpc::Service::MarkMethodAsync(8);
    }
    ~WithAsyncMethod_HalfDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status HalfDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestHalfDuplexCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* stream, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncBidiStreaming(8, context, stream, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodAsync(9);
    }
    ~WithAsyncMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedCall(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(9, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodAsync(10);
    }
    ~WithAsyncMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncWriter< ::google::protobuf::Empty>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(10, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_EmptyCall<WithAsyncMethod_UnaryCall<WithAsyncMethod_FailUnaryCall<WithAsyncMethod_CacheableUnaryCall<WithAsyncMethod_StreamingOutputCall<WithAsyncMethod_FailStreamingOutputCall<WithAsyncMethod_StreamingInputCall<WithAsyncMethod_FullDuplexCall<WithAsyncMethod_HalfDuplexCall<WithAsyncMethod_UnimplementedCall<WithAsyncMethod_UnimplementedStreamingOutputCall<Service > > > > > > > > > > > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_EmptyCall() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response) { return this->EmptyCall(context, request, response); }));}
    void SetMessageAllocatorFor_EmptyCall(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* EmptyCall(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_UnaryCall() {
      ::grpc::Service::MarkMethodCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response) { return this->UnaryCall(context, request, response); }));}
    void SetMessageAllocatorFor_UnaryCall(
        ::grpc::MessageAllocator< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(1);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodCallback(2,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response) { return this->FailUnaryCall(context, request, response); }));}
    void SetMessageAllocatorFor_FailUnaryCall(
        ::grpc::MessageAllocator< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(2);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* FailUnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodCallback(3,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::SimpleRequest* request, ::connectrpc::conformance::SimpleResponse* response) { return this->CacheableUnaryCall(context, request, response); }));}
    void SetMessageAllocatorFor_CacheableUnaryCall(
        ::grpc::MessageAllocator< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(3);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* CacheableUnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodCallback(4,
          new ::grpc::internal::CallbackServerStreamingHandler< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request) { return this->StreamingOutputCall(context, request); }));
    }
    ~WithCallbackMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* StreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodCallback(5,
          new ::grpc::internal::CallbackServerStreamingHandler< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::StreamingOutputCallRequest* request) { return this->FailStreamingOutputCall(context, request); }));
    }
    ~WithCallbackMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::connectrpc::conformance::StreamingOutputCallResponse>* FailStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_StreamingInputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_StreamingInputCall() {
      ::grpc::Service::MarkMethodCallback(6,
          new ::grpc::internal::CallbackClientStreamingHandler< ::connectrpc::conformance::StreamingInputCallRequest, ::connectrpc::conformance::StreamingInputCallResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, ::connectrpc::conformance::StreamingInputCallResponse* response) { return this->StreamingInputCall(context, response); }));
    }
    ~WithCallbackMethod_StreamingInputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingInputCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* /*reader*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerReadReactor< ::connectrpc::conformance::StreamingInputCallRequest>* StreamingInputCall(
      ::grpc::CallbackServerContext* /*context*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_FullDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_FullDuplexCall() {
      ::grpc::Service::MarkMethodCallback(7,
          new ::grpc::internal::CallbackBidiHandler< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](
                   ::grpc::CallbackServerContext* context) { return this->FullDuplexCall(context); }));
    }
    ~WithCallbackMethod_FullDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FullDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* FullDuplexCall(
      ::grpc::CallbackServerContext* /*context*/)
      { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_HalfDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_HalfDuplexCall() {
      ::grpc::Service::MarkMethodCallback(8,
          new ::grpc::internal::CallbackBidiHandler< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](
                   ::grpc::CallbackServerContext* context) { return this->HalfDuplexCall(context); }));
    }
    ~WithCallbackMethod_HalfDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status HalfDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerBidiReactor< ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* HalfDuplexCall(
      ::grpc::CallbackServerContext* /*context*/)
      { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodCallback(9,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response) { return this->UnimplementedCall(context, request, response); }));}
    void SetMessageAllocatorFor_UnimplementedCall(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(9);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnimplementedCall(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodCallback(10,
          new ::grpc::internal::CallbackServerStreamingHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request) { return this->UnimplementedStreamingOutputCall(context, request); }));
    }
    ~WithCallbackMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::google::protobuf::Empty>* UnimplementedStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_EmptyCall<WithCallbackMethod_UnaryCall<WithCallbackMethod_FailUnaryCall<WithCallbackMethod_CacheableUnaryCall<WithCallbackMethod_StreamingOutputCall<WithCallbackMethod_FailStreamingOutputCall<WithCallbackMethod_StreamingInputCall<WithCallbackMethod_FullDuplexCall<WithCallbackMethod_HalfDuplexCall<WithCallbackMethod_UnimplementedCall<WithCallbackMethod_UnimplementedStreamingOutputCall<Service > > > > > > > > > > > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_EmptyCall() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_UnaryCall() {
      ::grpc::Service::MarkMethodGeneric(1);
    }
    ~WithGenericMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodGeneric(2);
    }
    ~WithGenericMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodGeneric(3);
    }
    ~WithGenericMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodGeneric(4);
    }
    ~WithGenericMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodGeneric(5);
    }
    ~WithGenericMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_StreamingInputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_StreamingInputCall() {
      ::grpc::Service::MarkMethodGeneric(6);
    }
    ~WithGenericMethod_StreamingInputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingInputCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* /*reader*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_FullDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_FullDuplexCall() {
      ::grpc::Service::MarkMethodGeneric(7);
    }
    ~WithGenericMethod_FullDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FullDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_HalfDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_HalfDuplexCall() {
      ::grpc::Service::MarkMethodGeneric(8);
    }
    ~WithGenericMethod_HalfDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status HalfDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodGeneric(9);
    }
    ~WithGenericMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodGeneric(10);
    }
    ~WithGenericMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_EmptyCall() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestEmptyCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_UnaryCall() {
      ::grpc::Service::MarkMethodRaw(1);
    }
    ~WithRawMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnaryCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodRaw(2);
    }
    ~WithRawMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFailUnaryCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(2, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodRaw(3);
    }
    ~WithRawMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestCacheableUnaryCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(3, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodRaw(4);
    }
    ~WithRawMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncWriter< ::grpc::ByteBuffer>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(4, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodRaw(5);
    }
    ~WithRawMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFailStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncWriter< ::grpc::ByteBuffer>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(5, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_StreamingInputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_StreamingInputCall() {
      ::grpc::Service::MarkMethodRaw(6);
    }
    ~WithRawMethod_StreamingInputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingInputCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* /*reader*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStreamingInputCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReader< ::grpc::ByteBuffer, ::grpc::ByteBuffer>* reader, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncClientStreaming(6, context, reader, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_FullDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_FullDuplexCall() {
      ::grpc::Service::MarkMethodRaw(7);
    }
    ~WithRawMethod_FullDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FullDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestFullDuplexCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReaderWriter< ::grpc::ByteBuffer, ::grpc::ByteBuffer>* stream, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncBidiStreaming(7, context, stream, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_HalfDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_HalfDuplexCall() {
      ::grpc::Service::MarkMethodRaw(8);
    }
    ~WithRawMethod_HalfDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status HalfDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestHalfDuplexCall(::grpc::ServerContext* context, ::grpc::ServerAsyncReaderWriter< ::grpc::ByteBuffer, ::grpc::ByteBuffer>* stream, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncBidiStreaming(8, context, stream, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodRaw(9);
    }
    ~WithRawMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(9, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodRaw(10);
    }
    ~WithRawMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncWriter< ::grpc::ByteBuffer>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(10, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_EmptyCall() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->EmptyCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* EmptyCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_UnaryCall() {
      ::grpc::Service::MarkMethodRawCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->UnaryCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodRawCallback(2,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->FailUnaryCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* FailUnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodRawCallback(3,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->CacheableUnaryCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* CacheableUnaryCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodRawCallback(4,
          new ::grpc::internal::CallbackServerStreamingHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const::grpc::ByteBuffer* request) { return this->StreamingOutputCall(context, request); }));
    }
    ~WithRawCallbackMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::grpc::ByteBuffer>* StreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodRawCallback(5,
          new ::grpc::internal::CallbackServerStreamingHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const::grpc::ByteBuffer* request) { return this->FailStreamingOutputCall(context, request); }));
    }
    ~WithRawCallbackMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::grpc::ByteBuffer>* FailStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_StreamingInputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_StreamingInputCall() {
      ::grpc::Service::MarkMethodRawCallback(6,
          new ::grpc::internal::CallbackClientStreamingHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, ::grpc::ByteBuffer* response) { return this->StreamingInputCall(context, response); }));
    }
    ~WithRawCallbackMethod_StreamingInputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status StreamingInputCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReader< ::connectrpc::conformance::StreamingInputCallRequest>* /*reader*/, ::connectrpc::conformance::StreamingInputCallResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerReadReactor< ::grpc::ByteBuffer>* StreamingInputCall(
      ::grpc::CallbackServerContext* /*context*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_FullDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_FullDuplexCall() {
      ::grpc::Service::MarkMethodRawCallback(7,
          new ::grpc::internal::CallbackBidiHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context) { return this->FullDuplexCall(context); }));
    }
    ~WithRawCallbackMethod_FullDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status FullDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerBidiReactor< ::grpc::ByteBuffer, ::grpc::ByteBuffer>* FullDuplexCall(
      ::grpc::CallbackServerContext* /*context*/)
      { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_HalfDuplexCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_HalfDuplexCall() {
      ::grpc::Service::MarkMethodRawCallback(8,
          new ::grpc::internal::CallbackBidiHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context) { return this->HalfDuplexCall(context); }));
    }
    ~WithRawCallbackMethod_HalfDuplexCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status HalfDuplexCall(::grpc::ServerContext* /*context*/, ::grpc::ServerReaderWriter< ::connectrpc::conformance::StreamingOutputCallResponse, ::connectrpc::conformance::StreamingOutputCallRequest>* /*stream*/)  override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerBidiReactor< ::grpc::ByteBuffer, ::grpc::ByteBuffer>* HalfDuplexCall(
      ::grpc::CallbackServerContext* /*context*/)
      { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodRawCallback(9,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->UnimplementedCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnimplementedCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodRawCallback(10,
          new ::grpc::internal::CallbackServerStreamingHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const::grpc::ByteBuffer* request) { return this->UnimplementedStreamingOutputCall(context, request); }));
    }
    ~WithRawCallbackMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::grpc::ByteBuffer>* UnimplementedStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_EmptyCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_EmptyCall() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedEmptyCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_EmptyCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status EmptyCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedEmptyCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_UnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_UnaryCall() {
      ::grpc::Service::MarkMethodStreamed(1,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* streamer) {
                       return this->StreamedUnaryCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_UnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status UnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedUnaryCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::SimpleRequest,::connectrpc::conformance::SimpleResponse>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_FailUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_FailUnaryCall() {
      ::grpc::Service::MarkMethodStreamed(2,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* streamer) {
                       return this->StreamedFailUnaryCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_FailUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status FailUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedFailUnaryCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::SimpleRequest,::connectrpc::conformance::SimpleResponse>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_CacheableUnaryCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_CacheableUnaryCall() {
      ::grpc::Service::MarkMethodStreamed(3,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::SimpleRequest, ::connectrpc::conformance::SimpleResponse>* streamer) {
                       return this->StreamedCacheableUnaryCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_CacheableUnaryCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status CacheableUnaryCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::SimpleRequest* /*request*/, ::connectrpc::conformance::SimpleResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedCacheableUnaryCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::SimpleRequest,::connectrpc::conformance::SimpleResponse>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodStreamed(9,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedUnimplementedCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedUnimplementedCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_EmptyCall<WithStreamedUnaryMethod_UnaryCall<WithStreamedUnaryMethod_FailUnaryCall<WithStreamedUnaryMethod_CacheableUnaryCall<WithStreamedUnaryMethod_UnimplementedCall<Service > > > > > StreamedUnaryService;
  template <class BaseClass>
  class WithSplitStreamingMethod_StreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithSplitStreamingMethod_StreamingOutputCall() {
      ::grpc::Service::MarkMethodStreamed(4,
        new ::grpc::internal::SplitServerStreamingHandler<
          ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerSplitStreamer<
                     ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* streamer) {
                       return this->StreamedStreamingOutputCall(context,
                         streamer);
                  }));
    }
    ~WithSplitStreamingMethod_StreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status StreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with split streamed
    virtual ::grpc::Status StreamedStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ServerSplitStreamer< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* server_split_streamer) = 0;
  };
  template <class BaseClass>
  class WithSplitStreamingMethod_FailStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithSplitStreamingMethod_FailStreamingOutputCall() {
      ::grpc::Service::MarkMethodStreamed(5,
        new ::grpc::internal::SplitServerStreamingHandler<
          ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerSplitStreamer<
                     ::connectrpc::conformance::StreamingOutputCallRequest, ::connectrpc::conformance::StreamingOutputCallResponse>* streamer) {
                       return this->StreamedFailStreamingOutputCall(context,
                         streamer);
                  }));
    }
    ~WithSplitStreamingMethod_FailStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status FailStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::StreamingOutputCallRequest* /*request*/, ::grpc::ServerWriter< ::connectrpc::conformance::StreamingOutputCallResponse>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with split streamed
    virtual ::grpc::Status StreamedFailStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ServerSplitStreamer< ::connectrpc::conformance::StreamingOutputCallRequest,::connectrpc::conformance::StreamingOutputCallResponse>* server_split_streamer) = 0;
  };
  template <class BaseClass>
  class WithSplitStreamingMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithSplitStreamingMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodStreamed(10,
        new ::grpc::internal::SplitServerStreamingHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerSplitStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedUnimplementedStreamingOutputCall(context,
                         streamer);
                  }));
    }
    ~WithSplitStreamingMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with split streamed
    virtual ::grpc::Status StreamedUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ServerSplitStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_split_streamer) = 0;
  };
  typedef WithSplitStreamingMethod_StreamingOutputCall<WithSplitStreamingMethod_FailStreamingOutputCall<WithSplitStreamingMethod_UnimplementedStreamingOutputCall<Service > > > SplitStreamedService;
  typedef WithStreamedUnaryMethod_EmptyCall<WithStreamedUnaryMethod_UnaryCall<WithStreamedUnaryMethod_FailUnaryCall<WithStreamedUnaryMethod_CacheableUnaryCall<WithSplitStreamingMethod_StreamingOutputCall<WithSplitStreamingMethod_FailStreamingOutputCall<WithStreamedUnaryMethod_UnimplementedCall<WithSplitStreamingMethod_UnimplementedStreamingOutputCall<Service > > > > > > > > StreamedService;
};

// A simple service NOT implemented at servers so clients can test for
// that case.
class UnimplementedService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.UnimplementedService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    // A call that no server should implement
    virtual ::grpc::Status UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedCallRaw(context, request, cq));
    }
    // A call that no server should implement
    std::unique_ptr< ::grpc::ClientReaderInterface< ::google::protobuf::Empty>> UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) {
      return std::unique_ptr< ::grpc::ClientReaderInterface< ::google::protobuf::Empty>>(UnimplementedStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>> AsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>>(AsyncUnimplementedStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>> PrepareAsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedStreamingOutputCallRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      // A call that no server should implement
      virtual void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // A call that no server should implement
      virtual void UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::grpc::ClientReadReactor< ::google::protobuf::Empty>* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientReaderInterface< ::google::protobuf::Empty>* UnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>* AsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::google::protobuf::Empty>* PrepareAsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncUnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedCallRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientReader< ::google::protobuf::Empty>> UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) {
      return std::unique_ptr< ::grpc::ClientReader< ::google::protobuf::Empty>>(UnimplementedStreamingOutputCallRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>> AsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>>(AsyncUnimplementedStreamingOutputCallRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>> PrepareAsyncUnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::google::protobuf::Empty>>(PrepareAsyncUnimplementedStreamingOutputCallRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void UnimplementedCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
      void UnimplementedStreamingOutputCall(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::grpc::ClientReadReactor< ::google::protobuf::Empty>* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncUnimplementedCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientReader< ::google::protobuf::Empty>* UnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request) override;
    ::grpc::ClientAsyncReader< ::google::protobuf::Empty>* AsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReader< ::google::protobuf::Empty>* PrepareAsyncUnimplementedStreamingOutputCallRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_UnimplementedCall_;
    const ::grpc::internal::RpcMethod rpcmethod_UnimplementedStreamingOutputCall_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    // A call that no server should implement
    virtual ::grpc::Status UnimplementedCall(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response);
    // A call that no server should implement
    virtual ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::grpc::ServerWriter< ::google::protobuf::Empty>* writer);
  };
  template <class BaseClass>
  class WithAsyncMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedCall(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodAsync(1);
    }
    ~WithAsyncMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncWriter< ::google::protobuf::Empty>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(1, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_UnimplementedCall<WithAsyncMethod_UnimplementedStreamingOutputCall<Service > > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response) { return this->UnimplementedCall(context, request, response); }));}
    void SetMessageAllocatorFor_UnimplementedCall(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnimplementedCall(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodCallback(1,
          new ::grpc::internal::CallbackServerStreamingHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request) { return this->UnimplementedStreamingOutputCall(context, request); }));
    }
    ~WithCallbackMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::google::protobuf::Empty>* UnimplementedStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_UnimplementedCall<WithCallbackMethod_UnimplementedStreamingOutputCall<Service > > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodGeneric(1);
    }
    ~WithGenericMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodRaw(1);
    }
    ~WithRawMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncWriter< ::grpc::ByteBuffer>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(1, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->UnimplementedCall(context, request, response); }));
    }
    ~WithRawCallbackMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* UnimplementedCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodRawCallback(1,
          new ::grpc::internal::CallbackServerStreamingHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const::grpc::ByteBuffer* request) { return this->UnimplementedStreamingOutputCall(context, request); }));
    }
    ~WithRawCallbackMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerWriteReactor< ::grpc::ByteBuffer>* UnimplementedStreamingOutputCall(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_UnimplementedCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_UnimplementedCall() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedUnimplementedCall(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_UnimplementedCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status UnimplementedCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedUnimplementedCall(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_UnimplementedCall<Service > StreamedUnaryService;
  template <class BaseClass>
  class WithSplitStreamingMethod_UnimplementedStreamingOutputCall : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithSplitStreamingMethod_UnimplementedStreamingOutputCall() {
      ::grpc::Service::MarkMethodStreamed(1,
        new ::grpc::internal::SplitServerStreamingHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerSplitStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedUnimplementedStreamingOutputCall(context,
                         streamer);
                  }));
    }
    ~WithSplitStreamingMethod_UnimplementedStreamingOutputCall() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status UnimplementedStreamingOutputCall(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::grpc::ServerWriter< ::google::protobuf::Empty>* /*writer*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with split streamed
    virtual ::grpc::Status StreamedUnimplementedStreamingOutputCall(::grpc::ServerContext* context, ::grpc::ServerSplitStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_split_streamer) = 0;
  };
  typedef WithSplitStreamingMethod_UnimplementedStreamingOutputCall<Service > SplitStreamedService;
  typedef WithStreamedUnaryMethod_UnimplementedCall<WithSplitStreamingMethod_UnimplementedStreamingOutputCall<Service > > StreamedService;
};

// A service used to control reconnect server.
class ReconnectService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.ReconnectService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    virtual ::grpc::Status Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncStart(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncStartRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncStart(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncStartRaw(context, request, cq));
    }
    virtual ::grpc::Status Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::connectrpc::conformance::ReconnectInfo* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>> AsyncStop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>>(AsyncStopRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>> PrepareAsyncStop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>>(PrepareAsyncStopRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      virtual void Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      virtual void Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response, std::function<void(::grpc::Status)>) = 0;
      virtual void Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response, ::grpc::ClientUnaryReactor* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncStartRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncStartRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>* AsyncStopRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ReconnectInfo>* PrepareAsyncStopRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncStart(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncStartRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncStart(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncStartRaw(context, request, cq));
    }
    ::grpc::Status Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::connectrpc::conformance::ReconnectInfo* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>> AsyncStop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>>(AsyncStopRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>> PrepareAsyncStop(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>>(PrepareAsyncStopRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void Start(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
      void Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response, std::function<void(::grpc::Status)>) override;
      void Stop(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response, ::grpc::ClientUnaryReactor* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncStartRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncStartRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ReconnectParams& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>* AsyncStopRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ReconnectInfo>* PrepareAsyncStopRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_Start_;
    const ::grpc::internal::RpcMethod rpcmethod_Stop_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    virtual ::grpc::Status Start(::grpc::ServerContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response);
    virtual ::grpc::Status Stop(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response);
  };
  template <class BaseClass>
  class WithAsyncMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_Start() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStart(::grpc::ServerContext* context, ::connectrpc::conformance::ReconnectParams* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_Stop() {
      ::grpc::Service::MarkMethodAsync(1);
    }
    ~WithAsyncMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStop(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::ReconnectInfo>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_Start<WithAsyncMethod_Stop<Service > > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_Start() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::ReconnectParams, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::ReconnectParams* request, ::google::protobuf::Empty* response) { return this->Start(context, request, response); }));}
    void SetMessageAllocatorFor_Start(
        ::grpc::MessageAllocator< ::connectrpc::conformance::ReconnectParams, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::ReconnectParams, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Start(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_Stop() {
      ::grpc::Service::MarkMethodCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::connectrpc::conformance::ReconnectInfo>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::connectrpc::conformance::ReconnectInfo* response) { return this->Stop(context, request, response); }));}
    void SetMessageAllocatorFor_Stop(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::connectrpc::conformance::ReconnectInfo>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(1);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::connectrpc::conformance::ReconnectInfo>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Stop(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_Start<WithCallbackMethod_Stop<Service > > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_Start() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_Stop() {
      ::grpc::Service::MarkMethodGeneric(1);
    }
    ~WithGenericMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_Start() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStart(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_Stop() {
      ::grpc::Service::MarkMethodRaw(1);
    }
    ~WithRawMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestStop(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_Start() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->Start(context, request, response); }));
    }
    ~WithRawCallbackMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Start(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_Stop() {
      ::grpc::Service::MarkMethodRawCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->Stop(context, request, response); }));
    }
    ~WithRawCallbackMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Stop(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_Start : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_Start() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::ReconnectParams, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::ReconnectParams, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedStart(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_Start() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status Start(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ReconnectParams* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedStart(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::ReconnectParams,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_Stop : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_Stop() {
      ::grpc::Service::MarkMethodStreamed(1,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::connectrpc::conformance::ReconnectInfo>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::connectrpc::conformance::ReconnectInfo>* streamer) {
                       return this->StreamedStop(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_Stop() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status Stop(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::connectrpc::conformance::ReconnectInfo* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedStop(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::connectrpc::conformance::ReconnectInfo>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_Start<WithStreamedUnaryMethod_Stop<Service > > StreamedUnaryService;
  typedef Service SplitStreamedService;
  typedef WithStreamedUnaryMethod_Start<WithStreamedUnaryMethod_Stop<Service > > StreamedService;
};

// A service used to obtain stats for verifying LB behavior.
class LoadBalancerStatsService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.LoadBalancerStatsService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    // Gets the backend distribution for RPCs sent by a test client.
    virtual ::grpc::Status GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::connectrpc::conformance::LoadBalancerStatsResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>> AsyncGetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>>(AsyncGetClientStatsRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>> PrepareAsyncGetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>>(PrepareAsyncGetClientStatsRaw(context, request, cq));
    }
    // Gets the accumulated stats for RPCs sent by a test client.
    virtual ::grpc::Status GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>> AsyncGetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>>(AsyncGetClientAccumulatedStatsRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>> PrepareAsyncGetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>>(PrepareAsyncGetClientAccumulatedStatsRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      // Gets the backend distribution for RPCs sent by a test client.
      virtual void GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      // Gets the accumulated stats for RPCs sent by a test client.
      virtual void GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>* AsyncGetClientStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerStatsResponse>* PrepareAsyncGetClientStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* AsyncGetClientAccumulatedStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* PrepareAsyncGetClientAccumulatedStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::connectrpc::conformance::LoadBalancerStatsResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>> AsyncGetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>>(AsyncGetClientStatsRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>> PrepareAsyncGetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>>(PrepareAsyncGetClientStatsRaw(context, request, cq));
    }
    ::grpc::Status GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>> AsyncGetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>>(AsyncGetClientAccumulatedStatsRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>> PrepareAsyncGetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>>(PrepareAsyncGetClientAccumulatedStatsRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response, std::function<void(::grpc::Status)>) override;
      void GetClientStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
      void GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response, std::function<void(::grpc::Status)>) override;
      void GetClientAccumulatedStats(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>* AsyncGetClientStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerStatsResponse>* PrepareAsyncGetClientStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* AsyncGetClientAccumulatedStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* PrepareAsyncGetClientAccumulatedStatsRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_GetClientStats_;
    const ::grpc::internal::RpcMethod rpcmethod_GetClientAccumulatedStats_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    // Gets the backend distribution for RPCs sent by a test client.
    virtual ::grpc::Status GetClientStats(::grpc::ServerContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response);
    // Gets the accumulated stats for RPCs sent by a test client.
    virtual ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response);
  };
  template <class BaseClass>
  class WithAsyncMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_GetClientStats() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetClientStats(::grpc::ServerContext* context, ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::LoadBalancerStatsResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodAsync(1);
    }
    ~WithAsyncMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetClientAccumulatedStats(::grpc::ServerContext* context, ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_GetClientStats<WithAsyncMethod_GetClientAccumulatedStats<Service > > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_GetClientStats() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::LoadBalancerStatsRequest, ::connectrpc::conformance::LoadBalancerStatsResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::LoadBalancerStatsRequest* request, ::connectrpc::conformance::LoadBalancerStatsResponse* response) { return this->GetClientStats(context, request, response); }));}
    void SetMessageAllocatorFor_GetClientStats(
        ::grpc::MessageAllocator< ::connectrpc::conformance::LoadBalancerStatsRequest, ::connectrpc::conformance::LoadBalancerStatsResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::LoadBalancerStatsRequest, ::connectrpc::conformance::LoadBalancerStatsResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* GetClientStats(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* request, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* response) { return this->GetClientAccumulatedStats(context, request, response); }));}
    void SetMessageAllocatorFor_GetClientAccumulatedStats(
        ::grpc::MessageAllocator< ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(1);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* GetClientAccumulatedStats(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_GetClientStats<WithCallbackMethod_GetClientAccumulatedStats<Service > > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_GetClientStats() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodGeneric(1);
    }
    ~WithGenericMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_GetClientStats() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetClientStats(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodRaw(1);
    }
    ~WithRawMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetClientAccumulatedStats(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_GetClientStats() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->GetClientStats(context, request, response); }));
    }
    ~WithRawCallbackMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* GetClientStats(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodRawCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->GetClientAccumulatedStats(context, request, response); }));
    }
    ~WithRawCallbackMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* GetClientAccumulatedStats(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_GetClientStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_GetClientStats() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::LoadBalancerStatsRequest, ::connectrpc::conformance::LoadBalancerStatsResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::LoadBalancerStatsRequest, ::connectrpc::conformance::LoadBalancerStatsResponse>* streamer) {
                       return this->StreamedGetClientStats(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_GetClientStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status GetClientStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedGetClientStats(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::LoadBalancerStatsRequest,::connectrpc::conformance::LoadBalancerStatsResponse>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_GetClientAccumulatedStats : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_GetClientAccumulatedStats() {
      ::grpc::Service::MarkMethodStreamed(1,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* streamer) {
                       return this->StreamedGetClientAccumulatedStats(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_GetClientAccumulatedStats() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status GetClientAccumulatedStats(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest* /*request*/, ::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedGetClientAccumulatedStats(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::LoadBalancerAccumulatedStatsRequest,::connectrpc::conformance::LoadBalancerAccumulatedStatsResponse>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_GetClientStats<WithStreamedUnaryMethod_GetClientAccumulatedStats<Service > > StreamedUnaryService;
  typedef Service SplitStreamedService;
  typedef WithStreamedUnaryMethod_GetClientStats<WithStreamedUnaryMethod_GetClientAccumulatedStats<Service > > StreamedService;
};

// A service to remotely control health status of an xDS test server.
class XdsUpdateHealthService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.XdsUpdateHealthService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    virtual ::grpc::Status SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncSetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncSetServingRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncSetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncSetServingRaw(context, request, cq));
    }
    virtual ::grpc::Status SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> AsyncSetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(AsyncSetNotServingRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>> PrepareAsyncSetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>>(PrepareAsyncSetNotServingRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      virtual void SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
      virtual void SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) = 0;
      virtual void SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncSetServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncSetServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* AsyncSetNotServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::google::protobuf::Empty>* PrepareAsyncSetNotServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncSetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncSetServingRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncSetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncSetServingRaw(context, request, cq));
    }
    ::grpc::Status SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::google::protobuf::Empty* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> AsyncSetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(AsyncSetNotServingRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>> PrepareAsyncSetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>>(PrepareAsyncSetNotServingRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void SetServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
      void SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, std::function<void(::grpc::Status)>) override;
      void SetNotServing(::grpc::ClientContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response, ::grpc::ClientUnaryReactor* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncSetServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncSetServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* AsyncSetNotServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::google::protobuf::Empty>* PrepareAsyncSetNotServingRaw(::grpc::ClientContext* context, const ::google::protobuf::Empty& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_SetServing_;
    const ::grpc::internal::RpcMethod rpcmethod_SetNotServing_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    virtual ::grpc::Status SetServing(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response);
    virtual ::grpc::Status SetNotServing(::grpc::ServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response);
  };
  template <class BaseClass>
  class WithAsyncMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_SetServing() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestSetServing(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithAsyncMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_SetNotServing() {
      ::grpc::Service::MarkMethodAsync(1);
    }
    ~WithAsyncMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestSetNotServing(::grpc::ServerContext* context, ::google::protobuf::Empty* request, ::grpc::ServerAsyncResponseWriter< ::google::protobuf::Empty>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_SetServing<WithAsyncMethod_SetNotServing<Service > > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_SetServing() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response) { return this->SetServing(context, request, response); }));}
    void SetMessageAllocatorFor_SetServing(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* SetServing(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithCallbackMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_SetNotServing() {
      ::grpc::Service::MarkMethodCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::google::protobuf::Empty* request, ::google::protobuf::Empty* response) { return this->SetNotServing(context, request, response); }));}
    void SetMessageAllocatorFor_SetNotServing(
        ::grpc::MessageAllocator< ::google::protobuf::Empty, ::google::protobuf::Empty>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(1);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::google::protobuf::Empty, ::google::protobuf::Empty>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* SetNotServing(
      ::grpc::CallbackServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_SetServing<WithCallbackMethod_SetNotServing<Service > > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_SetServing() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithGenericMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_SetNotServing() {
      ::grpc::Service::MarkMethodGeneric(1);
    }
    ~WithGenericMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_SetServing() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestSetServing(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_SetNotServing() {
      ::grpc::Service::MarkMethodRaw(1);
    }
    ~WithRawMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestSetNotServing(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(1, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_SetServing() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->SetServing(context, request, response); }));
    }
    ~WithRawCallbackMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* SetServing(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_SetNotServing() {
      ::grpc::Service::MarkMethodRawCallback(1,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->SetNotServing(context, request, response); }));
    }
    ~WithRawCallbackMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* SetNotServing(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_SetServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_SetServing() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedSetServing(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_SetServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status SetServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedSetServing(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_SetNotServing : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_SetNotServing() {
      ::grpc::Service::MarkMethodStreamed(1,
        new ::grpc::internal::StreamedUnaryHandler<
          ::google::protobuf::Empty, ::google::protobuf::Empty>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::google::protobuf::Empty, ::google::protobuf::Empty>* streamer) {
                       return this->StreamedSetNotServing(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_SetNotServing() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status SetNotServing(::grpc::ServerContext* /*context*/, const ::google::protobuf::Empty* /*request*/, ::google::protobuf::Empty* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedSetNotServing(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::google::protobuf::Empty,::google::protobuf::Empty>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_SetServing<WithStreamedUnaryMethod_SetNotServing<Service > > StreamedUnaryService;
  typedef Service SplitStreamedService;
  typedef WithStreamedUnaryMethod_SetServing<WithStreamedUnaryMethod_SetNotServing<Service > > StreamedService;
};

// A service to dynamically update the configuration of an xDS test client.
class XdsUpdateClientConfigureService final {
 public:
  static constexpr char const* service_full_name() {
    return "connectrpc.conformance.XdsUpdateClientConfigureService";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    // Update the tes client's configuration.
    virtual ::grpc::Status Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::connectrpc::conformance::ClientConfigureResponse* response) = 0;
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>> AsyncConfigure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>>(AsyncConfigureRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>> PrepareAsyncConfigure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>>(PrepareAsyncConfigureRaw(context, request, cq));
    }
    class async_interface {
     public:
      virtual ~async_interface() {}
      // Update the tes client's configuration.
      virtual void Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response, std::function<void(::grpc::Status)>) = 0;
      virtual void Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response, ::grpc::ClientUnaryReactor* reactor) = 0;
    };
    typedef class async_interface experimental_async_interface;
    virtual class async_interface* async() { return nullptr; }
    class async_interface* experimental_async() { return async(); }
   private:
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>* AsyncConfigureRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) = 0;
    virtual ::grpc::ClientAsyncResponseReaderInterface< ::connectrpc::conformance::ClientConfigureResponse>* PrepareAsyncConfigureRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());
    ::grpc::Status Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::connectrpc::conformance::ClientConfigureResponse* response) override;
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>> AsyncConfigure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>>(AsyncConfigureRaw(context, request, cq));
    }
    std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>> PrepareAsyncConfigure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>>(PrepareAsyncConfigureRaw(context, request, cq));
    }
    class async final :
      public StubInterface::async_interface {
     public:
      void Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response, std::function<void(::grpc::Status)>) override;
      void Configure(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response, ::grpc::ClientUnaryReactor* reactor) override;
     private:
      friend class Stub;
      explicit async(Stub* stub): stub_(stub) { }
      Stub* stub() { return stub_; }
      Stub* stub_;
    };
    class async* async() override { return &async_stub_; }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    class async async_stub_{this};
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>* AsyncConfigureRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) override;
    ::grpc::ClientAsyncResponseReader< ::connectrpc::conformance::ClientConfigureResponse>* PrepareAsyncConfigureRaw(::grpc::ClientContext* context, const ::connectrpc::conformance::ClientConfigureRequest& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_Configure_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    // Update the tes client's configuration.
    virtual ::grpc::Status Configure(::grpc::ServerContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response);
  };
  template <class BaseClass>
  class WithAsyncMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithAsyncMethod_Configure() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestConfigure(::grpc::ServerContext* context, ::connectrpc::conformance::ClientConfigureRequest* request, ::grpc::ServerAsyncResponseWriter< ::connectrpc::conformance::ClientConfigureResponse>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_Configure<Service > AsyncService;
  template <class BaseClass>
  class WithCallbackMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithCallbackMethod_Configure() {
      ::grpc::Service::MarkMethodCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::ClientConfigureRequest, ::connectrpc::conformance::ClientConfigureResponse>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::connectrpc::conformance::ClientConfigureRequest* request, ::connectrpc::conformance::ClientConfigureResponse* response) { return this->Configure(context, request, response); }));}
    void SetMessageAllocatorFor_Configure(
        ::grpc::MessageAllocator< ::connectrpc::conformance::ClientConfigureRequest, ::connectrpc::conformance::ClientConfigureResponse>* allocator) {
      ::grpc::internal::MethodHandler* const handler = ::grpc::Service::GetHandler(0);
      static_cast<::grpc::internal::CallbackUnaryHandler< ::connectrpc::conformance::ClientConfigureRequest, ::connectrpc::conformance::ClientConfigureResponse>*>(handler)
              ->SetMessageAllocator(allocator);
    }
    ~WithCallbackMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Configure(
      ::grpc::CallbackServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/)  { return nullptr; }
  };
  typedef WithCallbackMethod_Configure<Service > CallbackService;
  typedef CallbackService ExperimentalCallbackService;
  template <class BaseClass>
  class WithGenericMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithGenericMethod_Configure() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawMethod_Configure() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestConfigure(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncResponseWriter< ::grpc::ByteBuffer>* response, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncUnary(0, context, request, response, new_call_cq, notification_cq, tag);
    }
  };
  template <class BaseClass>
  class WithRawCallbackMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithRawCallbackMethod_Configure() {
      ::grpc::Service::MarkMethodRawCallback(0,
          new ::grpc::internal::CallbackUnaryHandler< ::grpc::ByteBuffer, ::grpc::ByteBuffer>(
            [this](
                   ::grpc::CallbackServerContext* context, const ::grpc::ByteBuffer* request, ::grpc::ByteBuffer* response) { return this->Configure(context, request, response); }));
    }
    ~WithRawCallbackMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    virtual ::grpc::ServerUnaryReactor* Configure(
      ::grpc::CallbackServerContext* /*context*/, const ::grpc::ByteBuffer* /*request*/, ::grpc::ByteBuffer* /*response*/)  { return nullptr; }
  };
  template <class BaseClass>
  class WithStreamedUnaryMethod_Configure : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service* /*service*/) {}
   public:
    WithStreamedUnaryMethod_Configure() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::StreamedUnaryHandler<
          ::connectrpc::conformance::ClientConfigureRequest, ::connectrpc::conformance::ClientConfigureResponse>(
            [this](::grpc::ServerContext* context,
                   ::grpc::ServerUnaryStreamer<
                     ::connectrpc::conformance::ClientConfigureRequest, ::connectrpc::conformance::ClientConfigureResponse>* streamer) {
                       return this->StreamedConfigure(context,
                         streamer);
                  }));
    }
    ~WithStreamedUnaryMethod_Configure() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status Configure(::grpc::ServerContext* /*context*/, const ::connectrpc::conformance::ClientConfigureRequest* /*request*/, ::connectrpc::conformance::ClientConfigureResponse* /*response*/) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with streamed unary
    virtual ::grpc::Status StreamedConfigure(::grpc::ServerContext* context, ::grpc::ServerUnaryStreamer< ::connectrpc::conformance::ClientConfigureRequest,::connectrpc::conformance::ClientConfigureResponse>* server_unary_streamer) = 0;
  };
  typedef WithStreamedUnaryMethod_Configure<Service > StreamedUnaryService;
  typedef Service SplitStreamedService;
  typedef WithStreamedUnaryMethod_Configure<Service > StreamedService;
};

}  // namespace conformance
}  // namespace connectrpc


#endif  // GRPC_connectrpc_2fconformance_2ftest_2eproto__INCLUDED
