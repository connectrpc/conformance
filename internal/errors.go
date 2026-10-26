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

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectproto"
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
		connectErr = connect.NewError(connect.CodeUnknown, err.Error()).WithCause(err)
	}
	return connectErr
}

// ConvertErrorToProtoError converts the given error to a proto Error
// If err is nil, function will also return nil. If err is not
// of type connect.Error, a code representing Unknown is returned.
func ConvertErrorToProtoError(err error) *conformancev1.Error {
	if err == nil {
		return nil
	}
	connectErr := new(connect.Error)
	if !errors.As(err, &connectErr) {
		return &conformancev1.Error{
			Code:    conformancev1.Code_CODE_UNKNOWN,
			Message: new(err.Error()),
		}
	}
	return ConvertConnectToProtoError(connectErr)
}

// ConvertConnectToProtoError converts the given Connect error to a
// proto Error message. If err is nil, the function will also
// return nil.
func ConvertConnectToProtoError(err *connect.Error) *conformancev1.Error {
	if err == nil {
		return nil
	}
	protoErr := &conformancev1.Error{
		Code:    conformancev1.Code(int32(err.Code())),
		Message: new(err.Message()),
	}
	details := make([]*anypb.Any, 0, len(err.Details()))
	for _, detail := range err.Details() {
		// ErrorDetailToAny restores the type URL prefix that protoyaml adds
		// when reading expected results, so the two compare equal.
		details = append(details, connectproto.ErrorDetailToAny(detail))
	}
	protoErr.Details = details
	return protoErr
}

// ConvertProtoToConnectError creates a Connect error from the given proto Error message.
func ConvertProtoToConnectError(err *conformancev1.Error) *connect.Error {
	if err == nil {
		return nil
	}
	connectErr := connect.NewError(connect.Code(err.Code), err.GetMessage())
	for _, detail := range err.Details {
		connectDetail, err := connectproto.NewErrorDetail(detail)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
		}
		connectErr = connectErr.WithDetail(connectDetail)
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
