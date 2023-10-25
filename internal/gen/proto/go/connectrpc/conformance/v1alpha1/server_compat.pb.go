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
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: connectrpc/conformance/v1alpha1/server_compat.proto

package conformancev1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Describes one configuration for an RPC server. The server is
// expected to expose the connectrpc.conformance.v1alpha1.ConformanceService
// RPC service. The configuration does not include a port. The
// process should pick an available port, which is typically
// done by using port zero (0) when creating a network listener
// so that the OS selects an available ephemeral port.
//
// These properties are read from stdin. Once the server is
// listening, details about the server, in the form of a
// ServerCompatResponse, are written to stdout.
//
// Each test process is expected to start only one RPC server.
// When testing multiple configurations, multiple test processes
// will be started, each with different properties.
type ServerCompatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Protocol    Protocol    `protobuf:"varint,1,opt,name=protocol,proto3,enum=connectrpc.conformance.v1alpha1.Protocol" json:"protocol,omitempty"`
	HttpVersion HTTPVersion `protobuf:"varint,2,opt,name=http_version,json=httpVersion,proto3,enum=connectrpc.conformance.v1alpha1.HTTPVersion" json:"http_version,omitempty"`
	// if true, generate a self-signed cert and include it in the
	// ServerCompatResponse along with the actual port
	UseTls bool `protobuf:"varint,4,opt,name=use_tls,json=useTls,proto3" json:"use_tls,omitempty"`
}

func (x *ServerCompatRequest) Reset() {
	*x = ServerCompatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerCompatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerCompatRequest) ProtoMessage() {}

func (x *ServerCompatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerCompatRequest.ProtoReflect.Descriptor instead.
func (*ServerCompatRequest) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescGZIP(), []int{0}
}

func (x *ServerCompatRequest) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

func (x *ServerCompatRequest) GetHttpVersion() HTTPVersion {
	if x != nil {
		return x.HttpVersion
	}
	return HTTPVersion_HTTP_VERSION_UNSPECIFIED
}

func (x *ServerCompatRequest) GetUseTls() bool {
	if x != nil {
		return x.UseTls
	}
	return false
}

// The outcome of one ServerCompatRequest.
type ServerCompatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// The server's PEM-encoded certificate, so the
	// client can verify it when connecting via TLS.
	PemCert []byte `protobuf:"bytes,3,opt,name=pem_cert,json=pemCert,proto3" json:"pem_cert,omitempty"`
}

func (x *ServerCompatResponse) Reset() {
	*x = ServerCompatResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerCompatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerCompatResponse) ProtoMessage() {}

func (x *ServerCompatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerCompatResponse.ProtoReflect.Descriptor instead.
func (*ServerCompatResponse) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescGZIP(), []int{1}
}

func (x *ServerCompatResponse) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ServerCompatResponse) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ServerCompatResponse) GetPemCert() []byte {
	if x != nil {
		return x.PemCert
	}
	return nil
}

// The server doesn't support the requested protocol, or had a runtime error
// while starting up.
type ServerErrorResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ServerErrorResult) Reset() {
	*x = ServerErrorResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerErrorResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerErrorResult) ProtoMessage() {}

func (x *ServerErrorResult) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerErrorResult.ProtoReflect.Descriptor instead.
func (*ServerErrorResult) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescGZIP(), []int{2}
}

func (x *ServerErrorResult) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_connectrpc_conformance_v1alpha1_server_compat_proto protoreflect.FileDescriptor

var file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDesc = []byte{
	0x0a, 0x33, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x2c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72,
	0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x01, 0x0a, 0x13, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43,
	0x6f, 0x6d, 0x70, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x45, 0x0a, 0x08,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x12, 0x4f, 0x0a, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e,
	0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x68, 0x74, 0x74, 0x70, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x5f, 0x74, 0x6c, 0x73, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x75, 0x73, 0x65, 0x54, 0x6c, 0x73, 0x22, 0x59, 0x0a,
	0x14, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x19, 0x0a,
	0x08, 0x70, 0x65, 0x6d, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x70, 0x65, 0x6d, 0x43, 0x65, 0x72, 0x74, 0x22, 0x2d, 0x0a, 0x11, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0xbc, 0x02, 0x0a, 0x23, 0x63, 0x6f, 0x6d, 0x2e,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42,
	0x11, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x64, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65,
	0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e,
	0x63, 0x65, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x43, 0x58,
	0xaa, 0x02, 0x1f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0xca, 0x02, 0x1f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x5c,
	0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0xe2, 0x02, 0x2b, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x5c, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x21, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x3a,
	0x3a, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x3a, 0x3a, 0x56, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescOnce sync.Once
	file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescData = file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDesc
)

func file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescGZIP() []byte {
	file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescOnce.Do(func() {
		file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescData = protoimpl.X.CompressGZIP(file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescData)
	})
	return file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDescData
}

var file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_connectrpc_conformance_v1alpha1_server_compat_proto_goTypes = []interface{}{
	(*ServerCompatRequest)(nil),  // 0: connectrpc.conformance.v1alpha1.ServerCompatRequest
	(*ServerCompatResponse)(nil), // 1: connectrpc.conformance.v1alpha1.ServerCompatResponse
	(*ServerErrorResult)(nil),    // 2: connectrpc.conformance.v1alpha1.ServerErrorResult
	(Protocol)(0),                // 3: connectrpc.conformance.v1alpha1.Protocol
	(HTTPVersion)(0),             // 4: connectrpc.conformance.v1alpha1.HTTPVersion
}
var file_connectrpc_conformance_v1alpha1_server_compat_proto_depIdxs = []int32{
	3, // 0: connectrpc.conformance.v1alpha1.ServerCompatRequest.protocol:type_name -> connectrpc.conformance.v1alpha1.Protocol
	4, // 1: connectrpc.conformance.v1alpha1.ServerCompatRequest.http_version:type_name -> connectrpc.conformance.v1alpha1.HTTPVersion
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_connectrpc_conformance_v1alpha1_server_compat_proto_init() }
func file_connectrpc_conformance_v1alpha1_server_compat_proto_init() {
	if File_connectrpc_conformance_v1alpha1_server_compat_proto != nil {
		return
	}
	file_connectrpc_conformance_v1alpha1_config_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerCompatRequest); i {
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
		file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerCompatResponse); i {
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
		file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerErrorResult); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_connectrpc_conformance_v1alpha1_server_compat_proto_goTypes,
		DependencyIndexes: file_connectrpc_conformance_v1alpha1_server_compat_proto_depIdxs,
		MessageInfos:      file_connectrpc_conformance_v1alpha1_server_compat_proto_msgTypes,
	}.Build()
	File_connectrpc_conformance_v1alpha1_server_compat_proto = out.File
	file_connectrpc_conformance_v1alpha1_server_compat_proto_rawDesc = nil
	file_connectrpc_conformance_v1alpha1_server_compat_proto_goTypes = nil
	file_connectrpc_conformance_v1alpha1_server_compat_proto_depIdxs = nil
}
