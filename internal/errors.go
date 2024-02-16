// Copyright 2023-2024 The Connect Authors
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

package internal

import (
	"errors"
	"fmt"
	"strings"

	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConvertErrorToConnectError converts the given error to a Connect error
// If err is nil, function will also return nil. If err is not
// of type connect.Error, a Connect error of code Unknown is returned.
func ConvertErrorToConnectError(err error) *connect.Error {
	if err == nil {
		return nil
	}
	connectErr := new(connect.Error)
	if !errors.As(err, &connectErr) {
		connectErr = connect.NewError(connect.CodeUnknown, err)
	}
	return connectErr
}

// ConvertErrorToProtoError converts the given error to a proto Error
// If err is nil, function will also return nil. If err is not
// of type connect.Error, a code representing Unknown is returned.
func ConvertErrorToProtoError(err error) *v1.Error {
	if err == nil {
		return nil
	}
	connectErr := new(connect.Error)
	if !errors.As(err, &connectErr) {
		return &v1.Error{
			Code:    v1.Code(int32(connect.CodeUnknown)),
			Message: proto.String(err.Error()),
		}
	}
	return ConvertConnectToProtoError(connectErr)
}

// ConvertConnectToProtoError converts the given Connect error to a
// proto Error message. If err is nil, the function will also
// return nil.
func ConvertConnectToProtoError(err *connect.Error) *v1.Error {
	if err == nil {
		return nil
	}
	protoErr := &v1.Error{
		Code:    v1.Code(int32(err.Code())),
		Message: proto.String(err.Message()),
	}
	details := make([]*anypb.Any, 0, len(err.Details()))
	for _, detail := range err.Details() {
		details = append(details, &anypb.Any{
			// Connect Go strips the prefix from the type when calling Type()
			// but anypb.MarshalFrom adds the prefix explicitly. Since Protoyaml
			// uses anypb.MarshalFrom when reading an Any type from a yaml file,
			// it must be explicitly added back here so that we can successfully
			// compare the expected response from the yaml file into what
			// Connect Go returns.
			TypeUrl: DefaultAnyResolverPrefix + detail.Type(),
			Value:   detail.Bytes(),
		})
	}
	protoErr.Details = details
	return protoErr
}

// ConvertProtoToConnectError creates a Connect error from the given proto Error message.
func ConvertProtoToConnectError(err *v1.Error) *connect.Error {
	if err == nil {
		return nil
	}
	connectErr := connect.NewError(connect.Code(int32(err.Code)), errors.New(err.GetMessage()))
	for _, detail := range err.Details {
		connectDetail, err := connect.NewErrorDetail(detail)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		connectErr.AddDetail(connectDetail)
	}
	return connectErr
}

// EnsureFileName ensures that the given error includes the given filename. If it
// does not, it wraps the error in one that does include the filename. This is
// used to ensure that file-system-specific errors have good messages and
// unambiguously indicate which file was the cause of the error.
func EnsureFileName(err error, filename string) error {
	if strings.Contains(err.Error(), filename) {
		return err // already contains filename, nothing else to do
	}
	return fmt.Errorf("%s: %w", filename, err)
}
