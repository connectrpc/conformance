package app

import (
	"errors"

	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/anypb"
)

func AsConnectError(err error) *connect.Error {
	connectErr := new(connect.Error)

	if !errors.As(err, &connectErr) {
		connectErr = connect.NewError(connect.CodeUnknown, err)
	}
	return connectErr
}

func ConvertToProtoError(err *connect.Error) *v1alpha1.Error {
	protoErr := &v1alpha1.Error{}
	protoErr.Code = int32(err.Code())
	protoErr.Message = err.Message()
	details := make([]*anypb.Any, 0, len(err.Details()))
	for _, detail := range err.Details() {
		asAny := &anypb.Any{
			TypeUrl: detail.Type(),
			Value:   detail.Bytes(),
		}
		details = append(details, asAny)
	}
	protoErr.Details = details
	return protoErr
}

// ConvertToConnectError creates a Connect error from the given proto Error message
func ConvertToConnectError(err *v1alpha1.Error) *connect.Error {
	connectErr := connect.NewError(connect.Code(err.Code), errors.New(err.Message))
	for _, detail := range err.Details {
		connectDetail, err := connect.NewErrorDetail(detail)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		connectErr.AddDetail(connectDetail)
	}
	return connectErr
}
