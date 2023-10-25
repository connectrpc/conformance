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
// source: connectrpc/conformance/v1alpha1/config.proto

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

type HTTPVersion int32

const (
	HTTPVersion_HTTP_VERSION_UNSPECIFIED HTTPVersion = 0
	HTTPVersion_HTTP_VERSION_1           HTTPVersion = 1
	HTTPVersion_HTTP_VERSION_2           HTTPVersion = 2
	HTTPVersion_HTTP_VERSION_3           HTTPVersion = 3
)

// Enum value maps for HTTPVersion.
var (
	HTTPVersion_name = map[int32]string{
		0: "HTTP_VERSION_UNSPECIFIED",
		1: "HTTP_VERSION_1",
		2: "HTTP_VERSION_2",
		3: "HTTP_VERSION_3",
	}
	HTTPVersion_value = map[string]int32{
		"HTTP_VERSION_UNSPECIFIED": 0,
		"HTTP_VERSION_1":           1,
		"HTTP_VERSION_2":           2,
		"HTTP_VERSION_3":           3,
	}
)

func (x HTTPVersion) Enum() *HTTPVersion {
	p := new(HTTPVersion)
	*p = x
	return p
}

func (x HTTPVersion) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HTTPVersion) Descriptor() protoreflect.EnumDescriptor {
	return file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[0].Descriptor()
}

func (HTTPVersion) Type() protoreflect.EnumType {
	return &file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[0]
}

func (x HTTPVersion) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HTTPVersion.Descriptor instead.
func (HTTPVersion) EnumDescriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{0}
}

type Protocol int32

const (
	Protocol_PROTOCOL_UNSPECIFIED Protocol = 0
	Protocol_PROTOCOL_CONNECT     Protocol = 1
	Protocol_PROTOCOL_GRPC        Protocol = 2
	Protocol_PROTOCOL_GRPC_WEB    Protocol = 3
)

// Enum value maps for Protocol.
var (
	Protocol_name = map[int32]string{
		0: "PROTOCOL_UNSPECIFIED",
		1: "PROTOCOL_CONNECT",
		2: "PROTOCOL_GRPC",
		3: "PROTOCOL_GRPC_WEB",
	}
	Protocol_value = map[string]int32{
		"PROTOCOL_UNSPECIFIED": 0,
		"PROTOCOL_CONNECT":     1,
		"PROTOCOL_GRPC":        2,
		"PROTOCOL_GRPC_WEB":    3,
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
	return file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[1].Descriptor()
}

func (Protocol) Type() protoreflect.EnumType {
	return &file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[1]
}

func (x Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Protocol.Descriptor instead.
func (Protocol) EnumDescriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{1}
}

type Codec int32

const (
	Codec_CODEC_UNSPECIFIED Codec = 0
	Codec_CODEC_PROTO       Codec = 1
	Codec_CODEC_JSON        Codec = 2
	Codec_CODEC_TEXT        Codec = 3
)

// Enum value maps for Codec.
var (
	Codec_name = map[int32]string{
		0: "CODEC_UNSPECIFIED",
		1: "CODEC_PROTO",
		2: "CODEC_JSON",
		3: "CODEC_TEXT",
	}
	Codec_value = map[string]int32{
		"CODEC_UNSPECIFIED": 0,
		"CODEC_PROTO":       1,
		"CODEC_JSON":        2,
		"CODEC_TEXT":        3,
	}
)

func (x Codec) Enum() *Codec {
	p := new(Codec)
	*p = x
	return p
}

func (x Codec) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Codec) Descriptor() protoreflect.EnumDescriptor {
	return file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[2].Descriptor()
}

func (Codec) Type() protoreflect.EnumType {
	return &file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[2]
}

func (x Codec) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Codec.Descriptor instead.
func (Codec) EnumDescriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{2}
}

type Compression int32

const (
	Compression_COMPRESSION_UNSPECIFIED Compression = 0
	Compression_COMPRESSION_IDENTITY    Compression = 1
	Compression_COMPRESSION_GZIP        Compression = 2
	Compression_COMPRESSION_BR          Compression = 3
	Compression_COMPRESSION_ZSTD        Compression = 4
	Compression_COMPRESSION_DEFLATE     Compression = 5
	Compression_COMPRESSION_SNAPPY      Compression = 6
)

// Enum value maps for Compression.
var (
	Compression_name = map[int32]string{
		0: "COMPRESSION_UNSPECIFIED",
		1: "COMPRESSION_IDENTITY",
		2: "COMPRESSION_GZIP",
		3: "COMPRESSION_BR",
		4: "COMPRESSION_ZSTD",
		5: "COMPRESSION_DEFLATE",
		6: "COMPRESSION_SNAPPY",
	}
	Compression_value = map[string]int32{
		"COMPRESSION_UNSPECIFIED": 0,
		"COMPRESSION_IDENTITY":    1,
		"COMPRESSION_GZIP":        2,
		"COMPRESSION_BR":          3,
		"COMPRESSION_ZSTD":        4,
		"COMPRESSION_DEFLATE":     5,
		"COMPRESSION_SNAPPY":      6,
	}
)

func (x Compression) Enum() *Compression {
	p := new(Compression)
	*p = x
	return p
}

func (x Compression) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Compression) Descriptor() protoreflect.EnumDescriptor {
	return file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[3].Descriptor()
}

func (Compression) Type() protoreflect.EnumType {
	return &file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[3]
}

func (x Compression) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Compression.Descriptor instead.
func (Compression) EnumDescriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{3}
}

type StreamType int32

const (
	StreamType_STREAM_TYPE_UNSPECIFIED             StreamType = 0
	StreamType_STREAM_TYPE_UNARY                   StreamType = 1
	StreamType_STREAM_TYPE_CLIENT_STREAM           StreamType = 2
	StreamType_STREAM_TYPE_SERVER_STREAM           StreamType = 3
	StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM StreamType = 4
	StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM StreamType = 5
)

// Enum value maps for StreamType.
var (
	StreamType_name = map[int32]string{
		0: "STREAM_TYPE_UNSPECIFIED",
		1: "STREAM_TYPE_UNARY",
		2: "STREAM_TYPE_CLIENT_STREAM",
		3: "STREAM_TYPE_SERVER_STREAM",
		4: "STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM",
		5: "STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM",
	}
	StreamType_value = map[string]int32{
		"STREAM_TYPE_UNSPECIFIED":             0,
		"STREAM_TYPE_UNARY":                   1,
		"STREAM_TYPE_CLIENT_STREAM":           2,
		"STREAM_TYPE_SERVER_STREAM":           3,
		"STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM": 4,
		"STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM": 5,
	}
)

func (x StreamType) Enum() *StreamType {
	p := new(StreamType)
	*p = x
	return p
}

func (x StreamType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StreamType) Descriptor() protoreflect.EnumDescriptor {
	return file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[4].Descriptor()
}

func (StreamType) Type() protoreflect.EnumType {
	return &file_connectrpc_conformance_v1alpha1_config_proto_enumTypes[4]
}

func (x StreamType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StreamType.Descriptor instead.
func (StreamType) EnumDescriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{4}
}

// Config defines the configuration for running conformance tests.
// This enumerates all of the "flavors" of the test suite to run.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The features supported by the client or server under test.
	// This is used to filter the set of test cases that are run.
	// If absent, an empty message is used. See Features for more
	// on how empty/absent fields are interpreted.
	Features *Features `protobuf:"bytes,1,opt,name=features,proto3" json:"features,omitempty"`
	// This can indicate additional permutations that are supported
	// that might otherwise be excluded based on the above features.
	IncludeCases []*ConfigCase `protobuf:"bytes,2,rep,name=include_cases,json=includeCases,proto3" json:"include_cases,omitempty"`
	// This can indicates permutations that are not supported even
	// though their support might be implied by the above features.
	ExcludeCases []*ConfigCase `protobuf:"bytes,3,rep,name=exclude_cases,json=excludeCases,proto3" json:"exclude_cases,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetFeatures() *Features {
	if x != nil {
		return x.Features
	}
	return nil
}

func (x *Config) GetIncludeCases() []*ConfigCase {
	if x != nil {
		return x.IncludeCases
	}
	return nil
}

func (x *Config) GetExcludeCases() []*ConfigCase {
	if x != nil {
		return x.ExcludeCases
	}
	return nil
}

type Features struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If empty, HTTP 1.1 and HTTP/2 are assumed.
	Versions []HTTPVersion `protobuf:"varint,1,rep,packed,name=versions,proto3,enum=connectrpc.conformance.v1alpha1.HTTPVersion" json:"versions,omitempty"`
	// If empty, all three are assumed: Connect, gRPC, and gRPC-Web.
	Protocols []Protocol `protobuf:"varint,2,rep,packed,name=protocols,proto3,enum=connectrpc.conformance.v1alpha1.Protocol" json:"protocols,omitempty"`
	// If empty, "proto" and "json" are assumed.
	Codecs []Codec `protobuf:"varint,3,rep,packed,name=codecs,proto3,enum=connectrpc.conformance.v1alpha1.Codec" json:"codecs,omitempty"`
	// If empty, "gzip" alone is assumed.
	Compressions []Compression `protobuf:"varint,4,rep,packed,name=compressions,proto3,enum=connectrpc.conformance.v1alpha1.Compression" json:"compressions,omitempty"`
	// If absent, true is assumed.
	SupportsH2C *bool `protobuf:"varint,5,opt,name=supports_h2c,json=supportsH2c,proto3,oneof" json:"supports_h2c,omitempty"`
	// If absent, true is assumed.
	SupportsTls *bool `protobuf:"varint,6,opt,name=supports_tls,json=supportsTls,proto3,oneof" json:"supports_tls,omitempty"`
	// If absent, true is assumed. If false, implies that gRPC protocol is not allowed.
	SupportsTrailers *bool `protobuf:"varint,7,opt,name=supports_trailers,json=supportsTrailers,proto3,oneof" json:"supports_trailers,omitempty"`
	// If absent, false is assumed.
	SupportsHalfDuplexBidiOverHttp1 *bool `protobuf:"varint,8,opt,name=supports_half_duplex_bidi_over_http1,json=supportsHalfDuplexBidiOverHttp1,proto3,oneof" json:"supports_half_duplex_bidi_over_http1,omitempty"`
	// If absent, true is assumed.
	SupportsConnectGet *bool `protobuf:"varint,9,opt,name=supports_connect_get,json=supportsConnectGet,proto3,oneof" json:"supports_connect_get,omitempty"`
	// If absent, false is assumed.
	RequiresConnectVersionHeader *bool `protobuf:"varint,10,opt,name=requires_connect_version_header,json=requiresConnectVersionHeader,proto3,oneof" json:"requires_connect_version_header,omitempty"`
	// If empty, all stream types are assumed. This is usually for
	// clients, since some client environments may not be able to
	// support certain kinds of streaming operations, especially
	// bidirectional streams.
	StreamTypes []StreamType `protobuf:"varint,11,rep,packed,name=stream_types,json=streamTypes,proto3,enum=connectrpc.conformance.v1alpha1.StreamType" json:"stream_types,omitempty"`
}

func (x *Features) Reset() {
	*x = Features{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Features) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Features) ProtoMessage() {}

func (x *Features) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Features.ProtoReflect.Descriptor instead.
func (*Features) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{1}
}

func (x *Features) GetVersions() []HTTPVersion {
	if x != nil {
		return x.Versions
	}
	return nil
}

func (x *Features) GetProtocols() []Protocol {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *Features) GetCodecs() []Codec {
	if x != nil {
		return x.Codecs
	}
	return nil
}

func (x *Features) GetCompressions() []Compression {
	if x != nil {
		return x.Compressions
	}
	return nil
}

func (x *Features) GetSupportsH2C() bool {
	if x != nil && x.SupportsH2C != nil {
		return *x.SupportsH2C
	}
	return false
}

func (x *Features) GetSupportsTls() bool {
	if x != nil && x.SupportsTls != nil {
		return *x.SupportsTls
	}
	return false
}

func (x *Features) GetSupportsTrailers() bool {
	if x != nil && x.SupportsTrailers != nil {
		return *x.SupportsTrailers
	}
	return false
}

func (x *Features) GetSupportsHalfDuplexBidiOverHttp1() bool {
	if x != nil && x.SupportsHalfDuplexBidiOverHttp1 != nil {
		return *x.SupportsHalfDuplexBidiOverHttp1
	}
	return false
}

func (x *Features) GetSupportsConnectGet() bool {
	if x != nil && x.SupportsConnectGet != nil {
		return *x.SupportsConnectGet
	}
	return false
}

func (x *Features) GetRequiresConnectVersionHeader() bool {
	if x != nil && x.RequiresConnectVersionHeader != nil {
		return *x.RequiresConnectVersionHeader
	}
	return false
}

func (x *Features) GetStreamTypes() []StreamType {
	if x != nil {
		return x.StreamTypes
	}
	return nil
}

type ConfigCase struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If unspecified, indicates cases for all versions.
	Version HTTPVersion `protobuf:"varint,1,opt,name=version,proto3,enum=connectrpc.conformance.v1alpha1.HTTPVersion" json:"version,omitempty"`
	// If unspecified, indicates cases for all protocols.
	Protocol Protocol `protobuf:"varint,2,opt,name=protocol,proto3,enum=connectrpc.conformance.v1alpha1.Protocol" json:"protocol,omitempty"`
	// If unspecified, indicates cases for all codecs.
	Codec Codec `protobuf:"varint,3,opt,name=codec,proto3,enum=connectrpc.conformance.v1alpha1.Codec" json:"codec,omitempty"`
	// If unspecified, indicates cases for all compression algorithms.
	Compression Compression `protobuf:"varint,4,opt,name=compression,proto3,enum=connectrpc.conformance.v1alpha1.Compression" json:"compression,omitempty"`
	UseTls      bool        `protobuf:"varint,5,opt,name=use_tls,json=useTls,proto3" json:"use_tls,omitempty"`
	// If unspecified, indicates cases for all stream types.
	StreamType StreamType `protobuf:"varint,6,opt,name=stream_type,json=streamType,proto3,enum=connectrpc.conformance.v1alpha1.StreamType" json:"stream_type,omitempty"`
}

func (x *ConfigCase) Reset() {
	*x = ConfigCase{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigCase) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigCase) ProtoMessage() {}

func (x *ConfigCase) ProtoReflect() protoreflect.Message {
	mi := &file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigCase.ProtoReflect.Descriptor instead.
func (*ConfigCase) Descriptor() ([]byte, []int) {
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP(), []int{2}
}

func (x *ConfigCase) GetVersion() HTTPVersion {
	if x != nil {
		return x.Version
	}
	return HTTPVersion_HTTP_VERSION_UNSPECIFIED
}

func (x *ConfigCase) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

func (x *ConfigCase) GetCodec() Codec {
	if x != nil {
		return x.Codec
	}
	return Codec_CODEC_UNSPECIFIED
}

func (x *ConfigCase) GetCompression() Compression {
	if x != nil {
		return x.Compression
	}
	return Compression_COMPRESSION_UNSPECIFIED
}

func (x *ConfigCase) GetUseTls() bool {
	if x != nil {
		return x.UseTls
	}
	return false
}

func (x *ConfigCase) GetStreamType() StreamType {
	if x != nil {
		return x.StreamType
	}
	return StreamType_STREAM_TYPE_UNSPECIFIED
}

var File_connectrpc_conformance_v1alpha1_config_proto protoreflect.FileDescriptor

var file_connectrpc_conformance_v1alpha1_config_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22,
	0xf3, 0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x45, 0x0a, 0x08, 0x66, 0x65,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x46,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x52, 0x08, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x73, 0x12, 0x50, 0x0a, 0x0d, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x63, 0x61, 0x73,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x43, 0x61, 0x73, 0x65, 0x52, 0x0c, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x43, 0x61,
	0x73, 0x65, 0x73, 0x12, 0x50, 0x0a, 0x0d, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x63,
	0x61, 0x73, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61,
	0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x43, 0x61, 0x73, 0x65, 0x52, 0x0c, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65,
	0x43, 0x61, 0x73, 0x65, 0x73, 0x22, 0xf6, 0x06, 0x0a, 0x08, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x73, 0x12, 0x48, 0x0a, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x47, 0x0a, 0x09,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0e, 0x32,
	0x29, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x12, 0x3e, 0x0a, 0x06, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72,
	0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x63, 0x52, 0x06, 0x63,
	0x6f, 0x64, 0x65, 0x63, 0x73, 0x12, 0x50, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f,
	0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x63, 0x6f, 0x6d, 0x70, 0x72,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x26, 0x0a, 0x0c, 0x73, 0x75, 0x70, 0x70, 0x6f,
	0x72, 0x74, 0x73, 0x5f, 0x68, 0x32, 0x63, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52,
	0x0b, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x48, 0x32, 0x63, 0x88, 0x01, 0x01, 0x12,
	0x26, 0x0a, 0x0c, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x74, 0x6c, 0x73, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x08, 0x48, 0x01, 0x52, 0x0b, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74,
	0x73, 0x54, 0x6c, 0x73, 0x88, 0x01, 0x01, 0x12, 0x30, 0x0a, 0x11, 0x73, 0x75, 0x70, 0x70, 0x6f,
	0x72, 0x74, 0x73, 0x5f, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x65, 0x72, 0x73, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x08, 0x48, 0x02, 0x52, 0x10, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x54, 0x72,
	0x61, 0x69, 0x6c, 0x65, 0x72, 0x73, 0x88, 0x01, 0x01, 0x12, 0x52, 0x0a, 0x24, 0x73, 0x75, 0x70,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x68, 0x61, 0x6c, 0x66, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x65,
	0x78, 0x5f, 0x62, 0x69, 0x64, 0x69, 0x5f, 0x6f, 0x76, 0x65, 0x72, 0x5f, 0x68, 0x74, 0x74, 0x70,
	0x31, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x48, 0x03, 0x52, 0x1f, 0x73, 0x75, 0x70, 0x70, 0x6f,
	0x72, 0x74, 0x73, 0x48, 0x61, 0x6c, 0x66, 0x44, 0x75, 0x70, 0x6c, 0x65, 0x78, 0x42, 0x69, 0x64,
	0x69, 0x4f, 0x76, 0x65, 0x72, 0x48, 0x74, 0x74, 0x70, 0x31, 0x88, 0x01, 0x01, 0x12, 0x35, 0x0a,
	0x14, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x5f, 0x67, 0x65, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x48, 0x04, 0x52, 0x12, 0x73,
	0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x47, 0x65,
	0x74, 0x88, 0x01, 0x01, 0x12, 0x4a, 0x0a, 0x1f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x73,
	0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x48, 0x05, 0x52,
	0x1c, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x73, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x88, 0x01, 0x01,
	0x12, 0x4e, 0x0a, 0x0c, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x18, 0x0b, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x0b, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x73,
	0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x68, 0x32,
	0x63, 0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x74,
	0x6c, 0x73, 0x42, 0x14, 0x0a, 0x12, 0x5f, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f,
	0x74, 0x72, 0x61, 0x69, 0x6c, 0x65, 0x72, 0x73, 0x42, 0x27, 0x0a, 0x25, 0x5f, 0x73, 0x75, 0x70,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x68, 0x61, 0x6c, 0x66, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x65,
	0x78, 0x5f, 0x62, 0x69, 0x64, 0x69, 0x5f, 0x6f, 0x76, 0x65, 0x72, 0x5f, 0x68, 0x74, 0x74, 0x70,
	0x31, 0x42, 0x17, 0x0a, 0x15, 0x5f, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x5f, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x67, 0x65, 0x74, 0x42, 0x22, 0x0a, 0x20, 0x5f, 0x72,
	0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x22, 0x90,
	0x03, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x43, 0x61, 0x73, 0x65, 0x12, 0x46, 0x0a,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x48, 0x54, 0x54, 0x50, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x45, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x3c, 0x0a, 0x05,
	0x63, 0x6f, 0x64, 0x65, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6f,
	0x64, 0x65, 0x63, 0x52, 0x05, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x12, 0x4e, 0x0a, 0x0b, 0x63, 0x6f,
	0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x2c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x63,
	0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73,
	0x65, 0x5f, 0x74, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x75, 0x73, 0x65,
	0x54, 0x6c, 0x73, 0x12, 0x4c, 0x0a, 0x0b, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70,
	0x65, 0x2a, 0x67, 0x0a, 0x0b, 0x48, 0x54, 0x54, 0x50, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1c, 0x0a, 0x18, 0x48, 0x54, 0x54, 0x50, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x12,
	0x0a, 0x0e, 0x48, 0x54, 0x54, 0x50, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x31,
	0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x48, 0x54, 0x54, 0x50, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49,
	0x4f, 0x4e, 0x5f, 0x32, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x48, 0x54, 0x54, 0x50, 0x5f, 0x56,
	0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x33, 0x10, 0x03, 0x2a, 0x64, 0x0a, 0x08, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x18, 0x0a, 0x14, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43,
	0x4f, 0x4c, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x14, 0x0a, 0x10, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x43, 0x4f, 0x4e,
	0x4e, 0x45, 0x43, 0x54, 0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43,
	0x4f, 0x4c, 0x5f, 0x47, 0x52, 0x50, 0x43, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x50, 0x52, 0x4f,
	0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x47, 0x52, 0x50, 0x43, 0x5f, 0x57, 0x45, 0x42, 0x10, 0x03,
	0x2a, 0x4f, 0x0a, 0x05, 0x43, 0x6f, 0x64, 0x65, 0x63, 0x12, 0x15, 0x0a, 0x11, 0x43, 0x4f, 0x44,
	0x45, 0x43, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x4f, 0x44, 0x45, 0x43, 0x5f, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x10,
	0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x44, 0x45, 0x43, 0x5f, 0x4a, 0x53, 0x4f, 0x4e, 0x10,
	0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x44, 0x45, 0x43, 0x5f, 0x54, 0x45, 0x58, 0x54, 0x10,
	0x03, 0x2a, 0xb5, 0x01, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x1b, 0x0a, 0x17, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x18,
	0x0a, 0x14, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x49, 0x44,
	0x45, 0x4e, 0x54, 0x49, 0x54, 0x59, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x4d, 0x50,
	0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x47, 0x5a, 0x49, 0x50, 0x10, 0x02, 0x12, 0x12,
	0x0a, 0x0e, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x42, 0x52,
	0x10, 0x03, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f,
	0x4e, 0x5f, 0x5a, 0x53, 0x54, 0x44, 0x10, 0x04, 0x12, 0x17, 0x0a, 0x13, 0x43, 0x4f, 0x4d, 0x50,
	0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x44, 0x45, 0x46, 0x4c, 0x41, 0x54, 0x45, 0x10,
	0x05, 0x12, 0x16, 0x0a, 0x12, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e,
	0x5f, 0x53, 0x4e, 0x41, 0x50, 0x50, 0x59, 0x10, 0x06, 0x2a, 0xd0, 0x01, 0x0a, 0x0a, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x17, 0x53, 0x54, 0x52, 0x45,
	0x41, 0x4d, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46,
	0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x41, 0x52, 0x59, 0x10, 0x01, 0x12, 0x1d, 0x0a, 0x19,
	0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x4c, 0x49, 0x45,
	0x4e, 0x54, 0x5f, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x10, 0x02, 0x12, 0x1d, 0x0a, 0x19, 0x53,
	0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x45,
	0x52, 0x5f, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x10, 0x03, 0x12, 0x27, 0x0a, 0x23, 0x53, 0x54,
	0x52, 0x45, 0x41, 0x4d, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x48, 0x41, 0x4c, 0x46, 0x5f, 0x44,
	0x55, 0x50, 0x4c, 0x45, 0x58, 0x5f, 0x42, 0x49, 0x44, 0x49, 0x5f, 0x53, 0x54, 0x52, 0x45, 0x41,
	0x4d, 0x10, 0x04, 0x12, 0x27, 0x0a, 0x23, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x46, 0x55, 0x4c, 0x4c, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x45, 0x58, 0x5f, 0x42,
	0x49, 0x44, 0x49, 0x5f, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x10, 0x05, 0x42, 0xb6, 0x02, 0x0a,
	0x23, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x42, 0x0b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x64, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63,
	0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63,
	0x65, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x43, 0x58, 0xaa,
	0x02, 0x1f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0xca, 0x02, 0x1f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x5c, 0x43,
	0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0xe2, 0x02, 0x2b, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63,
	0x5c, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x21, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70, 0x63, 0x3a, 0x3a,
	0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_connectrpc_conformance_v1alpha1_config_proto_rawDescOnce sync.Once
	file_connectrpc_conformance_v1alpha1_config_proto_rawDescData = file_connectrpc_conformance_v1alpha1_config_proto_rawDesc
)

func file_connectrpc_conformance_v1alpha1_config_proto_rawDescGZIP() []byte {
	file_connectrpc_conformance_v1alpha1_config_proto_rawDescOnce.Do(func() {
		file_connectrpc_conformance_v1alpha1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_connectrpc_conformance_v1alpha1_config_proto_rawDescData)
	})
	return file_connectrpc_conformance_v1alpha1_config_proto_rawDescData
}

var file_connectrpc_conformance_v1alpha1_config_proto_enumTypes = make([]protoimpl.EnumInfo, 5)
var file_connectrpc_conformance_v1alpha1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_connectrpc_conformance_v1alpha1_config_proto_goTypes = []interface{}{
	(HTTPVersion)(0),   // 0: connectrpc.conformance.v1alpha1.HTTPVersion
	(Protocol)(0),      // 1: connectrpc.conformance.v1alpha1.Protocol
	(Codec)(0),         // 2: connectrpc.conformance.v1alpha1.Codec
	(Compression)(0),   // 3: connectrpc.conformance.v1alpha1.Compression
	(StreamType)(0),    // 4: connectrpc.conformance.v1alpha1.StreamType
	(*Config)(nil),     // 5: connectrpc.conformance.v1alpha1.Config
	(*Features)(nil),   // 6: connectrpc.conformance.v1alpha1.Features
	(*ConfigCase)(nil), // 7: connectrpc.conformance.v1alpha1.ConfigCase
}
var file_connectrpc_conformance_v1alpha1_config_proto_depIdxs = []int32{
	6,  // 0: connectrpc.conformance.v1alpha1.Config.features:type_name -> connectrpc.conformance.v1alpha1.Features
	7,  // 1: connectrpc.conformance.v1alpha1.Config.include_cases:type_name -> connectrpc.conformance.v1alpha1.ConfigCase
	7,  // 2: connectrpc.conformance.v1alpha1.Config.exclude_cases:type_name -> connectrpc.conformance.v1alpha1.ConfigCase
	0,  // 3: connectrpc.conformance.v1alpha1.Features.versions:type_name -> connectrpc.conformance.v1alpha1.HTTPVersion
	1,  // 4: connectrpc.conformance.v1alpha1.Features.protocols:type_name -> connectrpc.conformance.v1alpha1.Protocol
	2,  // 5: connectrpc.conformance.v1alpha1.Features.codecs:type_name -> connectrpc.conformance.v1alpha1.Codec
	3,  // 6: connectrpc.conformance.v1alpha1.Features.compressions:type_name -> connectrpc.conformance.v1alpha1.Compression
	4,  // 7: connectrpc.conformance.v1alpha1.Features.stream_types:type_name -> connectrpc.conformance.v1alpha1.StreamType
	0,  // 8: connectrpc.conformance.v1alpha1.ConfigCase.version:type_name -> connectrpc.conformance.v1alpha1.HTTPVersion
	1,  // 9: connectrpc.conformance.v1alpha1.ConfigCase.protocol:type_name -> connectrpc.conformance.v1alpha1.Protocol
	2,  // 10: connectrpc.conformance.v1alpha1.ConfigCase.codec:type_name -> connectrpc.conformance.v1alpha1.Codec
	3,  // 11: connectrpc.conformance.v1alpha1.ConfigCase.compression:type_name -> connectrpc.conformance.v1alpha1.Compression
	4,  // 12: connectrpc.conformance.v1alpha1.ConfigCase.stream_type:type_name -> connectrpc.conformance.v1alpha1.StreamType
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_connectrpc_conformance_v1alpha1_config_proto_init() }
func file_connectrpc_conformance_v1alpha1_config_proto_init() {
	if File_connectrpc_conformance_v1alpha1_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Features); i {
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
		file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfigCase); i {
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
	file_connectrpc_conformance_v1alpha1_config_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connectrpc_conformance_v1alpha1_config_proto_rawDesc,
			NumEnums:      5,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_connectrpc_conformance_v1alpha1_config_proto_goTypes,
		DependencyIndexes: file_connectrpc_conformance_v1alpha1_config_proto_depIdxs,
		EnumInfos:         file_connectrpc_conformance_v1alpha1_config_proto_enumTypes,
		MessageInfos:      file_connectrpc_conformance_v1alpha1_config_proto_msgTypes,
	}.Build()
	File_connectrpc_conformance_v1alpha1_config_proto = out.File
	file_connectrpc_conformance_v1alpha1_config_proto_rawDesc = nil
	file_connectrpc_conformance_v1alpha1_config_proto_goTypes = nil
	file_connectrpc_conformance_v1alpha1_config_proto_depIdxs = nil
}
