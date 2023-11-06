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

package grpcutil

import (
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConvertProtoToGrpcError converts a proto Error into a gRPC error.
func ConvertProtoToGrpcError(err *v1alpha1.Error) error {
	if err == nil {
		return nil
	}
	return status.ErrorProto(&statuspb.Status{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	})
}

// ConvertGrpcToProtoError converts the given gRPC error to a
// proto Error message. If err is nil, the function will also
// return nil.
func ConvertGrpcToProtoError(err error) *v1alpha1.Error {
	if err == nil {
		return nil
	}
	status, ok := status.FromError(err)
	if !ok {
		// If the given error is not a gRPC error, return unknown
		return &v1alpha1.Error{
			Code:    int32(connect.CodeUnknown),
			Message: "provided error is not a gRPC error",
		}
	}
	protoErr := &v1alpha1.Error{
		Code:    int32(status.Code()),
		Message: status.Message(),
	}
	details := make([]*anypb.Any, 0, len(status.Details()))
	for _, any := range status.Details() {
		// status.Details() returns a slice of 'any' instead of anypb.Any
		// so, first convert to a proto message so that we can convert that to
		// an anypb.Any
		if pm, ok := any.(proto.Message); ok {
			detail, err := anypb.New(pm)
			if err != nil {
				details = append(details, detail)
			}
		}
	}
	protoErr.Details = details
	return protoErr
}
