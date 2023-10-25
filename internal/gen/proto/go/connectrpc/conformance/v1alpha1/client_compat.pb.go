// Copyright 2023 The Connect Authors
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: connectrpc/conformance/v1alpha1/client_compat.proto

package conformancev1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Describes one call the client should make. The client reads
// these from stdin and, for each one, invokes an RPC as directed
// and writes the results (in the form of a ClientCompatResponse
// message) to stdout.
type ClientCompatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TestName    string      `protobuf:"bytes,1,opt,name=test_name,json=testName,proto3" json:"test_name,omitempty"`
	HttpVersion HTTPVersion `protobuf:"varint,2,opt,name=http_version,json=httpVersion,proto3,enum=connectrpc.conformance.v1alpha1.HTTPVersion" json:"http_version,omitempty"`
	Protocol    Protocol    `protobuf:"varint,3,opt,name=protocol,proto3,enum=connectrpc.conformance.v1alpha1.Protocol" json:"protocol,omitempty"`
	Codec       Codec       `protobuf:"varint,4,opt,name=codec,proto3,enum=connectrpc.conformance.v1alpha1.Codec" json:"codec,omitempty"`
	Compression Compression `protobuf:"varint,5,opt,name=compression,proto3,enum=connectrpc.conformance.v1alpha1.Compression" json:"compression,omitempty"`
	Host        string      `protobuf:"bytes,6,opt,name=host,proto3" json:"host,omitempty"`
	Port        uint32      `protobuf:"varint,7,opt,name=port,proto3" json:"port,omitempty"`
	// If non-empty, the server is using TLS. The bytes are the
	// server's PEM-encoded certificate, which the client should
	// verify and trust.
	ServerTlsCert  []byte     `protobuf:"bytes,8,opt,name=server_tls_cert,json=serverTlsCert,proto3" json:"server_tls_cert,omitempty"`
	Service        string     `protobuf:"bytes,9,opt,name=service,proto3" json:"service,omitempty"`
	Method         string     `protobuf:"bytes,10,opt,name=method,proto3" json:"method,omitempty"`
	StreamType     StreamType `protobuf:"varint,11,opt,name=stream_type,json=streamType,proto3,enum=connectrpc.conformance.v1alpha1.StreamType" json:"stream_type,omitempty"`
	RequestHeaders []*Header  `protobuf:"bytes,12,rep,name=request_headers,json=requestHeaders,proto3" json:"request_headers,omitempty"`
	// There will be exactly one for unary and server-stream methods.
	// For client- and bidi-stream methods, all entries will have the
	// same type URL (which matches the request type of the method).
	RequestMessages []*anypb.Any `protobuf:"bytes,13,rep,name=request_messages,json=requestMessages,proto3" json:"request_messages,omitempty"`
	TimeoutMs       *uint32      `protobuf:"varint,14,opt,name=timeout_ms,json=timeoutMs,proto3,oneof" json:"timeout_ms,omitempty"`
	// wait this many milliseconds before sending a request message
	ResponseDelayMs uint32 `protobuf:"varint,15,opt,name=response_delay_ms,json=responseDelayMs,proto3" json:"response_delay_ms,omitempty"`
	// Tells the server whether it should wait for each request
	// before sending a response.
	// If true, it indicates the server should effectively interleave the
	// stream so messages are sent in request->response pairs.
	// If false, then the response stream will be sent once all request messages
	// are finished sending with the only delays between messages
	// being the optional fixed milliseconds defined in the response
	// definition.
	// This field is only relevant in the first message in the stream
	// and should be ignored in subsequent messages.
	// Note, this is only applicable to bidi endpoints.
	FullDuplex bool `protobuf:"varint,16,opt,name=full_duplex,json=fullDuplex,proto3" json:"full_duplex,omitempty"`
}

func (x *ClientCompatRequest) Reset() {
	*x = ClientCompatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientCompatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientCompatRequest) ProtoMessage() {}

func (x *ClientCompatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientCompatRequest.ProtoReflect.Descriptor instead.
func (*ClientCompatRequest) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescGZIP(), []int{0}
}

func (x *ClientCompatRequest) GetTestName() string {
	if x != nil {
		return x.TestName
	}
	return ""
}

func (x *ClientCompatRequest) GetHttpVersion() HTTPVersion {
	if x != nil {
		return x.HttpVersion
	}
	return HTTPVersion_HTTP_VERSION_UNSPECIFIED
}

func (x *ClientCompatRequest) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

func (x *ClientCompatRequest) GetCodec() Codec {
	if x != nil {
		return x.Codec
	}
	return Codec_CODEC_UNSPECIFIED
}

func (x *ClientCompatRequest) GetCompression() Compression {
	if x != nil {
		return x.Compression
	}
	return Compression_COMPRESSION_UNSPECIFIED
}

func (x *ClientCompatRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ClientCompatRequest) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ClientCompatRequest) GetServerTlsCert() []byte {
	if x != nil {
		return x.ServerTlsCert
	}
	return nil
}

func (x *ClientCompatRequest) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *ClientCompatRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *ClientCompatRequest) GetStreamType() StreamType {
	if x != nil {
		return x.StreamType
	}
	return StreamType_STREAM_TYPE_UNSPECIFIED
}

func (x *ClientCompatRequest) GetRequestHeaders() []*Header {
	if x != nil {
		return x.RequestHeaders
	}
	return nil
}

func (x *ClientCompatRequest) GetRequestMessages() []*anypb.Any {
	if x != nil {
		return x.RequestMessages
	}
	return nil
}

func (x *ClientCompatRequest) GetTimeoutMs() uint32 {
	if x != nil && x.TimeoutMs != nil {
		return *x.TimeoutMs
	}
	return 0
}

func (x *ClientCompatRequest) GetResponseDelayMs() uint32 {
	if x != nil {
		return x.ResponseDelayMs
	}
	return 0
}

func (x *ClientCompatRequest) GetFullDuplex() bool {
	if x != nil {
		return x.FullDuplex
	}
	return false
}

// The outcome of one ClientCompatRequest.
type ClientCompatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TestName string `protobuf:"bytes,1,opt,name=test_name,json=testName,proto3" json:"test_name,omitempty"`
	// Types that are assignable to Result:
	//	*ClientCompatResponse_Response
	//	*ClientCompatResponse_Error
	Result isClientCompatResponse_Result `protobuf_oneof:"result"`
}

func (x *ClientCompatResponse) Reset() {
	*x = ClientCompatResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientCompatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientCompatResponse) ProtoMessage() {}

func (x *ClientCompatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientCompatResponse.ProtoReflect.Descriptor instead.
func (*ClientCompatResponse) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescGZIP(), []int{1}
}

func (x *ClientCompatResponse) GetTestName() string {
	if x != nil {
		return x.TestName
	}
	return ""
}

func (m *ClientCompatResponse) GetResult() isClientCompatResponse_Result {
	if m != nil {
		return m.Result
	}
	return nil
}

func (x *ClientCompatResponse) GetResponse() *ClientResponseResult {
	if x, ok := x.GetResult().(*ClientCompatResponse_Response); ok {
		return x.Response
	}
	return nil
}

func (x *ClientCompatResponse) GetError() *ClientErrorResult {
	if x, ok := x.GetResult().(*ClientCompatResponse_Error); ok {
		return x.Error
	}
	return nil
}

type isClientCompatResponse_Result interface {
	isClientCompatResponse_Result()
}

type ClientCompatResponse_Response struct {
	Response *ClientResponseResult `protobuf:"bytes,2,opt,name=response,proto3,oneof"`
}

type ClientCompatResponse_Error struct {
	Error *ClientErrorResult `protobuf:"bytes,3,opt,name=error,proto3,oneof"`
}

func (*ClientCompatResponse_Response) isClientCompatResponse_Result() {}

func (*ClientCompatResponse_Error) isClientCompatResponse_Result() {}

// The result of a ClientCompatRequest, which may or may bot be successful.
type ClientResponseResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResponseHeaders []*Header             `protobuf:"bytes,1,rep,name=response_headers,json=responseHeaders,proto3" json:"response_headers,omitempty"`
	Payloads        []*ConformancePayload `protobuf:"bytes,2,rep,name=payloads,proto3" json:"payloads,omitempty"`
	Error           *Error                `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	// In case the client cannot decode Any from JSON, it should instead return the received JSON
	ErrorDetailsRaw  []*structpb.Struct `protobuf:"bytes,7,rep,name=error_details_raw,json=errorDetailsRaw,proto3" json:"error_details_raw,omitempty"`
	ResponseTrailers []*Header          `protobuf:"bytes,4,rep,name=response_trailers,json=responseTrailers,proto3" json:"response_trailers,omitempty"`
}

func (x *ClientResponseResult) Reset() {
	*x = ClientResponseResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientResponseResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientResponseResult) ProtoMessage() {}

func (x *ClientResponseResult) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientResponseResult.ProtoReflect.Descriptor instead.
func (*ClientResponseResult) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescGZIP(), []int{2}
}

func (x *ClientResponseResult) GetResponseHeaders() []*Header {
	if x != nil {
		return x.ResponseHeaders
	}
	return nil
}

func (x *ClientResponseResult) GetPayloads() []*ConformancePayload {
	if x != nil {
		return x.Payloads
	}
	return nil
}

func (x *ClientResponseResult) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ClientResponseResult) GetErrorDetailsRaw() []*structpb.Struct {
	if x != nil {
		return x.ErrorDetailsRaw
	}
	return nil
}

func (x *ClientResponseResult) GetResponseTrailers() []*Header {
	if x != nil {
		return x.ResponseTrailers
	}
	return nil
}

// The client is not able to fulfill the ClientCompatRequest. This may be due
// to a runtime error, or because the requested protocol is not supported.
type ClientErrorResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ClientErrorResult) Reset() {
	*x = ClientErrorResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientErrorResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientErrorResult) ProtoMessage() {}

func (x *ClientErrorResult) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientErrorResult.ProtoReflect.Descriptor instead.
func (*ClientErrorResult) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescGZIP(), []int{3}
}

func (x *ClientErrorResult) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_connectrpc_conformance_v1alpha1_client_compat_proto protoreflect.FileDescriptor

var file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDesc = []byte{
	0x0a, 0x33, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x2c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72,
	0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2d, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63,
	0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbb, 0x06, 0x0a,
	0x13, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65, 0x73, 0x74, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x4f, 0x0a, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x68, 0x74, 0x74, 0x70, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x45, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52,
	0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x3c, 0x0a, 0x05, 0x63, 0x6f, 0x64,
	0x65, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x63,
	0x52, 0x05, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x12, 0x4e, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x72,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43,
	0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70,
	0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x6f, 0x72, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12,
	0x26, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x74, 0x6c, 0x73, 0x5f, 0x63, 0x65,
	0x72, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x54, 0x6c, 0x73, 0x43, 0x65, 0x72, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x4c, 0x0a, 0x0b, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x50, 0x0a, 0x0f, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x27, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f,
	0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x3f, 0x0a, 0x10, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x0d, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0f, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x22, 0x0a, 0x0a, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x6d, 0x73, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x00,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x88, 0x01, 0x01, 0x12, 0x2a,
	0x0a, 0x11, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x64, 0x65, 0x6c, 0x61, 0x79,
	0x5f, 0x6d, 0x73, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0f, 0x72, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x4d, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x75,
	0x6c, 0x6c, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x65, 0x78, 0x18, 0x10, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0a, 0x66, 0x75, 0x6c, 0x6c, 0x44, 0x75, 0x70, 0x6c, 0x65, 0x78, 0x42, 0x0d, 0x0a, 0x0b, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x6d, 0x73, 0x22, 0xde, 0x01, 0x0a, 0x14, 0x43,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x53, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x35, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x48, 0x00, 0x52, 0x08, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4a, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x42, 0x08, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x94, 0x03, 0x0a, 0x14,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x52, 0x0a, 0x10, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x4f, 0x0a, 0x08, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61,
	0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52,
	0x08, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x73, 0x12, 0x3c, 0x0a, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x43, 0x0a, 0x11, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x5f, 0x72, 0x61, 0x77, 0x18, 0x07, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x0f, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x61, 0x77, 0x12, 0x54, 0x0a, 0x11,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x65, 0x72,
	0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x52, 0x10, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x72, 0x61, 0x69, 0x6c, 0x65,
	0x72, 0x73, 0x22, 0x2d, 0x0a, 0x11, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x42, 0xbc, 0x02, 0x0a, 0x23, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x11, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x64,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63,
	0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f,
	0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x3b, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x43, 0x58, 0xaa, 0x02, 0x1f, 0x43, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61,
	0x6e, 0x63, 0x65, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02, 0x1f, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x5c, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xe2, 0x02,
	0x2b, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x5c, 0x43, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x21, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x3a, 0x3a, 0x43, 0x6f, 0x6e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescOnce sync.Once
	file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescData = file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDesc
)

func file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescGZIP() []byte {
	file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescOnce.Do(func() {
		file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescData = protoimpl.X.CompressGZIP(file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescData)
	})
	return file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDescData
}

var file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_connectrpc_conformance_v1alpha1_client_compat_proto_goTypes = []interface{}{
	(*ClientCompatRequest)(nil),  // 0: connectrpc.conformance.v1alpha1.ClientCompatRequest
	(*ClientCompatResponse)(nil), // 1: connectrpc.conformance.v1alpha1.ClientCompatResponse
	(*ClientResponseResult)(nil), // 2: connectrpc.conformance.v1alpha1.ClientResponseResult
	(*ClientErrorResult)(nil),    // 3: connectrpc.conformance.v1alpha1.ClientErrorResult
	(HTTPVersion)(0),             // 4: connectrpc.conformance.v1alpha1.HTTPVersion
	(Protocol)(0),                // 5: connectrpc.conformance.v1alpha1.Protocol
	(Codec)(0),                   // 6: connectrpc.conformance.v1alpha1.Codec
	(Compression)(0),             // 7: connectrpc.conformance.v1alpha1.Compression
	(StreamType)(0),              // 8: connectrpc.conformance.v1alpha1.StreamType
	(*Header)(nil),               // 9: connectrpc.conformance.v1alpha1.Header
	(*anypb.Any)(nil),            // 10: google.protobuf.Any
	(*ConformancePayload)(nil),   // 11: connectrpc.conformance.v1alpha1.ConformancePayload
	(*Error)(nil),                // 12: connectrpc.conformance.v1alpha1.Error
	(*structpb.Struct)(nil),      // 13: google.protobuf.Struct
}
var file_connectrpc_conformance_v1alpha1_client_compat_proto_depIdxs = []int32{
	4,  // 0: connectrpc.conformance.v1alpha1.ClientCompatRequest.http_version:type_name -> connectrpc.conformance.v1alpha1.HTTPVersion
	5,  // 1: connectrpc.conformance.v1alpha1.ClientCompatRequest.protocol:type_name -> connectrpc.conformance.v1alpha1.Protocol
	6,  // 2: connectrpc.conformance.v1alpha1.ClientCompatRequest.codec:type_name -> connectrpc.conformance.v1alpha1.Codec
	7,  // 3: connectrpc.conformance.v1alpha1.ClientCompatRequest.compression:type_name -> connectrpc.conformance.v1alpha1.Compression
	8,  // 4: connectrpc.conformance.v1alpha1.ClientCompatRequest.stream_type:type_name -> connectrpc.conformance.v1alpha1.StreamType
	9,  // 5: connectrpc.conformance.v1alpha1.ClientCompatRequest.request_headers:type_name -> connectrpc.conformance.v1alpha1.Header
	10, // 6: connectrpc.conformance.v1alpha1.ClientCompatRequest.request_messages:type_name -> google.protobuf.Any
	2,  // 7: connectrpc.conformance.v1alpha1.ClientCompatResponse.response:type_name -> connectrpc.conformance.v1alpha1.ClientResponseResult
	3,  // 8: connectrpc.conformance.v1alpha1.ClientCompatResponse.error:type_name -> connectrpc.conformance.v1alpha1.ClientErrorResult
	9,  // 9: connectrpc.conformance.v1alpha1.ClientResponseResult.response_headers:type_name -> connectrpc.conformance.v1alpha1.Header
	11, // 10: connectrpc.conformance.v1alpha1.ClientResponseResult.payloads:type_name -> connectrpc.conformance.v1alpha1.ConformancePayload
	12, // 11: connectrpc.conformance.v1alpha1.ClientResponseResult.error:type_name -> connectrpc.conformance.v1alpha1.Error
	13, // 12: connectrpc.conformance.v1alpha1.ClientResponseResult.error_details_raw:type_name -> google.protobuf.Struct
	9,  // 13: connectrpc.conformance.v1alpha1.ClientResponseResult.response_trailers:type_name -> connectrpc.conformance.v1alpha1.Header
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_connectrpc_conformance_v1alpha1_client_compat_proto_init() }
func file_connectrpc_conformance_v1alpha1_client_compat_proto_init() {
	if File_connectrpc_conformance_v1alpha1_client_compat_proto != nil {
		return
	}
	file_connectrpc_conformance_v1alpha1_config_proto_init()
	file_connectrpc_conformance_v1alpha1_service_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientCompatRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientCompatResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientResponseResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientErrorResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*ClientCompatResponse_Response)(nil),
		(*ClientCompatResponse_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_connectrpc_conformance_v1alpha1_client_compat_proto_goTypes,
		DependencyIndexes: file_connectrpc_conformance_v1alpha1_client_compat_proto_depIdxs,
		MessageInfos:      file_connectrpc_conformance_v1alpha1_client_compat_proto_msgTypes,
	}.Build()
	File_connectrpc_conformance_v1alpha1_client_compat_proto = out.File
	file_connectrpc_conformance_v1alpha1_client_compat_proto_rawDesc = nil
	file_connectrpc_conformance_v1alpha1_client_compat_proto_goTypes = nil
	file_connectrpc_conformance_v1alpha1_client_compat_proto_depIdxs = nil
}
