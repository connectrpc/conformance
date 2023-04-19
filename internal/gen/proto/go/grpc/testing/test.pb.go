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

// This is copied from gRPC's testing Protobuf definitions: https://github.com/grpc/grpc/blob/master/src/proto/grpc/testing/test.proto
//
// The TestService has been extended to include the following RPCs:
// FailUnaryCall(SimpleRequest) returns (SimpleResponse): this RPC is a unary
// call that always returns a readable non-ASCII error with error details.
// FailStreamingOutputCall(StreamingOutputCallRequest) returns (stream StreamingOutputCallResponse):
// this RPC is a server streaming call that always returns a readable non-ASCII error with error details.
// UnimplementedStreamingOutputCall(grpc.testing.Empty) returns (stream grpc.testing.Empty): this RPC
// is a server streaming call that will not be implemented.
//
// The UnimplementedService has been extended to include the following RPCs:
// UnimplementedStreamingOutputCall(grpc.testing.Empty) returns (stream grpc.testing.Empty): this RPC
// is a server streaming call that will not be implemented.

// Copyright 2015-2016 gRPC authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// An integration test service that covers all the method signature permutations
// of unary/streaming requests/responses.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: grpc/testing/test.proto

package testing

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_grpc_testing_test_proto protoreflect.FileDescriptor

var file_grpc_testing_test_proto_rawDesc = []byte{
	0x0a, 0x17, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x1a, 0x18, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0x9f,
	0x09, 0x0a, 0x0b, 0x54, 0x65, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x35,
	0x0a, 0x09, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x13, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x46, 0x0a, 0x09, 0x55, 0x6e, 0x61, 0x72, 0x79, 0x43, 0x61,
	0x6c, 0x6c, 0x12, 0x1b, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1c, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53,
	0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4e, 0x0a,
	0x0c, 0x47, 0x65, 0x74, 0x55, 0x6e, 0x61, 0x72, 0x79, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1b, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69, 0x6d,
	0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x12, 0x4a, 0x0a,
	0x0d, 0x46, 0x61, 0x69, 0x6c, 0x55, 0x6e, 0x61, 0x72, 0x79, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1b,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69,
	0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4f, 0x0a, 0x12, 0x43, 0x61, 0x63,
	0x68, 0x65, 0x61, 0x62, 0x6c, 0x65, 0x55, 0x6e, 0x61, 0x72, 0x79, 0x43, 0x61, 0x6c, 0x6c, 0x12,
	0x1b, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53,
	0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69, 0x6d, 0x70,
	0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6c, 0x0a, 0x13, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c,
	0x6c, 0x12, 0x28, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x74, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61,
	0x6c, 0x6c, 0x12, 0x28, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x30, 0x01, 0x12, 0x70,
	0x0a, 0x17, 0x46, 0x61, 0x69, 0x6c, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x28, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69,
	0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01,
	0x12, 0x69, 0x0a, 0x12, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70,
	0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x27, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x28, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c,
	0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x12, 0x69, 0x0a, 0x0e, 0x46,
	0x75, 0x6c, 0x6c, 0x44, 0x75, 0x70, 0x6c, 0x65, 0x78, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x28, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67,
	0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x28, 0x01, 0x30, 0x01, 0x12, 0x69, 0x0a, 0x0e, 0x48, 0x61, 0x6c, 0x66, 0x44, 0x75,
	0x70, 0x6c, 0x65, 0x78, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x28, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e,
	0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x29, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x30,
	0x01, 0x12, 0x3d, 0x0a, 0x11, 0x55, 0x6e, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x65, 0x64, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x12, 0x4e, 0x0a, 0x20, 0x55, 0x6e, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x65,
	0x64, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x43, 0x61, 0x6c, 0x6c, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x30, 0x01,
	0x32, 0xa5, 0x01, 0x0a, 0x14, 0x55, 0x6e, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x65, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x11, 0x55, 0x6e, 0x69,
	0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x65, 0x64, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x13,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x4e, 0x0a, 0x20, 0x55, 0x6e, 0x69, 0x6d,
	0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x65, 0x64, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69,
	0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x13, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x30, 0x01, 0x32, 0x89, 0x01, 0x0a, 0x10, 0x52, 0x65, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3b, 0x0a,
	0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73,
	0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x38, 0x0a, 0x04, 0x53, 0x74,
	0x6f, 0x70, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x49, 0x6e, 0x66, 0x6f, 0x32, 0x86, 0x02, 0x0a, 0x18, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x63, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x73, 0x12, 0x26, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42,
	0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x84, 0x01, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x41, 0x63, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x12, 0x31, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72,
	0x41, 0x63, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x53, 0x74, 0x61, 0x74, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e,
	0x63, 0x65, 0x72, 0x41, 0x63, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x53, 0x74,
	0x61, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0x8b, 0x01,
	0x0a, 0x16, 0x58, 0x64, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x36, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x12, 0x39, 0x0a, 0x0d, 0x53, 0x65, 0x74, 0x4e, 0x6f, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e,
	0x67, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x32, 0x7b, 0x0a, 0x1f, 0x58,
	0x64, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x58,
	0x0a, 0x09, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x12, 0x24, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x25, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xb8, 0x01, 0x0a, 0x10, 0x63, 0x6f, 0x6d,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x42, 0x09, 0x54,
	0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x48, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x66, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2d, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x74, 0x65, 0x73,
	0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73,
	0x74, 0x69, 0x6e, 0x67, 0xa2, 0x02, 0x03, 0x47, 0x54, 0x58, 0xaa, 0x02, 0x0c, 0x47, 0x72, 0x70,
	0x63, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0xca, 0x02, 0x0c, 0x47, 0x72, 0x70, 0x63,
	0x5c, 0x54, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0xe2, 0x02, 0x18, 0x47, 0x72, 0x70, 0x63, 0x5c,
	0x54, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x47, 0x72, 0x70, 0x63, 0x3a, 0x3a, 0x54, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_grpc_testing_test_proto_goTypes = []interface{}{
	(*Empty)(nil),                                // 0: grpc.testing.Empty
	(*SimpleRequest)(nil),                        // 1: grpc.testing.SimpleRequest
	(*StreamingOutputCallRequest)(nil),           // 2: grpc.testing.StreamingOutputCallRequest
	(*StreamingInputCallRequest)(nil),            // 3: grpc.testing.StreamingInputCallRequest
	(*ReconnectParams)(nil),                      // 4: grpc.testing.ReconnectParams
	(*LoadBalancerStatsRequest)(nil),             // 5: grpc.testing.LoadBalancerStatsRequest
	(*LoadBalancerAccumulatedStatsRequest)(nil),  // 6: grpc.testing.LoadBalancerAccumulatedStatsRequest
	(*ClientConfigureRequest)(nil),               // 7: grpc.testing.ClientConfigureRequest
	(*SimpleResponse)(nil),                       // 8: grpc.testing.SimpleResponse
	(*StreamingOutputCallResponse)(nil),          // 9: grpc.testing.StreamingOutputCallResponse
	(*StreamingInputCallResponse)(nil),           // 10: grpc.testing.StreamingInputCallResponse
	(*ReconnectInfo)(nil),                        // 11: grpc.testing.ReconnectInfo
	(*LoadBalancerStatsResponse)(nil),            // 12: grpc.testing.LoadBalancerStatsResponse
	(*LoadBalancerAccumulatedStatsResponse)(nil), // 13: grpc.testing.LoadBalancerAccumulatedStatsResponse
	(*ClientConfigureResponse)(nil),              // 14: grpc.testing.ClientConfigureResponse
}
var file_grpc_testing_test_proto_depIdxs = []int32{
	0,  // 0: grpc.testing.TestService.EmptyCall:input_type -> grpc.testing.Empty
	1,  // 1: grpc.testing.TestService.UnaryCall:input_type -> grpc.testing.SimpleRequest
	1,  // 2: grpc.testing.TestService.GetUnaryCall:input_type -> grpc.testing.SimpleRequest
	1,  // 3: grpc.testing.TestService.FailUnaryCall:input_type -> grpc.testing.SimpleRequest
	1,  // 4: grpc.testing.TestService.CacheableUnaryCall:input_type -> grpc.testing.SimpleRequest
	2,  // 5: grpc.testing.TestService.StreamingOutputCall:input_type -> grpc.testing.StreamingOutputCallRequest
	2,  // 6: grpc.testing.TestService.GetStreamingOutputCall:input_type -> grpc.testing.StreamingOutputCallRequest
	2,  // 7: grpc.testing.TestService.FailStreamingOutputCall:input_type -> grpc.testing.StreamingOutputCallRequest
	3,  // 8: grpc.testing.TestService.StreamingInputCall:input_type -> grpc.testing.StreamingInputCallRequest
	2,  // 9: grpc.testing.TestService.FullDuplexCall:input_type -> grpc.testing.StreamingOutputCallRequest
	2,  // 10: grpc.testing.TestService.HalfDuplexCall:input_type -> grpc.testing.StreamingOutputCallRequest
	0,  // 11: grpc.testing.TestService.UnimplementedCall:input_type -> grpc.testing.Empty
	0,  // 12: grpc.testing.TestService.UnimplementedStreamingOutputCall:input_type -> grpc.testing.Empty
	0,  // 13: grpc.testing.UnimplementedService.UnimplementedCall:input_type -> grpc.testing.Empty
	0,  // 14: grpc.testing.UnimplementedService.UnimplementedStreamingOutputCall:input_type -> grpc.testing.Empty
	4,  // 15: grpc.testing.ReconnectService.Start:input_type -> grpc.testing.ReconnectParams
	0,  // 16: grpc.testing.ReconnectService.Stop:input_type -> grpc.testing.Empty
	5,  // 17: grpc.testing.LoadBalancerStatsService.GetClientStats:input_type -> grpc.testing.LoadBalancerStatsRequest
	6,  // 18: grpc.testing.LoadBalancerStatsService.GetClientAccumulatedStats:input_type -> grpc.testing.LoadBalancerAccumulatedStatsRequest
	0,  // 19: grpc.testing.XdsUpdateHealthService.SetServing:input_type -> grpc.testing.Empty
	0,  // 20: grpc.testing.XdsUpdateHealthService.SetNotServing:input_type -> grpc.testing.Empty
	7,  // 21: grpc.testing.XdsUpdateClientConfigureService.Configure:input_type -> grpc.testing.ClientConfigureRequest
	0,  // 22: grpc.testing.TestService.EmptyCall:output_type -> grpc.testing.Empty
	8,  // 23: grpc.testing.TestService.UnaryCall:output_type -> grpc.testing.SimpleResponse
	8,  // 24: grpc.testing.TestService.GetUnaryCall:output_type -> grpc.testing.SimpleResponse
	8,  // 25: grpc.testing.TestService.FailUnaryCall:output_type -> grpc.testing.SimpleResponse
	8,  // 26: grpc.testing.TestService.CacheableUnaryCall:output_type -> grpc.testing.SimpleResponse
	9,  // 27: grpc.testing.TestService.StreamingOutputCall:output_type -> grpc.testing.StreamingOutputCallResponse
	9,  // 28: grpc.testing.TestService.GetStreamingOutputCall:output_type -> grpc.testing.StreamingOutputCallResponse
	9,  // 29: grpc.testing.TestService.FailStreamingOutputCall:output_type -> grpc.testing.StreamingOutputCallResponse
	10, // 30: grpc.testing.TestService.StreamingInputCall:output_type -> grpc.testing.StreamingInputCallResponse
	9,  // 31: grpc.testing.TestService.FullDuplexCall:output_type -> grpc.testing.StreamingOutputCallResponse
	9,  // 32: grpc.testing.TestService.HalfDuplexCall:output_type -> grpc.testing.StreamingOutputCallResponse
	0,  // 33: grpc.testing.TestService.UnimplementedCall:output_type -> grpc.testing.Empty
	0,  // 34: grpc.testing.TestService.UnimplementedStreamingOutputCall:output_type -> grpc.testing.Empty
	0,  // 35: grpc.testing.UnimplementedService.UnimplementedCall:output_type -> grpc.testing.Empty
	0,  // 36: grpc.testing.UnimplementedService.UnimplementedStreamingOutputCall:output_type -> grpc.testing.Empty
	0,  // 37: grpc.testing.ReconnectService.Start:output_type -> grpc.testing.Empty
	11, // 38: grpc.testing.ReconnectService.Stop:output_type -> grpc.testing.ReconnectInfo
	12, // 39: grpc.testing.LoadBalancerStatsService.GetClientStats:output_type -> grpc.testing.LoadBalancerStatsResponse
	13, // 40: grpc.testing.LoadBalancerStatsService.GetClientAccumulatedStats:output_type -> grpc.testing.LoadBalancerAccumulatedStatsResponse
	0,  // 41: grpc.testing.XdsUpdateHealthService.SetServing:output_type -> grpc.testing.Empty
	0,  // 42: grpc.testing.XdsUpdateHealthService.SetNotServing:output_type -> grpc.testing.Empty
	14, // 43: grpc.testing.XdsUpdateClientConfigureService.Configure:output_type -> grpc.testing.ClientConfigureResponse
	22, // [22:44] is the sub-list for method output_type
	0,  // [0:22] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_grpc_testing_test_proto_init() }
func file_grpc_testing_test_proto_init() {
	if File_grpc_testing_test_proto != nil {
		return
	}
	file_grpc_testing_empty_proto_init()
	file_grpc_testing_messages_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_grpc_testing_test_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   6,
		},
		GoTypes:           file_grpc_testing_test_proto_goTypes,
		DependencyIndexes: file_grpc_testing_test_proto_depIdxs,
	}.Build()
	File_grpc_testing_test_proto = out.File
	file_grpc_testing_test_proto_rawDesc = nil
	file_grpc_testing_test_proto_goTypes = nil
	file_grpc_testing_test_proto_depIdxs = nil
}
