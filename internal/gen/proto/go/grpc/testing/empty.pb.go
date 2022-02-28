// Copyright 2020-2022 Buf Technologies, Inc.
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

// This is copied from gRPC's testing Protobuf definitions: https://github.com/grpc/grpc/blob/master/src/proto/grpc/testing/empty.proto

// Copyright 2015 gRPC authors.
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: grpc/testing/empty.proto

package testing

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

// An empty message that you can re-use to avoid defining duplicated empty
// messages in your project. A typical example is to use it as argument or the
// return value of a service API. For instance:
//
//   service Foo {
//     rpc Bar (grpc.testing.Empty) returns (grpc.testing.Empty) { };
//   };
//
type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_testing_empty_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_testing_empty_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_grpc_testing_empty_proto_rawDescGZIP(), []int{0}
}

var File_grpc_testing_empty_proto protoreflect.FileDescriptor

var file_grpc_testing_empty_proto_rawDesc = []byte{
	0x0a, 0x18, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x67, 0x72, 0x70, 0x63,
	0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x42, 0xb9, 0x01, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x42, 0x0a, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x62, 0x75, 0x66, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x2d, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
	0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0xa2, 0x02,
	0x03, 0x47, 0x54, 0x58, 0xaa, 0x02, 0x0c, 0x47, 0x72, 0x70, 0x63, 0x2e, 0x54, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0xca, 0x02, 0x0c, 0x47, 0x72, 0x70, 0x63, 0x5c, 0x54, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0xe2, 0x02, 0x18, 0x47, 0x72, 0x70, 0x63, 0x5c, 0x54, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d,
	0x47, 0x72, 0x70, 0x63, 0x3a, 0x3a, 0x54, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpc_testing_empty_proto_rawDescOnce sync.Once
	file_grpc_testing_empty_proto_rawDescData = file_grpc_testing_empty_proto_rawDesc
)

func file_grpc_testing_empty_proto_rawDescGZIP() []byte {
	file_grpc_testing_empty_proto_rawDescOnce.Do(func() {
		file_grpc_testing_empty_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_testing_empty_proto_rawDescData)
	})
	return file_grpc_testing_empty_proto_rawDescData
}

var file_grpc_testing_empty_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_grpc_testing_empty_proto_goTypes = []interface{}{
	(*Empty)(nil), // 0: grpc.testing.Empty
}
var file_grpc_testing_empty_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_grpc_testing_empty_proto_init() }
func file_grpc_testing_empty_proto_init() {
	if File_grpc_testing_empty_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_testing_empty_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_grpc_testing_empty_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_grpc_testing_empty_proto_goTypes,
		DependencyIndexes: file_grpc_testing_empty_proto_depIdxs,
		MessageInfos:      file_grpc_testing_empty_proto_msgTypes,
	}.Build()
	File_grpc_testing_empty_proto = out.File
	file_grpc_testing_empty_proto_rawDesc = nil
	file_grpc_testing_empty_proto_goTypes = nil
	file_grpc_testing_empty_proto_depIdxs = nil
}
