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
	"errors"
	"fmt"
	"io"
	"time"

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
	stream v1alpha1.ConformanceService_ClientStreamServer,
) error {
	var responseDefinition *v1alpha1.UnaryResponseDefinition
	firstRecv := true
	var reqs []*anypb.Any

	for {
		if err := stream.Context().Err(); err != nil {
			return err
		}
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// If this is the first message received on the stream, save off the response definition we need to send
		if firstRecv {
			responseDefinition = msg.ResponseDefinition
			firstRecv = false
		}
		// Record all the requests received
		msgAsAny, err := asAny(msg)
		if err != nil {
			return err
		}
		reqs = append(reqs, msgAsAny)

	}

	grpcutil.AddHeaderMetadata(stream.Context(), responseDefinition.ResponseHeaders)
	grpcutil.AddTrailerMetadata(stream.Context(), responseDefinition.ResponseTrailers)

	md, _ := metadata.FromIncomingContext(stream.Context())
	payload, err := parseUnaryResponseDefinition(responseDefinition, md, reqs)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&v1alpha1.ClientStreamResponse{
		Payload: payload,
	})
}

func (c *conformanceServiceServer) ServerStream(
	req *v1alpha1.ServerStreamRequest,
	stream v1alpha1.ConformanceService_ServerStreamServer,
) error {
	responseDefinition := req.ResponseDefinition
	if responseDefinition != nil {
		grpcutil.AddHeaderMetadata(stream.Context(), responseDefinition.ResponseHeaders)
		grpcutil.AddTrailerMetadata(stream.Context(), responseDefinition.ResponseTrailers)
	}

	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req)
	if err != nil {
		return err
	}

	metadata, _ := metadata.FromIncomingContext(stream.Context())
	requestInfo := createRequestInfo(metadata, []*anypb.Any{msgAsAny})
	payload := &v1alpha1.ConformancePayload{
		RequestInfo: requestInfo,
	}

	for _, data := range responseDefinition.ResponseData {
		payload.Data = data

		resp := &v1alpha1.ServerStreamResponse{
			Payload: payload,
		}

		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
		}
		// Only echo back the request info in the first response
		payload.RequestInfo = nil
	}
	if responseDefinition.Error != nil {
		return grpcutil.ConvertProtoToGrpcError(responseDefinition.Error)
	}
	return nil
}

func (c *conformanceServiceServer) BidiStream(
	stream v1alpha1.ConformanceService_BidiStreamServer,
) error {
	var responseDefinition *v1alpha1.StreamResponseDefinition
	fullDuplex := false
	firstRecv := true
	respNum := 0
	var reqs []*anypb.Any
	for {
		if err := stream.Context().Err(); err != nil {
			return err
		}
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Reads are done, break the receive loop and send any remaining responses
				break
			}
			return fmt.Errorf("receive request: %w", err)
		}

		// Record all requests received
		msgAsAny, err := asAny(req)
		if err != nil {
			return err
		}
		reqs = append(reqs, msgAsAny)

		// If this is the first message in the stream, save off the total responses we need to send
		// plus whether this should be full or half duplex
		if firstRecv {
			responseDefinition = req.ResponseDefinition
			fullDuplex = req.FullDuplex
			firstRecv = false

			if err := grpcutil.AddHeaderMetadata(stream.Context(), req.ResponseDefinition.ResponseHeaders); err != nil {
				return err
			}
			if err := grpcutil.AddTrailerMetadata(stream.Context(), req.ResponseDefinition.ResponseTrailers); err != nil {
				return err
			}
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if respNum >= len(responseDefinition.ResponseData) {
				return status.Error(
					codes.Aborted,
					"received more requests than desired responses on a full duplex stream",
				)
			}
			metadata, _ := metadata.FromIncomingContext(stream.Context())
			requestInfo := createRequestInfo(metadata, reqs)
			resp := &v1alpha1.BidiStreamResponse{
				Payload: &v1alpha1.ConformancePayload{
					RequestInfo: requestInfo,
					Data:        responseDefinition.ResponseData[respNum],
				},
			}
			time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
			}
			respNum++
			reqs = nil
		}
	}

	// If we still have responses left to send, flush them now. This accommodates
	// both scenarios of half duplex (we haven't sent any responses yet) or full duplex
	// where the requested responses are greater than the total requests.
	for ; respNum < len(responseDefinition.ResponseData); respNum++ {
		metadata, _ := metadata.FromIncomingContext(stream.Context())
		requestInfo := createRequestInfo(metadata, reqs)
		resp := &v1alpha1.BidiStreamResponse{
			Payload: &v1alpha1.ConformancePayload{
				RequestInfo: requestInfo,
				Data:        responseDefinition.ResponseData[respNum],
			},
		}
		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
		}
	}

	if responseDefinition.Error != nil {
		return grpcutil.ConvertProtoToGrpcError(responseDefinition.Error)
	}
	return nil
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
			return nil, status.Errorf(codes.InvalidArgument, "provided UnaryRequest.Response has an unexpected type %T", respType)
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
