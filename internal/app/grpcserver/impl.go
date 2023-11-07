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

package grpcserver

import (
	"context"

	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/conformance/internal/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// NewConformanceServiceServer creates a new Conformance Service server.
func NewConformanceServiceServer() v1alpha1.ConformanceServiceServer {
	return &conformanceServiceServer{}
}

type conformanceServiceServer struct {
	v1alpha1.UnimplementedConformanceServiceServer
}

func (c *conformanceServiceServer) Unary(
	ctx context.Context,
	req *v1alpha1.UnaryRequest,
) (*v1alpha1.UnaryResponse, error) {
	msgAsAny, err := asAny(req)
	if err != nil {
		return nil, err
	}

	if err := grpcutil.AddHeaderMetadata(ctx, req.ResponseDefinition.ResponseHeaders); err != nil {
		return nil, err
	}
	if err := grpcutil.AddTrailerMetadata(ctx, req.ResponseDefinition.ResponseTrailers); err != nil {
		return nil, err
	}

	md, _ := metadata.FromIncomingContext(ctx)
	payload, grpcErr := parseUnaryResponseDefinition(
		req.ResponseDefinition,
		md,
		[]*anypb.Any{msgAsAny},
	)
	if grpcErr != nil {
		return nil, grpcErr
	}

	return &v1alpha1.UnaryResponse{
		Payload: payload,
	}, nil
}

func (c *conformanceServiceServer) ClientStream(
	_ v1alpha1.ConformanceService_ClientStreamServer,
) error {
	// TODO - Implement ClientStream
	return status.Error(codes.Unimplemented, "method ClientStream not implemented")
}

func (c *conformanceServiceServer) ServerStream(
	_ *v1alpha1.ServerStreamRequest,
	_ v1alpha1.ConformanceService_ServerStreamServer,
) error {
	// TODO - Implement ServerStream
	return status.Error(codes.Unimplemented, "method ServerStream not implemented")
}

func (c *conformanceServiceServer) BidiStream(
	_ v1alpha1.ConformanceService_BidiStreamServer,
) error {
	// TODO - Implement BidiStream
	return status.Error(codes.Unimplemented, "method BidiStream not implemented")
}

// Parses the given unary response definition and returns either
// a built payload or a gRPC error based on the definition.
func parseUnaryResponseDefinition(
	def *v1alpha1.UnaryResponseDefinition,
	metadata metadata.MD,
	reqs []*anypb.Any,
) (*v1alpha1.ConformancePayload, error) {
	if def != nil {
		switch respType := def.Response.(type) {
		case *v1alpha1.UnaryResponseDefinition_Error:
			return nil, grpcutil.ConvertProtoToGrpcError(respType.Error)

		case *v1alpha1.UnaryResponseDefinition_ResponseData, nil:
			requestInfo := createRequestInfo(metadata, reqs)
			payload := &v1alpha1.ConformancePayload{
				RequestInfo: requestInfo,
			}

			// If response data was provided, set that in the payload response
			if respType, ok := respType.(*v1alpha1.UnaryResponseDefinition_ResponseData); ok {
				payload.Data = respType.ResponseData
			}
			return payload, nil
		default:
			return nil, status.Errorf(
				codes.InvalidArgument,
				"provided UnaryRequest.Response has an unexpected type %T",
				respType,
			)
		}
	}
	return nil, status.Error(codes.InvalidArgument, "no response definition provided")
}

// Creates request info for a conformance payload.
func createRequestInfo(metadata metadata.MD, reqs []*anypb.Any) *v1alpha1.ConformancePayload_RequestInfo {
	headerInfo := grpcutil.ConvertMetadataToProtoHeader(metadata)

	// Set all observed request headers and requests in the response payload
	return &v1alpha1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
	}
}

// Converts the given message to an Any.
func asAny(msg proto.Message) (*anypb.Any, error) {
	msgAsAny, err := anypb.New(msg)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"unable to convert message: %s",
			err.Error(),
		)
	}
	return msgAsAny, nil
}
