// Copyright 2019 Google LLC.
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
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: google/geo/type/viewport.proto

package _type

import (
	_type "connectrpc.com/conformance/internal/gen/proto/go/google/type"
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

// A latitude-longitude viewport, represented as two diagonally opposite `low`
// and `high` points. A viewport is considered a closed region, i.e. it includes
// its boundary. The latitude bounds must range between -90 to 90 degrees
// inclusive, and the longitude bounds must range between -180 to 180 degrees
// inclusive. Various cases include:
//
//   - If `low` = `high`, the viewport consists of that single point.
//
//   - If `low.longitude` > `high.longitude`, the longitude range is inverted
//     (the viewport crosses the 180 degree longitude line).
//
//   - If `low.longitude` = -180 degrees and `high.longitude` = 180 degrees,
//     the viewport includes all longitudes.
//
//   - If `low.longitude` = 180 degrees and `high.longitude` = -180 degrees,
//     the longitude range is empty.
//
//   - If `low.latitude` > `high.latitude`, the latitude range is empty.
//
// Both `low` and `high` must be populated, and the represented box cannot be
// empty (as specified by the definitions above). An empty viewport will result
// in an error.
//
// For example, this viewport fully encloses New York City:
//
//	{
//	    "low": {
//	        "latitude": 40.477398,
//	        "longitude": -74.259087
//	    },
//	    "high": {
//	        "latitude": 40.91618,
//	        "longitude": -73.70018
//	    }
//	}
type Viewport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. The low point of the viewport.
	Low *_type.LatLng `protobuf:"bytes,1,opt,name=low,proto3" json:"low,omitempty"`
	// Required. The high point of the viewport.
	High *_type.LatLng `protobuf:"bytes,2,opt,name=high,proto3" json:"high,omitempty"`
}

func (x *Viewport) Reset() {
	*x = Viewport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_geo_type_viewport_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Viewport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Viewport) ProtoMessage() {}

func (x *Viewport) ProtoReflect() protoreflect.Message {
	mi := &file_google_geo_type_viewport_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Viewport.ProtoReflect.Descriptor instead.
func (*Viewport) Descriptor() ([]byte, []int) {
	return file_google_geo_type_viewport_proto_rawDescGZIP(), []int{0}
}

func (x *Viewport) GetLow() *_type.LatLng {
	if x != nil {
		return x.Low
	}
	return nil
}

func (x *Viewport) GetHigh() *_type.LatLng {
	if x != nil {
		return x.High
	}
	return nil
}

var File_google_geo_type_viewport_proto protoreflect.FileDescriptor

var file_google_geo_type_viewport_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x67, 0x65, 0x6f, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x2f, 0x76, 0x69, 0x65, 0x77, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x74, 0x79, 0x70,
	0x65, 0x1a, 0x18, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6c,
	0x61, 0x74, 0x6c, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5a, 0x0a, 0x08, 0x56,
	0x69, 0x65, 0x77, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x25, 0x0a, 0x03, 0x6c, 0x6f, 0x77, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x2e, 0x4c, 0x61, 0x74, 0x4c, 0x6e, 0x67, 0x52, 0x03, 0x6c, 0x6f, 0x77, 0x12, 0x27,
	0x0a, 0x04, 0x68, 0x69, 0x67, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x4c, 0x61, 0x74, 0x4c, 0x6e,
	0x67, 0x52, 0x04, 0x68, 0x69, 0x67, 0x68, 0x42, 0xc4, 0x01, 0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x65, 0x6f, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x42,
	0x0d, 0x56, 0x69, 0x65, 0x77, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x40, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x67, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x67, 0x65, 0x6f, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0xa2, 0x02, 0x03, 0x47, 0x47, 0x54, 0xaa, 0x02, 0x0f, 0x47, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x47, 0x65, 0x6f, 0x2e, 0x54, 0x79, 0x70, 0x65, 0xca, 0x02, 0x0f, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x5c, 0x47, 0x65, 0x6f, 0x5c, 0x54, 0x79, 0x70, 0x65, 0xe2, 0x02, 0x1b, 0x47,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c, 0x47, 0x65, 0x6f, 0x5c, 0x54, 0x79, 0x70, 0x65, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x11, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x3a, 0x3a, 0x47, 0x65, 0x6f, 0x3a, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_geo_type_viewport_proto_rawDescOnce sync.Once
	file_google_geo_type_viewport_proto_rawDescData = file_google_geo_type_viewport_proto_rawDesc
)

func file_google_geo_type_viewport_proto_rawDescGZIP() []byte {
	file_google_geo_type_viewport_proto_rawDescOnce.Do(func() {
		file_google_geo_type_viewport_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_geo_type_viewport_proto_rawDescData)
	})
	return file_google_geo_type_viewport_proto_rawDescData
}

var file_google_geo_type_viewport_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_geo_type_viewport_proto_goTypes = []interface{}{
	(*Viewport)(nil),     // 0: google.geo.type.Viewport
	(*_type.LatLng)(nil), // 1: google.type.LatLng
}
var file_google_geo_type_viewport_proto_depIdxs = []int32{
	1, // 0: google.geo.type.Viewport.low:type_name -> google.type.LatLng
	1, // 1: google.geo.type.Viewport.high:type_name -> google.type.LatLng
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_geo_type_viewport_proto_init() }
func file_google_geo_type_viewport_proto_init() {
	if File_google_geo_type_viewport_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_geo_type_viewport_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Viewport); i {
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
			RawDescriptor: file_google_geo_type_viewport_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_geo_type_viewport_proto_goTypes,
		DependencyIndexes: file_google_geo_type_viewport_proto_depIdxs,
		MessageInfos:      file_google_geo_type_viewport_proto_msgTypes,
	}.Build()
	File_google_geo_type_viewport_proto = out.File
	file_google_geo_type_viewport_proto_rawDesc = nil
	file_google_geo_type_viewport_proto_goTypes = nil
	file_google_geo_type_viewport_proto_depIdxs = nil
}
