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
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
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
	stat, _ := status.FromError(err)
	statProto := stat.Proto()
	return &v1alpha1.Error{
		Code:    int32(stat.Code()),
		Message: stat.Message(),
		Details: statProto.Details,
	}
}
