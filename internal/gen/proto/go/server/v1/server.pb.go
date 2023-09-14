// Copyright 2022-2023 The Connect Authors
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
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: server/v1/server.proto

package serverv1

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

type Protocol int32

const (
	Protocol_PROTOCOL_UNSPECIFIED Protocol = 0
	Protocol_PROTOCOL_GRPC        Protocol = 1
	Protocol_PROTOCOL_GRPC_WEB    Protocol = 2
)

// Enum value maps for Protocol.
var (
	Protocol_name = map[int32]string{
		0: "PROTOCOL_UNSPECIFIED",
		1: "PROTOCOL_GRPC",
		2: "PROTOCOL_GRPC_WEB",
	}
	Protocol_value = map[string]int32{
		"PROTOCOL_UNSPECIFIED": 0,
		"PROTOCOL_GRPC":        1,
		"PROTOCOL_GRPC_WEB":    2,
	}
)

func (x Protocol) Enum() *Protocol {
	p := new(Protocol)
	*p = x
	return p
}

func (x Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_server_v1_server_proto_enumTypes[0].Descriptor()
}

func (Protocol) Type() protoreflect.EnumType {
	return &file_server_v1_server_proto_enumTypes[0]
}

func (x Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Protocol.Descriptor instead.
func (Protocol) EnumDescriptor() ([]byte, []int) {
	return file_server_v1_server_proto_rawDescGZIP(), []int{0}
}

// ServerMetadata is the metadata returned from the server started by the server binary.
type ServerMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host      string             `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Protocols []*ProtocolSupport `protobuf:"bytes,2,rep,name=protocols,proto3" json:"protocols,omitempty"`
}

func (x *ServerMetadata) Reset() {
	*x = ServerMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_v1_server_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerMetadata) ProtoMessage() {}

func (x *ServerMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_server_v1_server_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerMetadata.ProtoReflect.Descriptor instead.
func (*ServerMetadata) Descriptor() ([]byte, []int) {
	return file_server_v1_server_proto_rawDescGZIP(), []int{0}
}

func (x *ServerMetadata) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ServerMetadata) GetProtocols() []*ProtocolSupport {
	if x != nil {
		return x.Protocols
	}
	return nil
}

type ProtocolSupport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Protocol     Protocol       `protobuf:"varint,1,opt,name=protocol,proto3,enum=server.v1.Protocol" json:"protocol,omitempty"`
	HttpVersions []*HTTPVersion `protobuf:"bytes,2,rep,name=http_versions,json=httpVersions,proto3" json:"http_versions,omitempty"`
	Port         string         `protobuf:"bytes,3,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *ProtocolSupport) Reset() {
	*x = ProtocolSupport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_v1_server_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtocolSupport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtocolSupport) ProtoMessage() {}

func (x *ProtocolSupport) ProtoReflect() protoreflect.Message {
	mi := &file_server_v1_server_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtocolSupport.ProtoReflect.Descriptor instead.
func (*ProtocolSupport) Descriptor() ([]byte, []int) {
	return file_server_v1_server_proto_rawDescGZIP(), []int{1}
}

func (x *ProtocolSupport) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

func (x *ProtocolSupport) GetHttpVersions() []*HTTPVersion {
	if x != nil {
		return x.HttpVersions
	}
	return nil
}

func (x *ProtocolSupport) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

type HTTPVersion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Major int32 `protobuf:"varint,1,opt,name=major,proto3" json:"major,omitempty"`
	Minor int32 `protobuf:"varint,2,opt,name=minor,proto3" json:"minor,omitempty"`
}

func (x *HTTPVersion) Reset() {
	*x = HTTPVersion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_v1_server_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HTTPVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPVersion) ProtoMessage() {}

func (x *HTTPVersion) ProtoReflect() protoreflect.Message {
	mi := &file_server_v1_server_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPVersion.ProtoReflect.Descriptor instead.
func (*HTTPVersion) Descriptor() ([]byte, []int) {
	return file_server_v1_server_proto_rawDescGZIP(), []int{2}
}

func (x *HTTPVersion) GetMajor() int32 {
	if x != nil {
		return x.Major
	}
	return 0
}

func (x *HTTPVersion) GetMinor() int32 {
	if x != nil {
		return x.Minor
	}
	return 0
}

var File_server_v1_server_proto protoreflect.FileDescriptor

var file_server_v1_server_proto_rawDesc = []byte{
	0x0a, 0x16, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x22, 0x5e, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x73, 0x22, 0x93, 0x01, 0x0a, 0x0f, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x53, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x2f, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x3b, 0x0a, 0x0d, 0x68, 0x74, 0x74, 0x70,
	0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x39, 0x0a, 0x0b, 0x48, 0x54, 0x54,
	0x50, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x61, 0x6a, 0x6f,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x12, 0x14,
	0x0a, 0x05, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6d,
	0x69, 0x6e, 0x6f, 0x72, 0x2a, 0x4e, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x12, 0x18, 0x0a, 0x14, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x11, 0x0a, 0x0d, 0x50, 0x52,
	0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x47, 0x52, 0x50, 0x43, 0x10, 0x01, 0x12, 0x15, 0x0a,
	0x11, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x47, 0x52, 0x50, 0x43, 0x5f, 0x57,
	0x45, 0x42, 0x10, 0x02, 0x42, 0xa6, 0x01, 0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x43, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x76,
	0x31, 0x3b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x58, 0x58,
	0xaa, 0x02, 0x09, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x09, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x15, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_server_v1_server_proto_rawDescOnce sync.Once
	file_server_v1_server_proto_rawDescData = file_server_v1_server_proto_rawDesc
)

func file_server_v1_server_proto_rawDescGZIP() []byte {
	file_server_v1_server_proto_rawDescOnce.Do(func() {
		file_server_v1_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_server_v1_server_proto_rawDescData)
	})
	return file_server_v1_server_proto_rawDescData
}

var file_server_v1_server_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_server_v1_server_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_server_v1_server_proto_goTypes = []interface{}{
	(Protocol)(0),           // 0: server.v1.Protocol
	(*ServerMetadata)(nil),  // 1: server.v1.ServerMetadata
	(*ProtocolSupport)(nil), // 2: server.v1.ProtocolSupport
	(*HTTPVersion)(nil),     // 3: server.v1.HTTPVersion
}
var file_server_v1_server_proto_depIdxs = []int32{
	2, // 0: server.v1.ServerMetadata.protocols:type_name -> server.v1.ProtocolSupport
	0, // 1: server.v1.ProtocolSupport.protocol:type_name -> server.v1.Protocol
	3, // 2: server.v1.ProtocolSupport.http_versions:type_name -> server.v1.HTTPVersion
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_server_v1_server_proto_init() }
func file_server_v1_server_proto_init() {
	if File_server_v1_server_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_server_v1_server_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerMetadata); i {
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
		file_server_v1_server_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtocolSupport); i {
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
		file_server_v1_server_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HTTPVersion); i {
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
			RawDescriptor: file_server_v1_server_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_server_v1_server_proto_goTypes,
		DependencyIndexes: file_server_v1_server_proto_depIdxs,
		EnumInfos:         file_server_v1_server_proto_enumTypes,
		MessageInfos:      file_server_v1_server_proto_msgTypes,
	}.Build()
	File_server_v1_server_proto = out.File
	file_server_v1_server_proto_rawDesc = nil
	file_server_v1_server_proto_goTypes = nil
	file_server_v1_server_proto_depIdxs = nil
}
