// Copyright 2016 Google Inc.
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
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: google/bytestream/bytestream.proto

package bytestream

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

// Request object for ByteStream.Read.
type ReadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the resource to read.
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// The offset for the first byte to return in the read, relative to the start
	// of the resource.
	//
	// A `read_offset` that is negative or greater than the size of the resource
	// will cause an `OUT_OF_RANGE` error.
	ReadOffset int64 `protobuf:"varint,2,opt,name=read_offset,json=readOffset,proto3" json:"read_offset,omitempty"`
	// The maximum number of `data` bytes the server is allowed to return in the
	// sum of all `ReadResponse` messages. A `read_limit` of zero indicates that
	// there is no limit, and a negative `read_limit` will cause an error.
	//
	// If the stream returns fewer bytes than allowed by the `read_limit` and no
	// error occurred, the stream includes all data from the `read_offset` to the
	// end of the resource.
	ReadLimit int64 `protobuf:"varint,3,opt,name=read_limit,json=readLimit,proto3" json:"read_limit,omitempty"`
}

func (x *ReadRequest) Reset() {
	*x = ReadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadRequest) ProtoMessage() {}

func (x *ReadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadRequest.ProtoReflect.Descriptor instead.
func (*ReadRequest) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{0}
}

func (x *ReadRequest) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *ReadRequest) GetReadOffset() int64 {
	if x != nil {
		return x.ReadOffset
	}
	return 0
}

func (x *ReadRequest) GetReadLimit() int64 {
	if x != nil {
		return x.ReadLimit
	}
	return 0
}

// Response object for ByteStream.Read.
type ReadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A portion of the data for the resource. The service **may** leave `data`
	// empty for any given `ReadResponse`. This enables the service to inform the
	// client that the request is still live while it is running an operation to
	// generate more data.
	Data []byte `protobuf:"bytes,10,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ReadResponse) Reset() {
	*x = ReadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadResponse) ProtoMessage() {}

func (x *ReadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadResponse.ProtoReflect.Descriptor instead.
func (*ReadResponse) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{1}
}

func (x *ReadResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// Request object for ByteStream.Write.
type WriteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the resource to write. This **must** be set on the first
	// `WriteRequest` of each `Write()` action. If it is set on subsequent calls,
	// it **must** match the value of the first request.
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// The offset from the beginning of the resource at which the data should be
	// written. It is required on all `WriteRequest`s.
	//
	// In the first `WriteRequest` of a `Write()` action, it indicates
	// the initial offset for the `Write()` call. The value **must** be equal to
	// the `committed_size` that a call to `QueryWriteStatus()` would return.
	//
	// On subsequent calls, this value **must** be set and **must** be equal to
	// the sum of the first `write_offset` and the sizes of all `data` bundles
	// sent previously on this stream.
	//
	// An incorrect value will cause an error.
	WriteOffset int64 `protobuf:"varint,2,opt,name=write_offset,json=writeOffset,proto3" json:"write_offset,omitempty"`
	// If `true`, this indicates that the write is complete. Sending any
	// `WriteRequest`s subsequent to one in which `finish_write` is `true` will
	// cause an error.
	FinishWrite bool `protobuf:"varint,3,opt,name=finish_write,json=finishWrite,proto3" json:"finish_write,omitempty"`
	// A portion of the data for the resource. The client **may** leave `data`
	// empty for any given `WriteRequest`. This enables the client to inform the
	// service that the request is still live while it is running an operation to
	// generate more data.
	Data []byte `protobuf:"bytes,10,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *WriteRequest) Reset() {
	*x = WriteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteRequest) ProtoMessage() {}

func (x *WriteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteRequest.ProtoReflect.Descriptor instead.
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{2}
}

func (x *WriteRequest) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *WriteRequest) GetWriteOffset() int64 {
	if x != nil {
		return x.WriteOffset
	}
	return 0
}

func (x *WriteRequest) GetFinishWrite() bool {
	if x != nil {
		return x.FinishWrite
	}
	return false
}

func (x *WriteRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// Response object for ByteStream.Write.
type WriteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The number of bytes that have been processed for the given resource.
	CommittedSize int64 `protobuf:"varint,1,opt,name=committed_size,json=committedSize,proto3" json:"committed_size,omitempty"`
}

func (x *WriteResponse) Reset() {
	*x = WriteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteResponse) ProtoMessage() {}

func (x *WriteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteResponse.ProtoReflect.Descriptor instead.
func (*WriteResponse) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{3}
}

func (x *WriteResponse) GetCommittedSize() int64 {
	if x != nil {
		return x.CommittedSize
	}
	return 0
}

// Request object for ByteStream.QueryWriteStatus.
type QueryWriteStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the resource whose write status is being requested.
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
}

func (x *QueryWriteStatusRequest) Reset() {
	*x = QueryWriteStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryWriteStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryWriteStatusRequest) ProtoMessage() {}

func (x *QueryWriteStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryWriteStatusRequest.ProtoReflect.Descriptor instead.
func (*QueryWriteStatusRequest) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{4}
}

func (x *QueryWriteStatusRequest) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

// Response object for ByteStream.QueryWriteStatus.
type QueryWriteStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The number of bytes that have been processed for the given resource.
	CommittedSize int64 `protobuf:"varint,1,opt,name=committed_size,json=committedSize,proto3" json:"committed_size,omitempty"`
	// `complete` is `true` only if the client has sent a `WriteRequest` with
	// `finish_write` set to true, and the server has processed that request.
	Complete bool `protobuf:"varint,2,opt,name=complete,proto3" json:"complete,omitempty"`
}

func (x *QueryWriteStatusResponse) Reset() {
	*x = QueryWriteStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_bytestream_bytestream_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryWriteStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryWriteStatusResponse) ProtoMessage() {}

func (x *QueryWriteStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_bytestream_bytestream_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryWriteStatusResponse.ProtoReflect.Descriptor instead.
func (*QueryWriteStatusResponse) Descriptor() ([]byte, []int) {
	return file_google_bytestream_bytestream_proto_rawDescGZIP(), []int{5}
}

func (x *QueryWriteStatusResponse) GetCommittedSize() int64 {
	if x != nil {
		return x.CommittedSize
	}
	return 0
}

func (x *QueryWriteStatusResponse) GetComplete() bool {
	if x != nil {
		return x.Complete
	}
	return false
}

var File_google_bytestream_bytestream_proto protoreflect.FileDescriptor

var file_google_bytestream_bytestream_proto_rawDesc = []byte{
	0x0a, 0x22, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x2f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74,
	0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x22, 0x72, 0x0a, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x72,
	0x65, 0x61, 0x64, 0x5f, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0a, 0x72, 0x65, 0x61, 0x64, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x72, 0x65, 0x61, 0x64, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x72, 0x65, 0x61, 0x64, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x22, 0x0a, 0x0c, 0x52,
	0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22,
	0x8d, 0x01, 0x0a, 0x0c, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x6f,
	0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x77, 0x72, 0x69,
	0x74, 0x65, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x69, 0x6e, 0x69,
	0x73, 0x68, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b,
	0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22,
	0x36, 0x0a, 0x0d, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x69,
	0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x74, 0x65, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x3e, 0x0a, 0x17, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x5d, 0x0a, 0x18, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64,
	0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x63, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f,
	0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x63, 0x6f,
	0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x32, 0x92, 0x02, 0x0a, 0x0a, 0x42, 0x79, 0x74, 0x65, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x49, 0x0a, 0x04, 0x52, 0x65, 0x61, 0x64, 0x12, 0x1e, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01,
	0x12, 0x4c, 0x0a, 0x05, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x57, 0x72,
	0x69, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x57,
	0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x12, 0x6b,
	0x0a, 0x10, 0x51, 0x75, 0x65, 0x72, 0x79, 0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x2a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x57, 0x72, 0x69, 0x74,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xd1, 0x01, 0x0a, 0x15,
	0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x62, 0x79, 0x74, 0x65, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x42, 0x0f, 0x42, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x42, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x6e, 0x63, 0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0xa2, 0x02, 0x03, 0x47,
	0x42, 0x58, 0xaa, 0x02, 0x11, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0xca, 0x02, 0x11, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0xe2, 0x02, 0x1d, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x5c, 0x42, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x12, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x3a, 0x3a, 0x42, 0x79, 0x74, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_bytestream_bytestream_proto_rawDescOnce sync.Once
	file_google_bytestream_bytestream_proto_rawDescData = file_google_bytestream_bytestream_proto_rawDesc
)

func file_google_bytestream_bytestream_proto_rawDescGZIP() []byte {
	file_google_bytestream_bytestream_proto_rawDescOnce.Do(func() {
		file_google_bytestream_bytestream_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_bytestream_bytestream_proto_rawDescData)
	})
	return file_google_bytestream_bytestream_proto_rawDescData
}

var file_google_bytestream_bytestream_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_google_bytestream_bytestream_proto_goTypes = []interface{}{
	(*ReadRequest)(nil),              // 0: google.bytestream.ReadRequest
	(*ReadResponse)(nil),             // 1: google.bytestream.ReadResponse
	(*WriteRequest)(nil),             // 2: google.bytestream.WriteRequest
	(*WriteResponse)(nil),            // 3: google.bytestream.WriteResponse
	(*QueryWriteStatusRequest)(nil),  // 4: google.bytestream.QueryWriteStatusRequest
	(*QueryWriteStatusResponse)(nil), // 5: google.bytestream.QueryWriteStatusResponse
}
var file_google_bytestream_bytestream_proto_depIdxs = []int32{
	0, // 0: google.bytestream.ByteStream.Read:input_type -> google.bytestream.ReadRequest
	2, // 1: google.bytestream.ByteStream.Write:input_type -> google.bytestream.WriteRequest
	4, // 2: google.bytestream.ByteStream.QueryWriteStatus:input_type -> google.bytestream.QueryWriteStatusRequest
	1, // 3: google.bytestream.ByteStream.Read:output_type -> google.bytestream.ReadResponse
	3, // 4: google.bytestream.ByteStream.Write:output_type -> google.bytestream.WriteResponse
	5, // 5: google.bytestream.ByteStream.QueryWriteStatus:output_type -> google.bytestream.QueryWriteStatusResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_google_bytestream_bytestream_proto_init() }
func file_google_bytestream_bytestream_proto_init() {
	if File_google_bytestream_bytestream_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_bytestream_bytestream_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadRequest); i {
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
		file_google_bytestream_bytestream_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadResponse); i {
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
		file_google_bytestream_bytestream_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteRequest); i {
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
		file_google_bytestream_bytestream_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteResponse); i {
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
		file_google_bytestream_bytestream_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryWriteStatusRequest); i {
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
		file_google_bytestream_bytestream_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryWriteStatusResponse); i {
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
			RawDescriptor: file_google_bytestream_bytestream_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_google_bytestream_bytestream_proto_goTypes,
		DependencyIndexes: file_google_bytestream_bytestream_proto_depIdxs,
		MessageInfos:      file_google_bytestream_bytestream_proto_msgTypes,
	}.Build()
	File_google_bytestream_bytestream_proto = out.File
	file_google_bytestream_bytestream_proto_rawDesc = nil
	file_google_bytestream_bytestream_proto_goTypes = nil
	file_google_bytestream_bytestream_proto_depIdxs = nil
}
